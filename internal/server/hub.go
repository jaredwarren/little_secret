package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

type ClientMessage struct {
	Action     string      `json:"action"`
	PlayerID   string      `json:"playerId"`
	RoomCode   string      `json:"roomCode"`
	PlayerName string      `json:"playerName"`
	Config     *GameConfig `json:"config,omitempty"`
	Clue       string      `json:"clue,omitempty"`
	TargetID   string      `json:"targetId,omitempty"`
	Guess      string      `json:"guess,omitempty"`
}

type ServerMessage struct {
	Type    string      `json:"type"` // "state", "error", "welcome"
	Lobby   interface{} `json:"lobby,omitempty"`
	Error   string      `json:"error,omitempty"`
	Welcome struct {
		PlayerID string `json:"playerId"`
		RoomCode string `json:"roomCode"`
	} `json:"welcome,omitempty"`
}

type Hub struct {
	mu       sync.RWMutex
	lobbies  map[string]*Lobby
	conns    map[string]map[string]*websocket.Conn // roomCode -> playerID -> Conn
	packsDir string
}

func NewHub(packsDir string) *Hub {
	return &Hub{
		lobbies:  make(map[string]*Lobby),
		conns:    make(map[string]map[string]*websocket.Conn),
		packsDir: packsDir,
	}
}

func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	var currentPlayerID string
	var currentRoomCode string

	defer func() {
		if currentRoomCode != "" && currentPlayerID != "" {
			h.mu.Lock()
			// Mark player disconnected
			if lobby, ok := h.lobbies[currentRoomCode]; ok {
				if player, ok := lobby.Players[currentPlayerID]; ok {
					player.Connected = false
					log.Printf("Player %s (%s) disconnected from room %s", player.Name, currentPlayerID, currentRoomCode)
				}
				// Clean up connection
				if roomConns, ok := h.conns[currentRoomCode]; ok {
					delete(roomConns, currentPlayerID)
				}
			}
			h.mu.Unlock()
			h.BroadcastLobbyState(currentRoomCode)
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			h.sendError(conn, "Invalid message format")
			continue
		}

		switch msg.Action {
		case "create_room":
			h.handleCreateRoom(conn, &msg, &currentPlayerID, &currentRoomCode)
		case "join_room":
			h.handleJoinRoom(conn, &msg, &currentPlayerID, &currentRoomCode)
		case "configure_room":
			h.handleConfigureRoom(conn, &msg)
		case "start_game":
			h.handleStartGame(conn, &msg)
		case "submit_clue":
			h.handleSubmitClue(conn, &msg)
		case "submit_vote":
			h.handleSubmitVote(conn, &msg)
		case "call_vote":
			h.handleCallVote(conn, &msg)
		case "reveal_role":
			h.handleRevealRole(conn, &msg)
		case "guess_password":
			h.handleGuessPassword(conn, &msg)
		case "next_round":
			h.handleNextRound(conn, &msg)
		case "restart_lobby":
			h.handleRestartLobby(conn, &msg)
		}
	}
}

func (h *Hub) handleCreateRoom(conn *websocket.Conn, msg *ClientMessage, currentPlayerID *string, currentRoomCode *string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomCode := GenerateRoomCode()
	// Ensure room code uniqueness
	for _, ok := h.lobbies[roomCode]; ok; _, ok = h.lobbies[roomCode] {
		roomCode = GenerateRoomCode()
	}

	playerID := msg.PlayerID
	if playerID == "" {
		playerID = GenerateRoomCode() // reusable helper for short unique ID
	}

	lobby := NewLobby(roomCode, playerID, msg.PlayerName)
	lobby.Players[playerID] = &Player{
		ID:        playerID,
		Name:      msg.PlayerName,
		IsHost:    true,
		Connected: true,
	}

	h.lobbies[roomCode] = lobby
	h.conns[roomCode] = map[string]*websocket.Conn{
		playerID: conn,
	}

	*currentPlayerID = playerID
	*currentRoomCode = roomCode

	log.Printf("Room %s created by %s (%s)", roomCode, msg.PlayerName, playerID)

	// Send welcome message
	welcomeMsg := ServerMessage{Type: "welcome"}
	welcomeMsg.Welcome.PlayerID = playerID
	welcomeMsg.Welcome.RoomCode = roomCode
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to write welcome message: %v", err)
	}

	// Broadcast lobby state
	go h.BroadcastLobbyState(roomCode)
}

