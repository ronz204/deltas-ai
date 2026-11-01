package fail

import "testing"

func TestINV02_Fails(t *testing.T) { t.Fatal("this claim is violated") }
