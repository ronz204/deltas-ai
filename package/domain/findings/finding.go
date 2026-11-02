package findings

import "fmt"

type Diagnostic struct {
	File    string
	Line    int
	Code    string
	Message string
	Warn    bool
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
