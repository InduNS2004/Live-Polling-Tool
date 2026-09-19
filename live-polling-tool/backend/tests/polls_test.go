package tests

import (
	"live-polling-tool/backend/models"
	"testing"
)

func TestResultTotal(t *testing.T) {
	r := models.Result{Options: []models.PollOption{{Votes: 2}, {Votes: 3}}}
	var total int64
	for _, o := range r.Options {
		total += o.Votes
	}
	if total != 5 {
		t.Fatalf("expected 5, got %d", total)
	}
}
func TestPollOptionIDs(t *testing.T) {
	p := models.Poll{Options: []models.PollOption{{ID: "a"}, {ID: "b"}}}
	if p.Options[0].ID == p.Options[1].ID {
		t.Fatal("option IDs must be unique")
	}
}
