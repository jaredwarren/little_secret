package server

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
)

type Role string

const (
	RoleGoodKitten     Role = "Good Kitten"
	RoleConfusedKitten Role = "Confused Kitten"
	RoleSpyPup         Role = "Spy Pup"
)

type Stage string

const (
	StageLobby    Stage = "LOBBY"
	StageClues    Stage = "CLUES"    // Online: typing clues
	StageDebate   Stage = "DEBATE"   // Online: review clues
	StageVoting   Stage = "VOTING"   // Online: submit votes
	StageReveal   Stage = "REVEAL"   // Online/In-person: show who is eliminated, handle Spy guess
	StageGameOver Stage = "GAMEOVER" // Both: game summary
	StageDealt    Stage = "DEALT"    // In-person: show cards, play verbally
)

type GameMode string

const (
	ModeOnline   GameMode = "ONLINE"
	ModeInPerson GameMode = "IN_PERSON"
)

type Player struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Role         Role   `json:"role,omitempty"` // Omitted in sanitized states
	Word         string `json:"word,omitempty"` // Omitted in sanitized states
	IsHost       bool   `json:"isHost"`
	IsEliminated bool   `json:"isEliminated"`
	Clue         string `json:"clue"`
	Vote         string `json:"vote"` // ID of player voted for
	Connected    bool   `json:"connected"`
}

type GameConfig struct {
	Mode            GameMode `json:"mode"`
	PackName        string   `json:"packName"`
	ConfusedCount   int      `json:"confusedCount"`
	SpyCount        int      `json:"spyCount"`
	ManualWordNum   int      `json:"manualWordNum"` // 0 for random, 1-21 for manual
	SequentialClues bool     `json:"sequentialClues"`
}

type Lobby struct {
	Code               string             `json:"code"`
	Players            map[string]*Player `json:"players"`
	Stage              Stage              `json:"stage"`
	Config             GameConfig         `json:"config"`
	GoodWord           string             `json:"goodWord,omitempty"`     // Omitted in sanitized states
	ConfusedWord       string             `json:"confusedWord,omitempty"` // Omitted in sanitized states
	ActiveWordNum      int                `json:"activeWordNum"`
	EliminatedThisTurn []string           `json:"eliminatedThisTurn"` // IDs of players eliminated in the current voting round
	Winner             string             `json:"winner"`             // "KITTENS", "SPY_PUP", "CONFUSED_KITTENS"
	TiePlayers         []string           `json:"tiePlayers"`         // Tied player IDs in voting
	IsTieVote          bool               `json:"isTieVote"`
	TurnOrder          []string           `json:"turnOrder"` // Ordered list of active player IDs
	CurrentTurnIdx     int                `json:"currentTurnIdx"`
}

func GenerateRoomCode() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "ABCD"
	}
	return strings.ToUpper(hex.EncodeToString(bytes))
}

func NewLobby(code string, hostID, hostName string) *Lobby {
	return &Lobby{
		Code:    code,
		Players: make(map[string]*Player),
		Stage:   StageLobby,
		Config: GameConfig{
			Mode:          ModeOnline,
			PackName:      DefaultPackName,
			ConfusedCount: -1, // -1 indicates default by player count
			SpyCount:      -1,
			ManualWordNum: 0, // random
		},
	}
}

// GetDefaultRoleCounts returns the standard flyer distribution for a player count.
func GetDefaultRoleCounts(numPlayers int) (confused int, spy int) {
	switch numPlayers {
	case 4:
		return 1, 0
	case 5:
		return 1, 1
	case 6:
		return 1, 1
	case 7:
		return 2, 1
	case 8:
		return 2, 1
	default:
		if numPlayers < 4 {
			return 1, 0
		}
		// For 9+ players, scale Confused Kittens and keep 1 Spy Pup
		return (numPlayers - 1) / 3, 1
	}
}

