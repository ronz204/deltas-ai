// Package diag holds the finding every validator reports.
package diag

import "fmt"

// Diagnostic is a problem found in a file. Hooks and CI key on Code, never on Message.
type Diagnostic struct {
	File    string
	Line    int // 1-based; 0 when the finding is about the file as a whole
	Code    string
	Message string
	Warn    bool // reported, but it does not fail a run
}

func (d Diagnostic) String() string {
	code := d.Code
	if d.Warn {
		code = "warning " + code
	}
	if d.Line == 0 {
		return fmt.Sprintf("%s: %s: %s", d.File, code, d.Message)
	}
	return fmt.Sprintf("%s:%d: %s: %s", d.File, d.Line, code, d.Message)
}
