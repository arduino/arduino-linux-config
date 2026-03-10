// exec.go
package carrier

import "os/exec"

// execCommand is a variable so it can be replaced in tests.
var execCommand = exec.Command