// StartRound assigns roles, sets the secret words, and transitions the game stage.
func (l *Lobby) StartRound(pack MissionPack) {
	numPlayers := len(l.Players)
	if numPlayers < 4 {
		return // Cannot start with less than 4 players
	}

	// Determine role distribution
	confusedCount := l.Config.ConfusedCount
	spyCount := l.Config.SpyCount
	if confusedCount < 0 || spyCount < 0 {
		confusedCount, spyCount = GetDefaultRoleCounts(numPlayers)
	}

	// Validate role counts do not exceed player counts
	if confusedCount+spyCount >= numPlayers {
		confusedCount, spyCount = GetDefaultRoleCounts(numPlayers)
	}
	goodCount := numPlayers - confusedCount - spyCount

	// Choose word number (1-21)
	wordIdx := 0
	if l.Config.ManualWordNum >= 1 && l.Config.ManualWordNum <= len(pack.Words) {
		wordIdx = l.Config.ManualWordNum - 1
	} else {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pack.Words))))
		wordIdx = int(n.Int64())
	}
	l.ActiveWordNum = wordIdx + 1

	wordPair := pack.Words[wordIdx]

	// Randomly decide which word is Good and which is Confused
	flipResult, _ := rand.Int(rand.Reader, big.NewInt(2))
	if flipResult.Int64() == 0 {
		l.GoodWord = wordPair.Good
		l.ConfusedWord = wordPair.Confused
	} else {
		l.GoodWord = wordPair.Confused
		l.ConfusedWord = wordPair.Good
	}

	// Create a pool of roles
	var rolePool []Role
	for i := 0; i < goodCount; i++ {
		rolePool = append(rolePool, RoleGoodKitten)
	}
	for i := 0; i < confusedCount; i++ {
		rolePool = append(rolePool, RoleConfusedKitten)
	}
	for i := 0; i < spyCount; i++ {
		rolePool = append(rolePool, RoleSpyPup)
	}

	// Shuffle roles
	for i := len(rolePool) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		rolePool[i], rolePool[j] = rolePool[j], rolePool[i]
	}

	// Reset player states and assign roles
	idx := 0
	for _, p := range l.Players {
		p.IsEliminated = false
		p.Clue = ""
		p.Vote = ""
		p.Role = rolePool[idx]
		switch p.Role {
		case RoleGoodKitten:
			p.Word = l.GoodWord
		case RoleConfusedKitten:
			p.Word = l.ConfusedWord
		default:
			p.Word = "" // Spy Pup gets no word
		}
		idx++
	}

	l.EliminatedThisTurn = nil
	l.Winner = ""
	l.TiePlayers = nil
	l.IsTieVote = false

	// Build TurnOrder for sequential clues
	var turnOrder []string
	for id, p := range l.Players {
		if !p.IsEliminated {
			turnOrder = append(turnOrder, id)
		}
	}
	// Shuffle turnOrder to randomize clue order
	for i := len(turnOrder) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		turnOrder[i], turnOrder[j] = turnOrder[j], turnOrder[i]
	}
	l.TurnOrder = turnOrder
	l.CurrentTurnIdx = 0

	if l.Config.Mode == ModeInPerson {
		l.Stage = StageDealt
	} else {
		l.Stage = StageClues
	}
}

// SubmitClue records a player's clue and advances to Debate if all have submitted.
func (l *Lobby) SubmitClue(playerID string, clue string) bool {
	if l.Stage != StageClues {
		return false
	}
	p, ok := l.Players[playerID]
	if !ok || p.IsEliminated {
		return false
	}

	if l.Config.SequentialClues {
		// Verify it is this player's turn
		if len(l.TurnOrder) == 0 || l.TurnOrder[l.CurrentTurnIdx] != playerID {
			return false
		}
		p.Clue = strings.TrimSpace(clue)
		l.CurrentTurnIdx++
		if l.CurrentTurnIdx >= len(l.TurnOrder) {
			l.Stage = StageDebate
			return true
		}
		return true
	}

	p.Clue = strings.TrimSpace(clue)

	// Check if all non-eliminated players have submitted a clue
	allSubmitted := true
	for _, p := range l.Players {
		if !p.IsEliminated && p.Clue == "" {
			allSubmitted = false
			break
		}
	}

	if allSubmitted {
		l.Stage = StageDebate
		return true
	}
	return false
}