func (h *Hub) handleJoinRoom(conn *websocket.Conn, msg *ClientMessage, currentPlayerID *string, currentRoomCode *string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomCode := strings.ToUpper(strings.TrimSpace(msg.RoomCode))
	lobby, exists := h.lobbies[roomCode]
	if !exists {
		h.sendError(conn, "Room not found")
		return
	}

	playerID := msg.PlayerID
	if playerID == "" {
		playerID = GenerateRoomCode()
	}

	// Check if player is reconnecting
	player, ok := lobby.Players[playerID]
	if ok {
		// Reconnecting
		player.Connected = true
		player.Name = msg.PlayerName // Update name if changed
	} else {
		// New player joining
		if lobby.Stage != StageLobby {
			h.sendError(conn, "Game already in progress")
			return
		}
		player = &Player{
			ID:        playerID,
			Name:      msg.PlayerName,
			IsHost:    false,
			Connected: true,
		}
		lobby.Players[playerID] = player
	}

	if h.conns[roomCode] == nil {
		h.conns[roomCode] = make(map[string]*websocket.Conn)
	}
	h.conns[roomCode][playerID] = conn

	*currentPlayerID = playerID
	*currentRoomCode = roomCode

	log.Printf("Player %s (%s) joined room %s", msg.PlayerName, playerID, roomCode)

	// Send welcome
	welcomeMsg := ServerMessage{Type: "welcome"}
	welcomeMsg.Welcome.PlayerID = playerID
	welcomeMsg.Welcome.RoomCode = roomCode
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to write welcome message: %v", err)
	}

	// Broadcast
	go h.BroadcastLobbyState(roomCode)
}

