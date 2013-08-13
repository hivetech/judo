package main

import (
    "fmt"
    "os/exec"
    "strings"
)

// Error reports the failure of a LXC command.
type Error struct {
	Name   string
	Err    error
	Output []string
}

func (e Error) Error() string {
	if e.Output == nil {
		return fmt.Sprintf("error executing %q: %v", e.Name, e.Err)
	}
	if len(e.Output) == 1 {
		return fmt.Sprintf("error executing %q: %v", e.Name, e.Output[0])
	}
	return fmt.Sprintf("error executing %q: %s", e.Name, strings.Join(e.Output, "; "))
}

var DockerCmd string = "docker"

func init() {
    fmt.Println("Initializing box package...")
}

// run executes the passed command and returns the out.
func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	// LXC tools do not use stdout and stderr in a predictable
	// way; based on experimentation, the most convenient
	// solution is to combine them and leave the client to
	// determine sanity as best it can.
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", runError(name, err, out)
	}
	return string(out), err
}

// runError creates an error if run fails.
func runError(name string, err error, out []byte) error {
	e := &Error{name, err, nil}
	for _, l := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(l, name+": ") {
			// LXC tools do not always print their output with
			// the command name as prefix. The name is part of
			// the error struct, so stip it from the output if
			// printed.
			l = l[len(name)+2:]
		}
		if l != "" {
			e.Output = append(e.Output, l)
		}
	}
	return e
}

func List() (int, error) {
    //out, err := run(DockerCmd, "ps", "-a")
    out, err := run("lxc-ls", "-1")
    if err != nil {
        return 1, err
    }
    fmt.Printf("Containers: %s", out)
    return 0, nil
}

func main() {
    status, error := List() 
    if (error != nil) {
        fmt.Println(status)
    }
}