// CallVote transitions the lobby from DEBATE to VOTING stage.
func (l *Lobby) CallVote() bool {
	if l.Stage != StageDebate {
		return false
	}
	l.Stage = StageVoting
	// Clear any previous voting data
	for _, p := range l.Players {
		p.Vote = ""
	}
	l.IsTieVote = false
	l.TiePlayers = nil
	return true
}

// SubmitVote records a player's vote and resolves voting if all have voted.
func (l *Lobby) SubmitVote(voterID string, targetID string) bool {
	if l.Stage != StageVoting {
		return false
	}
	voter, ok := l.Players[voterID]
	if !ok || voter.IsEliminated {
		return false
	}
	if targetID != "" {
		target, ok := l.Players[targetID]
		if !ok || target.IsEliminated {
			return false
		}
	}

	voter.Vote = targetID

	// Check if all active players have voted
	allVoted := true
	for _, p := range l.Players {
		if !p.IsEliminated && p.Vote == "" {
			allVoted = false
			break
		}
	}

	if allVoted {
		l.ResolveVoting()
		return true
	}
	return false
}

// ResolveVoting calculates votes and eliminates the top-voted player or flags a tie.
func (l *Lobby) ResolveVoting() {
	voteCounts := make(map[string]int)
	for _, p := range l.Players {
		if !p.IsEliminated && p.Vote != "" {
			voteCounts[p.Vote]++
		}
	}

	maxVotes := -1
	var topPlayers []string

	for id, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
			topPlayers = []string{id}
		} else if count == maxVotes {
			topPlayers = append(topPlayers, id)
		}
	}

	// If no one got any votes (unlikely, but possible if everyone skips)
	if len(topPlayers) == 0 {
		l.IsTieVote = false
		l.TiePlayers = nil
		l.Stage = StageReveal
		return
	}

	if len(topPlayers) > 1 {
		// Tie
		l.IsTieVote = true
		l.TiePlayers = topPlayers
		// Reset votes for tie breaker
		for _, p := range l.Players {
			p.Vote = ""
		}
		// Stay in StageVoting or let clients know they need a tie-breaker.
		// Usually we go to StageVoting again but with only the tied players eligible for defense.
		// For simplicity, we keep stage as VOTING but specify IsTieVote and TiePlayers so UI can adapt.
		return
	}

	// Single top player is eliminated
	eliminatedID := topPlayers[0]
	l.IsTieVote = false
	l.TiePlayers = nil
	l.EliminatePlayer(eliminatedID)
}

// EliminatePlayer marks a player as eliminated, checks win conditions.
func (l *Lobby) EliminatePlayer(playerID string) {
	p, ok := l.Players[playerID]
	if !ok {
		return
	}
	p.IsEliminated = true
	l.EliminatedThisTurn = []string{playerID}
	l.Stage = StageReveal

	l.CheckWinConditions()
}

// RevealPlayerRole is used in In-Person mode when a player is eliminated in real life.
func (l *Lobby) RevealPlayerRole(playerID string) {
	p, ok := l.Players[playerID]
	if !ok || p.IsEliminated {
		return
	}
	p.IsEliminated = true
	l.EliminatedThisTurn = []string{playerID}

	// Keep track of elimination in in-person mode
	l.Stage = StageReveal
	l.CheckWinConditions()
}

// GuessPassword allows the eliminated Spy Pup to guess the Good Kitten word.
func (l *Lobby) GuessPassword(playerID string, guess string) bool {
	p, ok := l.Players[playerID]
	if !ok || p.Role != RoleSpyPup {
		return false
	}

	// Normalize comparison
	cleanGuess := strings.ToLower(strings.TrimSpace(guess))
	cleanGood := strings.ToLower(strings.TrimSpace(l.GoodWord))

	if cleanGuess == cleanGood {
		l.Winner = "SPY_PUP"
		l.Stage = StageGameOver
		return true
	}

	// Wrong guess, game continues (or is checked if game over since Spy failed)
	l.CheckWinConditions()
	return false
}

