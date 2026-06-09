package server

import (
	"testing"
)

func TestGetDefaultRoleCounts(t *testing.T) {
	tests := []struct {
		players  int
		expectedConfused int
		expectedSpy      int
	}{
		{4, 1, 0},
		{5, 1, 1},
		{6, 1, 1},
		{7, 2, 1},
		{8, 2, 1},
		{9, 2, 1},
		{10, 3, 1},
	}

	for _, tt := range tests {
		c, s := GetDefaultRoleCounts(tt.players)
		if c != tt.expectedConfused || s != tt.expectedSpy {
			t.Errorf("For %d players, expected confused=%d, spy=%d; got confused=%d, spy=%d",
				tt.players, tt.expectedConfused, tt.expectedSpy, c, s)
		}
	}
}

func TestStartRoundAndGameplayOnline(t *testing.T) {
	pack := MissionPack{
		Name: "Test Pack",
		Words: []WordPair{
			{"GoodWord", "ConfusedWord"},
		},
	}

	lobby := NewLobby("TEST", "host1", "Alice")
	lobby.Players["host1"] = &Player{ID: "host1", Name: "Alice", IsHost: true, Connected: true}
	lobby.Players["p2"] = &Player{ID: "p2", Name: "Bob", Connected: true}
	lobby.Players["p3"] = &Player{ID: "p3", Name: "Charlie", Connected: true}
	lobby.Players["p4"] = &Player{ID: "p4", Name: "David", Connected: true}

	// Start round
	lobby.StartRound(pack)

	if lobby.Stage != StageClues {
		t.Errorf("Expected stage to be CLUES, got %s", lobby.Stage)
	}

	// Verify roles and words
	goodKittens := 0
	confusedKittens := 0
	spyPups := 0

	for _, p := range lobby.Players {
		switch p.Role {
		case RoleGoodKitten:
			goodKittens++
			if p.Word != "GoodWord" {
				t.Errorf("Good Kitten got wrong word: %s", p.Word)
			}
		case RoleConfusedKitten:
			confusedKittens++
			if p.Word != "ConfusedWord" {
				t.Errorf("Confused Kitten got wrong word: %s", p.Word)
			}
		case RoleSpyPup:
			spyPups++
			if p.Word != "" {
				t.Errorf("Spy Pup got a word: %s", p.Word)
			}
		}
	}

	if goodKittens != 3 || confusedKittens != 1 || spyPups != 0 {
		t.Errorf("Expected 3 Good, 1 Confused, 0 Spy; got %d, %d, %d", goodKittens, confusedKittens, spyPups)
	}

	// Force deterministic role setup to avoid test flakes from random assignment
	lobby.Players["host1"].Role = RoleGoodKitten
	lobby.Players["host1"].Word = "GoodWord"
	lobby.Players["p2"].Role = RoleGoodKitten
	lobby.Players["p2"].Word = "GoodWord"
	lobby.Players["p3"].Role = RoleGoodKitten
	lobby.Players["p3"].Word = "GoodWord"
	lobby.Players["p4"].Role = RoleConfusedKitten
	lobby.Players["p4"].Word = "ConfusedWord"

	// Submit clues
	lobby.SubmitClue("host1", "clue1")
	lobby.SubmitClue("p2", "clue2")
	lobby.SubmitClue("p3", "clue3")
	
	// Stage shouldn't change yet
	if lobby.Stage != StageClues {
		t.Errorf("Stage changed too early: %s", lobby.Stage)
	}

	lobby.SubmitClue("p4", "clue4")

	// Now stage should be DEBATE
	if lobby.Stage != StageDebate {
		t.Errorf("Expected stage DEBATE, got %s", lobby.Stage)
	}

	// Move to voting (simulated by helper)
	lobby.Stage = StageVoting

	// Submit votes
	// Everyone votes for Bob (p2)
	lobby.SubmitVote("host1", "p2")
	lobby.SubmitVote("p2", "p3")
	lobby.SubmitVote("p3", "p2")
	lobby.SubmitVote("p4", "p2")

	// Bob should be eliminated
	if !lobby.Players["p2"].IsEliminated {
		t.Errorf("Bob was not eliminated")
	}

	if lobby.Stage != StageReveal {
		t.Errorf("Expected stage REVEAL, got %s", lobby.Stage)
	}
}

