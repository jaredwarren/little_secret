# Little Secret 🐱🐶

A sleek, cyberpunk-themed real-time multiplayer social deduction party game built with Go, WebSockets, and vanilla Web technologies (HTML, CSS, and JS). 

Inspired by party games like *Undercover* and *Little Secret*, this digital version tasks players with identifying hidden identities through word association clues and debate, set against an aesthetic cyberpunk backdrop.

---

## 🎮 Game Overview & Rules

Little Secret is best played with **4 to 10+ players**. Players are assigned one of three secret roles, and must figure out who is who by submitting single-word clues.

### Roles
1. **Good Kitten (Majority)** 🐱
   - **Secret Card:** Receives the **Good Word** (e.g., "Tea").
   - **Objective:** Detect and eliminate all **Spy Pups** and **Confused Kittens** before the game ends.
2. **Confused Kitten (1-2 players)** 😿
   - **Secret Card:** Receives a related but different **Confused Word** (e.g., "Coffee").
   - **Objective:** Survive! They start out thinking they are a Good Kitten, but must realize they are "Confused" by paying attention to clues and adapt their strategy to survive alongside the Spy Pup.
3. **Spy Pup (1-2 players)** 🐶
   - **Secret Card:** Receives **no word** (shown as a `?`).
   - **Objective:** Blend in with the kittens by providing believable clues. If voted out, the Spy Pup gets one final chance to win by guessing the Good Kitten's word.

### Game Stages
1. **Lobby (`LOBBY`)**: Players join the room using a 6-character room code. The host selects the pack, game mode, and starts the round.
2. **Clues (`CLUES`)**: Every player types a one-word clue that relates to their secret word. (Supports *Simultaneous* or *Sequential* turn-by-turn typing styles).
3. **Debate (`DEBATE`)**: All clues are revealed. Players verbally debate who is the Spy Pup or Confused Kitten.
4. **Voting (`VOTING`)**: Players vote on who to eliminate. If there is a tie, a tiebreaker vote is held.
5. **Reveal (`REVEAL`)**: The eliminated player is revealed. If it is a Spy Pup, they can enter their final password guess.
6. **Game Over (`GAMEOVER`)**: Shows the final standings, all players' roles, and the secret words.

---

## ⚙️ Gameplay Modes

- **Online Mode (Default)**: Full digital execution. Clue submission, debate presentation, and secret voting are handled entirely inside the browser.
- **In-Person Mode**: Face-to-face setup. The browser acts purely as a card dealer to distribute secret roles/words. Clues, debates, and votes are performed verbally in real-life, and players click "Reveal Role" on their device only when voted out.

---

## 🛠️ Tech Stack & Architecture

- **Backend**: Go (Golang) standard library HTTP multiplexer combined with `github.com/gorilla/websocket` for real-time state synchronization.
- **Frontend**: Single Page Application (SPA) utilizing vanilla HTML5, CSS3 (with custom cyberpunk animations and layout), and modern JavaScript (WebSocket API, reactive state rendering).
- **Communication Protocol**: Full-duplex JSON payloads via WebSockets. Every client action updates the room state on the server, which is then sanitized (hiding secret roles/words of other players) and broadcasted back to all clients in real-time.

---

## 🚀 Getting Started

### Prerequisites
- [Go](https://go.dev/doc/install) (version 1.23 or newer)
- [Make](https://www.gnu.org/software/make/) (optional, but recommended)
- [Docker](https://www.docker.com/) (optional, for containerized execution)

### Option 1: Local Development (With Make)
The included [Makefile](Makefile) contains easy shortcuts to manage the project:

```bash
# Print all available Makefile targets
make help

# Run the Go server locally on default port 8080
make run

# Run the Go server on a custom port
PORT=9090 make run

# Run unit tests
make test

# Format the codebase
make fmt

# Vet/lint the codebase
make lint

# Compile production binary to bin/littlesecrets
make build
```

### Option 2: Docker & Docker Compose
To build and run the application in a lightweight container:

```bash
# Start the server using Docker Compose
docker compose up --build
```
The application will be accessible at `http://localhost:8080`.

---

## 📦 Custom Word Packs

Little Secret supports uploading custom card packs. The default game comes with a **Classic Pack**, but you can add your own via the **Pack Manager** modal in the top header.

Custom packs must be in JSON format and contain exactly **21 word pairs**:

```json
{
  "name": "Pop Culture",
  "words": [
    {"good": "Batman", "confused": "Ironman"},
    {"good": "Harry Potter", "confused": "Percy Jackson"},
    {"good": "Star Wars", "confused": "Star Trek"},
    ...
  ]
}
```

Packs are saved as JSON files in the `data/packs/` directory and loaded dynamically by the server.

---

## 📂 Project Structure

```
├── .github/workflows/   # CI/CD Workflows
│   └── go.yml           # GitHub Actions configuration
├── cmd/
│   └── game/            # Go entrypoint directory
│       └── main.go      # Main runner file
├── data/
│   └── packs/           # JSON files containing word/card packs
├── internal/
│   └── server/          # Core Go game backend
│       ├── hub.go       # WebSocket Connection and Lobby Hub
│       ├── lobby.go     # Lobby State and game rules/engine
│       ├── lobby_test.go # Unit tests for game rules and states
│       └── pack.go      # Pack loading/saving utilities
├── static/              # Frontend Assets
│   ├── index.html       # Single Page Application HTML structure
│   ├── index.css        # Cyberpunk stylesheet & grid layout
│   ├── app.js           # Client websocket controller and UI renderer
│   └── *.png            # Cyberpunk assets
├── Dockerfile           # Multi-stage production container setup
├── docker-compose.yml   # Docker compose configuration
├── go.mod               # Go module description
├── Makefile             # Development helper tasks
└── README.md            # You are here!
```