func (h *Hub) handleConfigureRoom(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	lobby, ok := h.lobbies[msg.RoomCode]
	if !ok {
		h.sendError(conn, "Room not found")
		return
	}

	p, ok := lobby.Players[msg.PlayerID]
	if !ok || !p.IsHost {
		h.sendError(conn, "Only the host can configure settings")
		return
	}

	if msg.Config != nil {
		lobby.Config = *msg.Config
	}

	log.Printf("Room %s settings updated", msg.RoomCode)
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleStartGame(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	lobby, ok := h.lobbies[msg.RoomCode]
	h.mu.Unlock()

	if !ok {
		h.sendError(conn, "Room not found")
		return
	}

	p, ok := lobby.Players[msg.PlayerID]
	if !ok || !p.IsHost {
		h.sendError(conn, "Only host can start game")
		return
	}

	// Load selected pack
	packs, err := LoadPacks(h.packsDir)
	if err != nil {
		h.sendError(conn, "Failed to load card packs")
		return
	}

	packName := lobby.Config.PackName
	pack, exists := packs[packName]
	if !exists {
		// Fallback to Classic Pack
		pack, exists = packs[DefaultPackName]
		if !exists {
			h.sendError(conn, "Card pack not found")
			return
		}
	}

	h.mu.Lock()
	lobby.StartRound(pack)
	h.mu.Unlock()

	log.Printf("Room %s game started with pack %s", msg.RoomCode, pack.Name)
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleSubmitClue(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	lobby, ok := h.lobbies[msg.RoomCode]
	if !ok {
		h.mu.Unlock()
		h.sendError(conn, "Room not found")
		return
	}

	advanced := lobby.SubmitClue(msg.PlayerID, msg.Clue)
	h.mu.Unlock()

	if advanced {
		log.Printf("Room %s advanced to DEBATE stage", msg.RoomCode)
	}
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleSubmitVote(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	lobby, ok := h.lobbies[msg.RoomCode]
	if !ok {
		h.mu.Unlock()
		h.sendError(conn, "Room not found")
		return
	}

	advanced := lobby.SubmitVote(msg.PlayerID, msg.TargetID)
	h.mu.Unlock()

	if advanced {
		log.Printf("Room %s resolved voting", msg.RoomCode)
	}
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleCallVote(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	lobby, ok := h.lobbies[msg.RoomCode]
	if !ok {
		h.mu.Unlock()
		h.sendError(conn, "Room not found")
		return
	}

	p, ok := lobby.Players[msg.PlayerID]
	if !ok || !p.IsHost {
		h.mu.Unlock()
		h.sendError(conn, "Only the host can call the vote")
		return
	}

	advanced := lobby.CallVote()
	h.mu.Unlock()

	if advanced {
		log.Printf("Room %s: Vote called by host", msg.RoomCode)
	}
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleRevealRole(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	lobby, ok := h.lobbies[msg.RoomCode]
	if !ok {
		h.mu.Unlock()
		h.sendError(conn, "Room not found")
		return
	}

	// In-person mode player reveal
	lobby.RevealPlayerRole(msg.PlayerID)
	h.mu.Unlock()

	log.Printf("Room %s: Player %s revealed role", msg.RoomCode, msg.PlayerID)
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleGuessPassword(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	lobby, ok := h.lobbies[msg.RoomCode]
	if !ok {
		h.mu.Unlock()
		h.sendError(conn, "Room not found")
		return
	}

	correct := lobby.GuessPassword(msg.PlayerID, msg.Guess)
	h.mu.Unlock()

	log.Printf("Room %s: Spy guess correct=%v", msg.RoomCode, correct)
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleNextRound(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	lobby, ok := h.lobbies[msg.RoomCode]
	h.mu.Unlock()

	if !ok {
		h.sendError(conn, "Room not found")
		return
	}

	p, ok := lobby.Players[msg.PlayerID]
	if !ok || !p.IsHost {
		h.sendError(conn, "Only host can advance round")
		return
	}

	packs, err := LoadPacks(h.packsDir)
	if err != nil {
		h.sendError(conn, "Failed to load card packs")
		return
	}

	pack, exists := packs[lobby.Config.PackName]
	if !exists {
		pack = packs[DefaultPackName]
	}

	h.mu.Lock()
	// Increment manual word num if it is a manual sequence to help them go to next card
	if lobby.Config.ManualWordNum > 0 {
		lobby.Config.ManualWordNum = (lobby.Config.ManualWordNum % len(pack.Words)) + 1
	}
	lobby.StartRound(pack)
	h.mu.Unlock()

	log.Printf("Room %s started next round", msg.RoomCode)
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) handleRestartLobby(conn *websocket.Conn, msg *ClientMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	lobby, ok := h.lobbies[msg.RoomCode]
	if !ok {
		h.sendError(conn, "Room not found")
		return
	}

	p, ok := lobby.Players[msg.PlayerID]
	if !ok || !p.IsHost {
		h.sendError(conn, "Only host can reset to lobby")
		return
	}

	lobby.Stage = StageLobby
	lobby.Winner = ""
	lobby.EliminatedThisTurn = nil
	for _, player := range lobby.Players {
		player.Clue = ""
		player.Vote = ""
		player.IsEliminated = false
		player.Role = ""
		player.Word = ""
	}

	log.Printf("Room %s reset to lobby stage", msg.RoomCode)
	go h.BroadcastLobbyState(msg.RoomCode)
}

func (h *Hub) BroadcastLobbyState(roomCode string) {
	h.mu.RLock()
	lobby, lobbyExists := h.lobbies[roomCode]
	roomConns, connsExists := h.conns[roomCode]
	h.mu.RUnlock()

	if !lobbyExists || !connsExists {
		return
	}

	// Send state to each client, sanitized specifically for them
	for playerID, conn := range roomConns {
		sanitizedState := lobby.SanitizeState(playerID)
		msg := ServerMessage{
			Type:  "state",
			Lobby: sanitizedState,
		}

		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Write error to player %s in room %s: %v", playerID, roomCode, err)
		}
	}
}

func (h *Hub) sendError(conn *websocket.Conn, message string) {
	msg := ServerMessage{
		Type:  "error",
		Error: message,
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to write error message: %v", err)
	}
}