func TestResolveVotingTie(t *testing.T) {
	lobby := NewLobby("TEST", "host1", "Alice")
	lobby.Players["host1"] = &Player{ID: "host1", Name: "Alice", IsHost: true, Connected: true}
	lobby.Players["p2"] = &Player{ID: "p2", Name: "Bob", Connected: true}
	lobby.Players["p3"] = &Player{ID: "p3", Name: "Charlie", Connected: true}
	lobby.Players["p4"] = &Player{ID: "p4", Name: "David", Connected: true}
	lobby.Stage = StageVoting

	// Vote tie: Alice and Bob get 2 votes each
	lobby.SubmitVote("host1", "p2")
	lobby.SubmitVote("p2", "host1")
	lobby.SubmitVote("p3", "p2")
	lobby.SubmitVote("p4", "host1")

	if !lobby.IsTieVote {
		t.Errorf("Expected a tie vote")
	}

	if len(lobby.TiePlayers) != 2 {
		t.Errorf("Expected 2 tied players, got %v", lobby.TiePlayers)
	}

	// Resolve tie by voting again
	lobby.SubmitVote("host1", "p2")
	lobby.SubmitVote("p2", "host1")
	lobby.SubmitVote("p3", "p2")
	lobby.SubmitVote("p4", "p2") // p2 gets 3 votes, host1 gets 1

	if lobby.Players["p2"].IsEliminated == false {
		t.Errorf("Expected p2 to be eliminated after tie-breaker")
	}
}

func TestWinConditions(t *testing.T) {
	type testPlayer struct {
		id           string
		role         Role
		isEliminated bool
	}

	tests := []struct {
		name           string
		initialStage   Stage
		goodWord       string
		players        []testPlayer
		action         func(l *Lobby)
		expectedWinner string
		expectedStage  Stage
	}{
		{
			name:         "Kittens win when all bad guys eliminated",
			initialStage: StageReveal,
			players: []testPlayer{
				{"p1", RoleGoodKitten, false},
				{"p2", RoleGoodKitten, false},
				{"p3", RoleConfusedKitten, true}, // Confused Kitten is eliminated
				{"p4", RoleSpyPup, true},          // Spy Pup is eliminated
			},
			action: func(l *Lobby) {
				l.CheckWinConditions()
			},
			expectedWinner: "KITTENS",
			expectedStage:  StageGameOver,
		},
		{
			name:         "Spy Pup wins when only 2 active players remain and Spy is active",
			initialStage: StageReveal,
			players: []testPlayer{
				{"p1", RoleGoodKitten, false},
				{"p2", RoleGoodKitten, true},  // Good Kitten eliminated
				{"p3", RoleGoodKitten, true},  // Good Kitten eliminated
				{"p4", RoleSpyPup, false},      // Spy Pup is still active
			},
			action: func(l *Lobby) {
				l.CheckWinConditions()
			},
			expectedWinner: "SPY_PUP",
			expectedStage:  StageGameOver,
		},
		{
			name:         "Confused Kittens win when only 2 active players remain and Confused is active",
			initialStage: StageReveal,
			players: []testPlayer{
				{"p1", RoleGoodKitten, false},
				{"p2", RoleGoodKitten, true},
				{"p3", RoleConfusedKitten, false}, // Confused Kitten still active
				{"p4", RoleSpyPup, true},
			},
			action: func(l *Lobby) {
				l.CheckWinConditions()
			},
			expectedWinner: "CONFUSED_KITTENS",
			expectedStage:  StageGameOver,
		},
		{
			name:         "Spy Pup wins instantly by correct password guess",
			initialStage: StageReveal,
			goodWord:     "Coffee",
			players: []testPlayer{
				{"p1", RoleGoodKitten, false},
				{"p2", RoleGoodKitten, false},
				{"p3", RoleConfusedKitten, false},
				{"p4", RoleSpyPup, true}, // Spy Pup is eliminated and guessing
			},
			action: func(l *Lobby) {
				l.GuessPassword("p4", " Coffee ") // check whitespace cleaning too
			},
			expectedWinner: "SPY_PUP",
			expectedStage:  StageGameOver,
		},
		{
			name:         "Game continues when Spy Pup guesses wrong and total active > 2",
			initialStage: StageReveal,
			goodWord:     "Coffee",
			players: []testPlayer{
				{"p1", RoleGoodKitten, false},
				{"p2", RoleGoodKitten, false},
				{"p3", RoleConfusedKitten, false}, // Confused kitten is active
				{"p4", RoleSpyPup, true},          // Spy Pup is eliminated
			},
			action: func(l *Lobby) {
				l.GuessPassword("p4", "wrong-word")
			},
			expectedWinner: "",
			expectedStage:  StageReveal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lobby := NewLobby("TEST", "p1", "Alice")
			lobby.Stage = tt.initialStage
			lobby.GoodWord = tt.goodWord

			for _, p := range tt.players {
				lobby.Players[p.id] = &Player{
					ID:           p.id,
					Role:         p.role,
					IsEliminated: p.isEliminated,
				}
			}

			if tt.action != nil {
				tt.action(lobby)
			}

			if lobby.Winner != tt.expectedWinner {
				t.Errorf("Expected winner %q, got %q", tt.expectedWinner, lobby.Winner)
			}

			if lobby.Stage != tt.expectedStage {
				t.Errorf("Expected stage %s, got %s", tt.expectedStage, lobby.Stage)
			}
		})
	}
}

