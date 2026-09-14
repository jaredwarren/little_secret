# Little Secret Online - Feature Roadmap

This document outlines a potential feature roadmap for **Little Secret Online**, drawing inspiration from the original flyer rules, physical game components, and standard enhancements for web-based party games.

---

## 1. Physical Game Rules & Mechanics (High Priority)

### 🛡️ Secret Weapon Cards (6+ Players)
The flyer mentions an optional "Secret Weapon" card dealt to each player at the start of the game. If implemented digitally:
*   **Ability Dealing**: When the game starts, each player is dealt a secret, single-use action card.
*   **Activation Triggers**: Players can click to activate their card either during the *Clues Stage*, *Debate Stage*, or just before *Voting*.
*   **Abilities Draft**:
    *   **Bulletproof**: Protect a chosen player (or yourself) from elimination for the current round if they receive the most votes.
    *   **Double Agent**: Your vote counts as two votes.
    *   **Spyglass**: Inspect a player's card (shows if they are a Kitten or Spy, but not their specific word).
    *   **Jammer**: Prevent a chosen player from typing a clue or chatting for one round.
    *   **Decoy**: Redirect all votes cast against you to another player.

### ⏱️ Verbal Tie-Breaker Defense Timer
The flyer states: *"In case of a tie, you have about ten seconds to defend yourself."*
*   **Digital Defense Stage**: When a voting tie occurs, transition the lobby to a brief `DEFENSE` stage.
*   **On-Screen Countdown**: Displays a 10-second visual count-down timer for each tied player to defend themselves verbally before the tie-breaker vote is opened.

### 🗳️ Volunteer Selector for Password Number
Currently, the host selects the card number (1-21) or lets the server choose randomly.
*   **Volunteer Turn**: Let the server select a random "volunteer" player in the lobby at the start of the round to pick a card number (1-21) on their device, matching the physical game flow.

---

## 2. Content & Customization (Medium Priority)

### 🎨 Custom Pack Creator UI
Instead of requiring players to upload raw JSON files, build a visual pack editor:
*   **Form Grid**: A spreadsheet-like table with 21 rows.
*   **Fields**: Word A (Good) and Word B (Confused) columns, plus a Theme/Pack Name input.
*   **Validation**: Real-time checking to ensure exactly 21 pairs are entered.
*   **Export/Save**: Save directly to the server's data directory.

### 🦊 Cyberpunk Avatars Selector
Expand the visual theme with selectable profile pictures:
*   Use the generated cyberpunk kitten and pup assets as default avatars.
*   Let players select from a grid of cute neon hacker cats and spy dogs in the lobby.

---

## 3. Remote Gameplay Enhancements (Low Priority)

### 🎙️ WebRTC Voice Chat Integration
For groups playing entirely online on separate networks:
*   **Voice Lobby**: Integrate a peer-to-peer audio room directly into the lobby.
*   **Mute Controls**: Automatically mute players who have been eliminated.

### 💬 In-App Text Chat
*   Add a side-drawer text chat panel for the lobby and during the debate phase to allow typing discussions if players are not on an external voice call.

---

## 4. Quality of Life & Polish

### 🎵 Immersive Synthwave Audio & Sound Effects
*   **Ambient Music**: Subtle synthwave loop playing in the background (with a mute toggle).
*   **Sound Effects**: Play clean retro-futuristic sound cues when:
    *   A player joins the lobby.
    *   Your turn starts in sequential clues mode.
    *   The card is flipped.
    *   A player is eliminated.
    *   A winner is declared.

### 📊 Persistent Player Stats & Leaderboard
Using a lightweight local SQLite database:
*   Track win/loss ratios.
*   Record achievements (e.g., "Perfect Bluff" - win as Spy Pup without getting any votes).
