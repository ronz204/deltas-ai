package claim

import "testing"

func TestCheckEARS(t *testing.T) {
	good := []struct{ name, in string }{
		{"ubiquitous", "Stock never imports from checkout."},
		{"ubiquitous with shall", "The system shall reject duplicate sequence numbers."},
		{"event", "When a percentage discount is applied, the amount rounds half up to the nearest cent."},
		{"state", "While a reservation is older than 15 minutes, it does not count as reserved."},
		{"optional", "Where a tax rate is set, tax is computed on the discounted subtotal."},
		{"unwanted", "If payment fails, then the stock reservation is released."},
		{"complex", "While the cart is open, when an item is added, the total is recomputed."},
		{"trigger with commas", "When a, b or c happens, the system reacts."},
		{"upper case", "WHEN it rains, the roof holds."},
		{"spanish event", "Cuando se emite una factura, el sistema debe redondear cada línea a 2 decimales."},
		{"spanish unwanted", "Si el pago falla, entonces se libera la reserva."},
		{"spanish state", "Mientras la reserva tenga más de 15 minutos, no cuenta como reservada."},
		{"padded", "  When x, y  "},
	}
	for _, c := range good {
		t.Run(c.name, func(t *testing.T) {
			if err := CheckStatement(Invariant, c.in); err != nil {
				t.Fatalf("rejected %q: %v", c.in, err)
			}
		})
	}
	bad := []struct{ name, in string }{
		{"empty", ""},
		{"blank", "   "},
		{"no comma", "When a discount is applied the amount rounds."},
		{"keyword alone", "When"},
		{"empty clause", "When , the amount rounds."},
		{"no response", "When a discount is applied,"},
		{"if without then", "If payment fails, the reservation is released."},
		{"if without response", "If payment fails, then"},
		{"complex without response", "While open, when added,"},
	}
	for _, c := range bad {
		t.Run("bad/"+c.name, func(t *testing.T) {
			if err := CheckStatement(Invariant, c.in); err == nil {
				t.Fatalf("accepted %q", c.in)
			}
		})
	}
}
