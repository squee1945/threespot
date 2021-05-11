package game

import (
	"testing"

	"github.com/squee1945/threespot/server/pkg/storage"
)

func TestNewRules(t *testing.T) {
	rules := NewRules()

	if got, want := rules.PassCard(), false; got != want {
		t.Errorf("PassCard()=%t want=%t", got, want)
	}
}

func TestSetPassCard(t *testing.T) {
	rules := NewRules()
	rules.SetPassCard(true)
	if got, want := rules.PassCard(), true; got != want {
		t.Errorf("PassCard()=%t want=%t", got, want)
	}
	rules.SetPassCard(false)
	if got, want := rules.PassCard(), false; got != want {
		t.Errorf("PassCard()=%t want=%t", got, want)
	}
}

func TestRulesFromStorage(t *testing.T) {
	sr := storage.Rules{
		PassCard: true,
	}

	rules := rulesFromStorage(sr)

	if got, want := rules.PassCard(), true; got != want {
		t.Errorf("PassCard()=%t want=%t", got, want)
	}
}
