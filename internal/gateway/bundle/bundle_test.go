package bundle

import (
	"strings"
	"testing"
)

func TestSpecIsTheSpecTemplate(t *testing.T) {
	src, err := Bundle{}.Spec()
	if err != nil || !strings.Contains(string(src), "format: 1") {
		t.Fatalf("Spec() = %.60q, %v", src, err)
	}
}
