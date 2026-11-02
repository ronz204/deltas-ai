package claimset

import "testing"

func TestCheckScenario(t *testing.T) {
	good := []string{
		"Given a cart with two items, when checkout runs with a failing card, then no order is saved.",
		"given x, WHEN y, Then z",
		"Given a cart, when it is, empty, then nothing happens.",
		"Dado que existe la factura 1001, cuando se intenta emitir otra con 1001, entonces se rechaza.",
		"Dada la factura 1001, cuando se repite, entonces se rechaza.",
	}
	for _, in := range good {
		if err := CheckStatement(Acceptance, in); err != nil {
			t.Errorf("rejected %q: %v", in, err)
		}
	}
	bad := []string{
		"",
		"The suite passes.",
		"Given a cart, when it checks out",
		"Given a cart, then an order is saved, when it checks out",
		"When it checks out, then an order is saved",
		"Given , when x, then y",
		"Given a cart, when x, then",
		"Givenacart, when x, then y",
	}
	for _, in := range bad {
		if err := CheckStatement(Acceptance, in); err == nil {
			t.Errorf("accepted %q", in)
		}
	}
}

func TestCheckStatementUnknownKind(t *testing.T) {
	if err := CheckStatement(0, "When x, y"); err == nil {
		t.Fatal("a claim with no kind must not pass")
	}
}
