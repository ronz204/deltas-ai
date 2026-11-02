package english

import "testing"

func TestEnglishPasses(t *testing.T) {
	for _, text := range []string{
		"",
		"When a payment is refunded, the balance returns to the customer.",
		"The café menu may show the name José once; that is a name, not a language.",
		"- INV-01 [test: go test ./billing]\n  When the total is zero, the invoice is not issued.",
		"If you do not know it, ask the user; do not invent it for the sake of the contract.",
	} {
		if NotEnglish(text) {
			t.Errorf("%q was taken for another language", text)
		}
	}
}

func TestSpanishIsCaught(t *testing.T) {
	for _, text := range []string{
		"Cuando el usuario paga una factura, el saldo se actualiza para que el cliente vea el total.",
		"La facturación nunca debe cobrar dos veces; también debe registrar cada intento.",
		"Validación, conciliación, atención y devolución.",
	} {
		if !NotEnglish(text) {
			t.Errorf("%q passed as English", text)
		}
	}
}
