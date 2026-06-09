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
