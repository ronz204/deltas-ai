// Command delta is the composition root: it will build the gateways, hand them to the use
// cases and the use cases to the adapters once there are commands to run.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "delta: no commands yet")
	os.Exit(2)
}