// CheckWinConditions checks if Kittens or Spy/Confused won.
func (l *Lobby) CheckWinConditions() {
	// If Spy Pup won via correct guess, winner is already set
	if l.Winner == "SPY_PUP" {
		return
	}

	activeGoodKittens := 0
	activeConfusedKittens := 0
	activeSpyPups := 0

	for _, p := range l.Players {
		if !p.IsEliminated {
			switch p.Role {
			case RoleGoodKitten:
				activeGoodKittens++
			case RoleConfusedKitten:
				activeConfusedKittens++
			case RoleSpyPup:
				activeSpyPups++
			}
		}
	}

	// Good Kittens win if all Spy Pups and Confused Kittens are eliminated
	if activeSpyPups == 0 && activeConfusedKittens == 0 {
		l.Winner = "KITTENS"
		l.Stage = StageGameOver
		return
	}

	// Confused Kittens / Spy Pup win if only 2 players remain and not all bad guys are eliminated
	totalActive := activeGoodKittens + activeConfusedKittens + activeSpyPups
	if totalActive <= 2 {
		if activeConfusedKittens > 0 {
			l.Winner = "CONFUSED_KITTENS"
		} else if activeSpyPups > 0 {
			l.Winner = "SPY_PUP"
		}
		l.Stage = StageGameOver
		return
	}
}

// SanitizeState returns a copy of the lobby with hidden fields based on who is viewing it.
// playerID can be empty if viewing as spectator.
func (l *Lobby) SanitizeState(viewerID string) interface{} {
	type SanitizedPlayer struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		IsHost       bool   `json:"isHost"`
		IsEliminated bool   `json:"isEliminated"`
		Clue         string `json:"clue"`
		Vote         string `json:"vote"`
		Connected    bool   `json:"connected"`
		// Revealed at game over or when eliminated
		Role Role   `json:"role,omitempty"`
		Word string `json:"word,omitempty"`
	}

	sanitizedPlayers := make(map[string]*SanitizedPlayer)
	for id, p := range l.Players {
		clueToSend := ""
		if l.Stage != StageClues || id == viewerID || (l.Config.SequentialClues && p.Clue != "") {
			clueToSend = p.Clue
		}

		sp := &SanitizedPlayer{
			ID:           p.ID,
			Name:         p.Name,
			IsHost:       p.IsHost,
			IsEliminated: p.IsEliminated,
			Clue:         clueToSend,
			Vote:         p.Vote,
			Connected:    p.Connected,
		}

		// Reveal role & word if:
		// 1. Game is over
		// 2. Player is eliminated
		// 3. The viewer is this player themselves (they should see their own word/role)
		if l.Stage == StageGameOver || p.IsEliminated || id == viewerID {
			sp.Role = p.Role
			sp.Word = p.Word
		}

		sanitizedPlayers[id] = sp
	}

	type SanitizedLobby struct {
		Code               string                      `json:"code"`
		Players            map[string]*SanitizedPlayer `json:"players"`
		Stage              Stage                       `json:"stage"`
		Config             GameConfig                  `json:"config"`
		ActiveWordNum      int                         `json:"activeWordNum"`
		EliminatedThisTurn []string                    `json:"eliminatedThisTurn"`
		Winner             string                      `json:"winner"`
		TiePlayers         []string                    `json:"tiePlayers"`
		IsTieVote          bool                        `json:"isTieVote"`
		TurnOrder          []string                    `json:"turnOrder"`
		CurrentTurnIdx     int                         `json:"currentTurnIdx"`
		// Only show the real passwords at game over
		GoodWord     string `json:"goodWord,omitempty"`
		ConfusedWord string `json:"confusedWord,omitempty"`
	}

	sl := SanitizedLobby{
		Code:               l.Code,
		Players:            sanitizedPlayers,
		Stage:              l.Stage,
		Config:             l.Config,
		ActiveWordNum:      l.ActiveWordNum,
		EliminatedThisTurn: l.EliminatedThisTurn,
		Winner:             l.Winner,
		TiePlayers:         l.TiePlayers,
		IsTieVote:          l.IsTieVote,
		TurnOrder:          l.TurnOrder,
		CurrentTurnIdx:     l.CurrentTurnIdx,
	}

	if l.Stage == StageGameOver {
		sl.GoodWord = l.GoodWord
		sl.ConfusedWord = l.ConfusedWord
	}

	return sl
}
