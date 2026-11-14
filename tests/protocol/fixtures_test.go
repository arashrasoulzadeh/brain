package protocol_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoginFixtureIsTransportNeutral(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "protocol", "fixtures", "login.json")
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Protocol  int    `json:"protocol"`
		Fixture   string `json:"fixture"`
		ProjectID string `json:"projectId"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		Expected  struct {
			Authenticated bool   `json:"authenticated"`
			SessionCookie string `json:"sessionCookie"`
		} `json:"expected"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Protocol != 1 || fixture.Fixture != "login" || fixture.ProjectID == "" ||
		fixture.Username == "" || fixture.Password == "" || !fixture.Expected.Authenticated ||
		fixture.Expected.SessionCookie != "__Host-brain_session" {
		t.Fatalf("invalid login fixture: %+v", fixture)
	}
}

func TestRoomAdmissionFixtureIsTransportNeutral(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "protocol", "fixtures", "room-admission.json")
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Protocol         int    `json:"protocol"`
		Fixture          string `json:"fixture"`
		ProjectID        string `json:"projectId"`
		RoomID           string `json:"roomId"`
		PlayerID         string `json:"playerId"`
		ConnectionTicket string `json:"connectionTicket"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Protocol != 1 || fixture.Fixture != "room-admission" || fixture.ProjectID == "" ||
		fixture.RoomID == "" || fixture.PlayerID == "" || fixture.ConnectionTicket == "" {
		t.Fatalf("invalid room admission fixture: %+v", fixture)
	}
}

func TestAssignmentFixtureIsTransportNeutral(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "protocol", "fixtures", "assignment.json")
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Protocol  int      `json:"protocol"`
		Fixture   string   `json:"fixture"`
		ProjectID string   `json:"projectId"`
		MatchID   string   `json:"matchId"`
		RoomID    string   `json:"roomId"`
		Players   []string `json:"players"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Protocol != 1 || fixture.Fixture != "assignment" || fixture.ProjectID == "" ||
		fixture.MatchID == "" || fixture.RoomID == "" || len(fixture.Players) != 2 {
		t.Fatalf("invalid assignment fixture: %+v", fixture)
	}
}