func TestSequentialClues(t *testing.T) {
	pack := MissionPack{
		Name: "Test Pack",
		Words: []WordPair{
			{"GoodWord", "ConfusedWord"},
		},
	}

	lobby := NewLobby("TEST", "p1", "Alice")
	lobby.Config.SequentialClues = true
	lobby.Players["p1"] = &Player{ID: "p1", Name: "Alice", IsHost: true, Connected: true}
	lobby.Players["p2"] = &Player{ID: "p2", Name: "Bob", Connected: true}
	lobby.Players["p3"] = &Player{ID: "p3", Name: "Charlie", Connected: true}
	lobby.Players["p4"] = &Player{ID: "p4", Name: "David", Connected: true}

	lobby.StartRound(pack)

	// Since StartRound shuffles TurnOrder, we will override TurnOrder to be deterministic
	lobby.TurnOrder = []string{"p1", "p2", "p3", "p4"}
	lobby.CurrentTurnIdx = 0

	// Try submitting for p2 (who is NOT the active player) -> should fail
	if ok := lobby.SubmitClue("p2", "clue-bob"); ok {
		t.Errorf("Expected SubmitClue to fail for p2 because it's p1's turn")
	}

	// Submit for p1 (active) -> should succeed and advance index
	if ok := lobby.SubmitClue("p1", "clue-alice"); !ok {
		t.Errorf("Expected SubmitClue to succeed for p1")
	}
	if lobby.CurrentTurnIdx != 1 {
		t.Errorf("Expected CurrentTurnIdx to be 1, got %d", lobby.CurrentTurnIdx)
	}

	// Submit for remaining players
	lobby.SubmitClue("p2", "clue-bob")
	lobby.SubmitClue("p3", "clue-charlie")
	
	// Stage shouldn't be DEBATE yet
	if lobby.Stage != StageClues {
		t.Errorf("Expected stage CLUES before last player submits, got %s", lobby.Stage)
	}

	// Submit for last player p4 -> should advance to DEBATE stage
	lobby.SubmitClue("p4", "clue-david")
	if lobby.Stage != StageDebate {
		t.Errorf("Expected stage DEBATE after last player submits, got %s", lobby.Stage)
	}
}
