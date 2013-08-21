// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

// godocker - Go package to interact with Linux Containers (LXC) through docker
//
// (https://github.com/dotcloud/docker).
//
package godocker

import (
	"fmt"
    "os"
	"os/exec"
	"strings"
    "strconv"
    "io/ioutil"
    "encoding/json"

    "github.com/dotcloud/docker"
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

// State represents a container state.
type State string

const (
    StateUnknown  State = "UNKNOWN"
    StateStopped  State = "STOPPED"
    StateStarting State = "STARTING"
    StateRunning  State = "RUNNING"
    StateAborting State = "ABORTING"
    StateStopping State = "STOPPING"
)

// LogLevel represents a container's log level.
type LogLevel string

const (
	LogDebug    LogLevel = "DEBUG"
	LogInfo     LogLevel = "INFO"
	LogNotice   LogLevel = "NOTICE"
	LogWarning  LogLevel = "WARN"
	LogError    LogLevel = "ERROR"
	LogCritical LogLevel = "CRIT"
	LogFatal    LogLevel = "FATAL"
)

// Container represents a linux container instance and provides
// operations to create, maintain and destroy the container.
type Container interface {

	// Name returns the name of the container.
	Name() string

    // After creation, change Name() to docker container id
    SetId(id string)

	// Create creates a new container based on the given template.
	Create(configFile, template string, templateArgs ...string) error

	// Start runs the container as a daemon.
	Start(configFile, consoleFile string) error

	// Stop terminates the running container.
	Stop() error

	// Clone creates a copy of the container, giving the copy the specified name.
	Clone(name string) (Container, error)

	// Freeze freezes all the container's processes.
	Freeze() error

	// Unfreeze thaws all frozen container's processes.
	Unfreeze() error

	// Destroy stops and removes the container.
	Destroy() error

	// Wait waits for one of the specified container states.
	Wait(states ...State) error

	// Info returns the status and the process id of the container.
	Info() (State, int, error)

	// IsConstructed checks if the container image exists.
	IsConstructed() bool

	// IsRunning checks if the state of the container is 'RUNNING'.
	IsRunning() bool

	// String returns information about the container, like the name, state,
	// and process id.
	String() string

	// LogFile returns the current filename used for the LogFile.
	LogFile() string

	// LogLevel returns the current logging level (only used if the
	// LogFile is not "").
	LogLevel() LogLevel

	// SetLogFile sets both the LogFile and LogLevel.
	SetLogFile(filename string, level LogLevel)

    // Get container informations
    GetKeyInfo(key string) (interface{}, error)

    // Get container informations
    containerHome() string
}

// ContainerFactory represents the methods used to create Containers.
type ContainerFactory interface {
	// New returns a container instance which can then be used for operations
	// like Create(), Start(), Stop() or Destroy().
	New(string) Container

	// List returns all the existing containers on the system.
	List() ([]Container, error)
}

// Factory provides the standard ContainerFactory.
func Factory() ContainerFactory {
	return &containerFactory{}
}

type container struct {
	name     string
	nickname string
	logFile  string
	logLevel LogLevel
}

type containerFactory struct{}

//FIXME glxc tests compliant but "what the fuck" regarding c.containerHome() ?
func ContainerHome(c Container) string {
	return c.containerHome()
}

func (*containerFactory) New(name string) Container {
	return &container{
		name     : name,
		nickname : name,
		logLevel : LogWarning,
	}
}

func (factory *containerFactory) List() ([]Container, error) {
    out, err := run("docker", "ps", "-a")
    if err != nil {
        return nil, err
    }
    names := nameSet(out)
	containers := make([]Container, len(names))
	for i, name := range names {
		containers[i] = factory.New(name)
	}
	return containers, nil
}

// Name returns the name of the container.
func (c *container) Name() string {
	return c.nickname
}

func (c *container) SetId(id string) {
    c.name = id
}

// LogFile returns the current filename used for the LogFile.
func (c *container) LogFile() string {
	return c.logFile
}

// LogLevel returns the current logging level, this is only used if the
// LogFile is not "".
func (c *container) LogLevel() LogLevel {
	return c.logLevel
}

// SetLogFile sets both the LogFile and LogLevel.
func (c *container) SetLogFile(filename string, level LogLevel) {
    // LogFile with docker is automatic
    //c.logFile = filename
    complete_id, _ := c.GetKeyInfo("ID")
    c.logFile = fmt.Sprintf("%s/%s-jdon.log", c.containerHome(), complete_id.(string))
	c.logLevel = level
}

// Create creates a new container based on the given template.
func (c *container) Create(configFile, template string, templateArgs ...string) error {
    //NOTE 1st pass: configFile and templateArgs ignored
    // Configfile automatic location: containerHome() + "/" + long_id + "/config.json" 
    configFile = ""
    templateArgs = []string{}
    // template = image ! ex: ubuntu or app/node-js-sample
    // series the tag !   ex: v1 in app/node-js-sample:v1

    /*
     * Is template:(-r series) on the system ?
     *  oui, mais pas container    -> docker run template:series /bin/bash -c "/check_validity.sh"
     *  oui, mais container & run  -> alors erreur "already created and runnning"
     *                      & stop -> 'void'
     *  !!TODO : if template exist but series doesn't 
     *  
     *  non     -> docker build -t c.name dir/with/dockerfile_qui_contient | url
     *  ou: non -> docker pull c.name && docker run c.name /bin/bash -c "echo \"pouet\""
     */
    if ! IsExisting(template) {
		return fmt.Errorf("No base image found for container %q", c.Name())
    }

	if c.IsConstructed() {
		return fmt.Errorf("container %q is already created", c.Name())
	}

    cid_filename_array := []string{"/tmp/", c.name, ".cid"}
    cid_filename := strings.Join(cid_filename_array, "")
    if err := os.Remove(cid_filename); err != nil {
		fmt.Printf("Error removing %s: %s", cid_filename, err)
	}

    dummy_command := "ps"
	args := []string{
        "run", "-d",  // daemon
        "-h", c.name,
        "-cidfile", cid_filename,
        template,
        "/bin/bash", "-c", dummy_command,
	}

    /*
	 *if configFile != "" {
	 *    args = append(args, "-f", configFile)
	 *}
	 *if len(templateArgs) != 0 {
     *    args = append(args, "--")
	 *    args = append(args, templateArgs...)
	 *}
     */
     if _, err := run("docker", args...); err != nil {
		return fmt.Errorf("Could not run docker container %s", strings.Join(args, " "))
	}

    id, err := ioutil.ReadFile(cid_filename)
    if err != nil {
		return fmt.Errorf("Could not read cid file")
    }
    c.SetId(string(id))
	return nil
}

// Start runs the container as a daemon.
func (c *container) Start(configFile, consoleFile string) error {
    // With docker, log and config files have automatic location
	if !c.IsConstructed() {
		return fmt.Errorf("container %q is not yet created", c.name)
	}
	args := []string{
		"--daemon",
		"-n", c.name,
	}
	if configFile != "" {
		args = append(args, "-f", configFile)
	}
	if consoleFile != "" {
		args = append(args, "-c", consoleFile)
	}
	if c.logFile != "" {
		args = append(args, "-o", c.logFile, "-l", string(c.logLevel))
	}
	_, err := run("docker run", args...)
	if err != nil {
		return err
	}
	return c.Wait(StateRunning)
}

// Stop terminates the running container.
func (c *container) Stop() error {
	if !c.IsConstructed() {
		return fmt.Errorf("container %q is not yet created", c.name)
	}
	args := []string{
		"stop", c.name,
	}
	_, err := run("docker", args...)
	if err != nil {
		return err
	}
	return c.Wait(StateStopped)
}

// Clone creates a copy of the container, it gets the given name.
func (c *container) Clone(name string) (Container, error) {
	if !c.IsConstructed() {
		return nil, fmt.Errorf("container %q is not yet created", c.name)
	}
	cc := &container{
		name: name,
	}
	if cc.IsConstructed() {
		return cc, nil
	}
	args := []string{
		c.name,
		name,
	}
	_, err := run("docker commit", args...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// Docker doesn't need you !
func (c *container) Freeze() error {
    return fmt.Errorf("not implemented")
}

// Docker doesn't need you !
func (c *container) Unfreeze() error {
    return fmt.Errorf("not implemented")
}

// Destroy stops and removes the container.
func (c *container) Destroy() error {
	if !c.IsConstructed() {
		return fmt.Errorf("container %q is not yet created", c.name)
	}
	if err := c.Stop(); err != nil {
		return err
	}
    _, err := run("docker", "rm", c.name)
	if err != nil {
		return err
	}

    cid_filename_array := []string{"/tmp/", c.name, ".cid"}
    cid_filename := strings.Join(cid_filename_array, "")
    if err := os.Remove(cid_filename); err != nil {
		fmt.Printf("Error removing %s: %s", cid_filename, err)
	}
	return nil
}

// Wait waits for one of the specified container states.
func (c *container) Wait(states ...State) error {
    //NOTE Only STOPPED state is implemented
	if len(states) == 0 {
		return fmt.Errorf("no states specified")
	}
	stateStrs := make([]string, len(states))
	for i, state := range states {
        //TODO Check if STOPPED is present
		stateStrs[i] = string(state)
	}
	waitStates := strings.Join(stateStrs, "|")
    fmt.Printf(waitStates)  //my awesome log
	_, err := run("docker", "wait", c.name) 
	if err != nil {
		return err
	}
	return nil
}

// Info returns the status and the process id of the container.
func (c *container) Info() (State, int, error) {
    //TODO Use GetKeyInfo
    // The pid first
    args := []string{"-ef", "|", "grep", c.name, "|", "grev", "-v", "grep", "|", "awk", "'{printf $2}'"}
	pid_str, err := run("ps", args...)
	if err != nil {
		return StateUnknown, -1, err
	}

    // And now container state
    state := State("UNKNOWN")
    if out, _ := run("docker", "ps", "|", "grep", c.name); out != "" {
        state = State("RUNNING")
    } else if out, _ := run("docker", "ps", "-a", "|", "grep", c.name); out != "" {
        state = State("STOPPED")
    }

	pid, err := strconv.Atoi(pid_str)
	if err != nil {
		return StateUnknown, -1, fmt.Errorf("cannot read the pid: %v", err)
	}
	return state, pid, nil
}

// IsConstructed checks if the container image exists.
func (c *container) IsConstructed() bool {
    if _, err := os.Stat(c.containerHome()); os.IsNotExist(err) {
        return false
    }
    //out, err := run("docker", "ps", "-a", "|", "grep", c.name)
	//if out == "" || err != nil {
		//return false
	//}
	return true
}

// IsExisting checks if the image exists.
func IsExisting(name string) bool {
    //FIXME Tag unsupported
    //TODO Check first locally
    str_tab := strings.Split(name, ":")
    template := str_tab[0]
    // Last agument is a dns file
    srv, err := docker.NewServer("/var/lib/docker/graph", false, true, nil)
    if err != nil {
        fmt.Println("** Error creating server")
        os.Exit(-1)
    }

    if images, err := srv.ImagesSearch(template); err != nil {
        fmt.Errorf("Error searching for an image\n")
        return false
    } else {
        if len(images) == 0 {
            return false
        }
    }
    return true


    //TODO Check repos as well: return false if not on the system and pull fails
    //args := []string{"images", "|", "grep ", name}
    //out, err := run("docker", args...)
    //data := []byte(out + name)
    
    //cid_filename_array := []string{"tmp.txt"}
    //cid_filename := strings.Join(cid_filename_array, "")
    //_, err = run("rm", cid_filename)
    //ioutil.WriteFile("tmp.txt", data, 0666)
	//if out == "" || err != nil {
		//return false
	//}
	//return true
}

// IsRunning checks if the state of the container is 'RUNNING'.
func (c *container) IsRunning() bool {
	state, _, err := c.Info()
	if err != nil {
		return false
	}
	return state == StateRunning
}

// String returns information about the container.
func (c *container) String() string {
	state, pid, err := c.Info()
	if err != nil {
		return fmt.Sprintf("cannot retrieve container info for %q: %v", c.name, err)
	}
	return fmt.Sprintf("container %q (%s, pid %d)", c.name, state, pid)
}

// containerHome returns the name of the container directory.
func (c *container) containerHome() string {
    // c.name is a truncated version of the id, we need the original
    complete_id, err := c.GetKeyInfo("ID")
    if err != nil {
        return ""
    }
	return "/var/lib/docker/containers" + "/" + complete_id.(string)
}

// rootfs returns the name of the directory containing the
// root filesystem of the container.
func (c *container) rootfs() string {
	return c.containerHome() + "/rootfs/"
}

func (c *container) GetKeyInfo(key string) (interface{}, error) {
    cid_filename_array := []string{"/tmp/", c.nickname, ".cid"}
    cid_filename := strings.Join(cid_filename_array, "")
    id, err := ioutil.ReadFile(cid_filename)
    if err != nil {
		return nil, fmt.Errorf("Could not read cid file")
    }
    args := []string{"inspect", string(id)}
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", runError("docker", err, out)
    }
	//} else if strings.Contains(string(out), "Error") {
        //return "", fmt.Errorf("Could not inspect %s", c.name)
    //}

    // Parse the json answer
    var f interface{}
    //NOTE Trick to avoid root array
    err = json.Unmarshal(out[1:len(out)-1], &f)
    if err != nil {
        return "", fmt.Errorf("Could not decode informations for %s\n", c.name)
    }
    msg := f.(map[string]interface{})

    return msg[key], nil
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
	return string(out), nil
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

// nameSet retrieves a set of names out of a command out.
func nameSet(raw string) []string {
    collector := map[string]struct{}{}
    set := []string{}
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
        // Skip header and last empty line
        if (i != 0 && i < 6) {
            fields := strings.Fields(line)
            var name string = fields[0]
            if name != "" {
                collector[name] = struct{}{}
            }
        }
    }
    for name := range collector {
        set = append(set, name)
    }
    return set
}

/*
 *type LXCSuite struct {
 *    factory ContainerFactory
 *}
 *
 *func main() {
 *    var f containerFactory = containerFactory{}
 *    ids, err := f.List()
 *    if (err != nil) {
 *        fmt.Println("ERROR")
 *    }
 *    for _, box := range ids {
 *        if box.IsConstructed() {
 *            fmt.Printf("Box %s exists\n", box.Name())
 *            id, _ := box.GetKeyInfo("ID")
 *            fmt.Printf("\tId:\n", id)
 *            path, _ := box.GetKeyInfo("Path")
 *            fmt.Printf("\tPath:\n", path)
 *            state, _ := box.GetKeyInfo("State")
 *            fmt.Printf("\tState:\n", state)
 *            if box.Name() == "6ef7bc15a11b" {
 *                fmt.Println("Stopping This f*** box")
 *                //box.Stop()
 *            }
 *        }
 *    }
 *}
 */
