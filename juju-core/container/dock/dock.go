// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package dock

import (
	"fmt"
    "bytes"
    "io"
	"io/ioutil"
	"os"
    "os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"launchpad.net/loggo"

	"launchpad.net/juju-core/agent/tools"
	"launchpad.net/juju-core/constraints"
	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/environs/cloudinit"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/instance"
	"launchpad.net/juju-core/names"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/utils"

    "github.com/dotcloud/docker"
    dockerutils "github.com/dotcloud/docker/utils"
)

var logger = loggo.GetLogger("juju.container.docker")

//some pain in the ass!

type lxcInstance struct {
	id string
}

var _ instance.Instance = (*lxcInstance)(nil)

// Id implements instance.Instance.Id.
func (lxc *lxcInstance) Id() instance.Id {
	return instance.Id(lxc.id)
}

func (lxc *lxcInstance) Addresses() ([]instance.Address, error) {
	logger.Errorf("lxcInstance.Addresses not implemented")
	return nil, nil
}

// DNSName implements instance.Instance.DNSName.
func (lxc *lxcInstance) DNSName() (string, error) {
	return "", instance.ErrNoDNSName
}

// WaitDNSName implements instance.Instance.WaitDNSName.
func (lxc *lxcInstance) WaitDNSName() (string, error) {
	return "", instance.ErrNoDNSName
}

// OpenPorts implements instance.Instance.OpenPorts.
func (lxc *lxcInstance) OpenPorts(machineId string, ports []instance.Port) error {
	return fmt.Errorf("not implemented")
}

// ClosePorts implements instance.Instance.ClosePorts.
func (lxc *lxcInstance) ClosePorts(machineId string, ports []instance.Port) error {
	return fmt.Errorf("not implemented")
}

// Ports implements instance.Instance.Ports.
func (lxc *lxcInstance) Ports(machineId string) ([]instance.Port, error) {
	return nil, fmt.Errorf("not implemented")
}

// Add a string representation of the id.
func (lxc *lxcInstance) String() string {
	return fmt.Sprintf("lxc:%s", lxc.id)
}

//COINCOIN IS BACK


var (
    defaultTemplate     = "base"
	containerDir        = "/var/lib/juju/containers"
	removedContainerDir = "/var/lib/juju/removed-containers"
	//lxcContainerDir     = "/var/lib/lxc"
	dockerContainerDir  = "/var/lib/docker/containers"
    //FIXME Autostart is a flag at creation, does this dir exist ?
	lxcRestartDir       = "/etc/lxc/auto"
    //FIXME Might disapear
	aptHTTPProxyRE      = regexp.MustCompile(`(?i)^Acquire::HTTP::Proxy\s+"([^"]+)";$`)
)

const (
	// BridgeNetwork will have the container use the lxc bridge.
	bridgeNetwork = "bridge"
	// PhyscialNetwork will have the container use a specified network device.
	physicalNetwork = "physical"
	// DefaultDockerBridge is the package created container bridge
	DefaultDockerBridge = docker.DefaultNetworkBridge
)

// NetworkConfig defines how the container network will be configured.
type NetworkConfig struct {
	networkType string
	device      string
}

// DefaultNetworkConfig returns a valid NetworkConfig to use the
// defaultDockerBridge that is created by the lxc package.
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{bridgeNetwork, DefaultDockerBridge}
}

// BridgeNetworkConfig returns a valid NetworkConfig to use the specified
// device as a network bridge for the container.
func BridgeNetworkConfig(device string) *NetworkConfig {
	return &NetworkConfig{bridgeNetwork, device}
}

// PhysicalNetworkConfig returns a valid NetworkConfig to use the specified
// device as the network device for the container.
func PhysicalNetworkConfig(device string) *NetworkConfig {
	return &NetworkConfig{physicalNetwork, device}
}

// ManagerConfig contains the initialization parameters for the ContainerManager.
type ManagerConfig struct {
	Name   string
	LogDir string
}

// ContainerManager is responsible for starting containers, and stopping and
// listing containers that it has started.  The name of the manager is used to
// namespace the docker containers on the machine.
type ContainerManager interface {
	// StartContainer creates and starts a new lxc container for the specified machine.
	StartContainer(
		machineId, series, nonce string,
		network *NetworkConfig,
		tools *tools.Tools,
		environConfig *config.Config,
		stateInfo *state.Info,
		apiInfo *api.Info) (instance.Instance, error)
	// StopContainer stops and destroyes the lxc container identified by Instance.
	StopContainer(instance.Instance) error
	// ListContainers return a list of containers that have been started by
	// this manager.
	ListContainers() ([]instance.Instance, error)
}

type containerManager struct {
	name   string
	logdir string
}

// NewContainerManager returns a manager object that can start and stop docker
// containers. The containers that are created are namespaced by the name
// parameter.
func NewContainerManager(conf ManagerConfig) ContainerManager {
	logdir := "/var/log/juju"
	if conf.LogDir != "" {
		logdir = conf.LogDir
	}

    if conf.Name == "" {
        fmt.Errorf("Custm manager name not supported with docker provider, setting to empty (%s)\n", conf.Name)
        conf.Name = ""
    }
    return &containerManager{name: conf.Name, logdir: logdir}
}

func (manager *containerManager) execute(args []string) error {
    //protoAddrParts := strings.SplitN(manager.uri, "://", 2)
    if err:= docker.ParseCommands("tcp", "127.0.0.1:4243", args...); err != nil {
        return fmt.Errorf("** Error docker.ParseCommands: %s\n", err)
    }
    return nil
}

func getLastContainer() (string, error){

    //return "06132b8f2ed79daa999cc517d7f2597d862ab364c550299958d8238df212a12d", nil

    //args := []string{ 
                      //"-lrth", "/var/lib/docker/containers/", "|",
                      //"tail", "-n", "1", "|",
                      //"awk", "'{print $9}'",
                      
                      //" /var/lib/docker/containers/ | tail -n 1 | awk '{print $9}'",
                    //} 
    //var out bytes.Buffer
    //cmd.Stdout = &out  // => out.String()



    //FUCK IT : https://gist.github.com/dagoof/1477401


    c1 := exec.Command("ls", "-lrth", "var/lib/docker/containers/")
    c2 := exec.Command("tail", "-n", "1")
    c3 := exec.Command("awk", "'{print $9}'")

    r,w := io.Pipe()
    c1.Stdout = w
    c2.Stdin = r

    r2,w2 := io.Pipe()
    c2.Stdout = w2
    c3.Stdin = r2

    var b2 bytes.Buffer
    c3.Stdout = &b2

    c1.Start()
    c2.Start()
    c3.Start()
    c1.Wait()
    w.Close()
    c2.Wait()
    w2.Close()
    c3.Wait()

    written, _ := io.Copy(os.Stdout, &b2), nil
    return string(b2), nil


    //cmd := exec.Command("/bin/ls", args...)
    //id , err := cmd.CombinedOutput();
    //if err != nil {
    //    logger.Errorf("failed to get result from cmd CombinedOutput : %v", string(id))
    //    return "",err
    //}
    //return string(id), nil


}

func (manager *containerManager) StartContainer(
	machineId, series, nonce string,
	network *NetworkConfig,
	tools *tools.Tools,
	environConfig *config.Config,
	stateInfo *state.Info,
	apiInfo *api.Info) (instance.Instance, error) {

	name := names.MachineTag(machineId)
	if manager.name != "" {
		name = fmt.Sprintf("%s-%s", manager.name, name)
	}

	// Create the cloud-init.
    directory := jujuContainerDirectory(name)
    logger.Tracef("create directory: %s", directory)
    if err := os.MkdirAll(directory, 0755); err != nil {
        logger.Errorf("failed to create container directory: %v", err)
        return nil, err
    }

	logger.Tracef("write cloud-init")
	//userDataFilename, err := writeUserData(directory, machineId, nonce, tools, environConfig, stateInfo, apiInfo)
	_, err := writeUserData(directory, machineId, nonce, tools, environConfig, stateInfo, apiInfo)
	if err != nil {
		logger.Errorf("failed to write user data: %v", err)
		return nil, err
	}
    /*
     * Note : Docker use a conf.lxc in container.root directory, automatically
     *logger.Tracef("write the lxc.conf file")
	 *configFile, err := writeLxcConfig(network, directory, manager.logdir)
	 *if err != nil {
	 *    logger.Errorf("failed to write config file: %v", err)
	 *    return nil, err
	 *}
     */

    //TODO Find a way to put cloud-config into the container
    command := "apt-get update && apt-get install -y cloud-init && cloud-init init -f /mnt/cloud-config"
	templateParams := []string{
        "run", "-d",  // detach mode
        "-h", name,   // default is id, may be fine
        //TODO How the fuck to use it ?
        //"-v", "/var/log/juju:/var/log/juju",
        "-v", directory + ":/mnt",
		//"--userdata", userDataFilename, // Our groovey cloud-init
        //NOTE Interesting: "-entrypoint", "/var/lib/juju/containers/$directory?"
        //FIXME "-u", "ubuntu", makes the container exit with error (no password found for user ubuntu)
		series,
        "/bin/bash", "-c", command,
	}
	// Create the container.
	logger.Tracef("create the container")
    // Note : bash execution outputs the id, maybe appropriate here ?
    if err := manager.execute(templateParams); err != nil {
        return nil, fmt.Errorf("Create container %s\n", err)
    }
    // Fetching back the id of last created container
    cid, err := getLastContainer(); 
    if err != nil {
        return nil, fmt.Errorf("%s\n",err)
    }

    
    // Committing the container let us set its image name to, well, name
    logger.Tracef("Commit the container")
    if err := manager.execute([]string{"commit", cid, name}); err != nil {
        return nil, fmt.Errorf("Create container %s\n", err)
    }

	// Make sure that the mount dir has been created.
    //FIXME What is this step about ?
	//logger.Tracef("make the mount dir for the shard logs")
	//if err := os.MkdirAll(internalLogDir(container.ID), 0755); err != nil {
		//logger.Errorf("failed to create internal /var/log/juju mount dir: %v", err)
		//return nil, err
	//}
	logger.Tracef("lxc container created")

	// Now symlink the config file into the restart directory.
    /*
     * TODO find out docker way to handle restart
	 *containerConfigFile := filepath.Join(lxcContainerDir, name, "config")
	 *if err := os.Symlink(containerConfigFile, restartSymlink(name)); err != nil {
	 *    return nil, err
	 *}
	 *logger.Tracef("auto-restart link created")
     */

	// Start the lxc container with the appropriate settings for grabbing the
	// console output and a log file.
    // Docker use container.root + id + "-json.log", Symlink to directory + "console.log"
	//consoleFile := filepath.Join(directory, "console.log")
	containerLogFile := filepath.Join(dockerContainerDir, cid, cid + "-json.log")
	if err := os.Symlink(containerLogFile, directory + "console.log"); err != nil {
	    return nil, err
	}
    logger.Tracef("Container logs linked to juju container directory")

    return &lxcInstance{id: name}, nil
}

func (manager *containerManager) StopContainer(instance instance.Instance) error {
	name := string(instance.Id())

	//id := string(instance.DockerId())
    //TODO id := findContainer(name)
    id := ""
    if err := manager.execute([]string{"stop", id}); err != nil {
		logger.Errorf("failed to stop lxc container: %v", err)
		return err
	}
	// Destroy removes the restart symlink for us.
    if err := manager.execute([]string{"rm", id}); err != nil {
		logger.Errorf("failed to destroy lxc container: %v", err)
		return err
	}

	// Move the directory.
	logger.Tracef("create old container dir: %s", removedContainerDir)
	if err := os.MkdirAll(removedContainerDir, 0755); err != nil {
		logger.Errorf("failed to create removed container directory: %v", err)
		return err
	}
	removedDir, err := uniqueDirectory(removedContainerDir, name)
	if err != nil {
		logger.Errorf("was not able to generate a unique directory: %v", err)
		return err
	}
	if err := os.Rename(jujuContainerDirectory(name), removedDir); err != nil {
		logger.Errorf("failed to rename container directory: %v", err)
		return err
	}
	return nil
}

func (manager *containerManager) ListContainers() (result []instance.Instance, err error) {
    autorestart := true
    enablecors := false
    flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%d", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = dockerutils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker/graph", autorestart, enablecors, flHosts)
    if err != nil {
        fmt.Println("** Error creating server")
        return nil,nil
    }

    get_size := true
    all := false  // We're only searching for running containers
    limit := -1   // i.e. no limit, dude
    containers := srv.Containers(all, get_size, limit, "", "")

	managerPrefix := ""
	if manager.name != "" {
		managerPrefix = fmt.Sprintf("%s-", manager.name)
	}

	for _, container := range containers {
		// Filter out those not starting with our name.
		name := container.ID
		if !strings.HasPrefix(name, managerPrefix) {
			continue
		}
        // Note : Should be useless thanks to all = false parameter above
        if !strings.Contains(container.Status, "Exit") {
            result = append(result, &lxcInstance{id: container.Image})
		}
	}
	return
}

func jujuContainerDirectory(containerName string) string {
	return filepath.Join(containerDir, containerName)
}

const internalLogDirTemplate = "%s/%s/rootfs/var/log/juju"

func internalLogDir(containerName string) string {
	//return fmt.Sprintf(internalLogDirTemplate, lxcContainerDir, containerName)
	return fmt.Sprintf(internalLogDirTemplate, dockerContainerDir, containerName)
}

func restartSymlink(name string) string {
	return filepath.Join(lxcRestartDir, name+".conf")
}

const localConfig = `%s
lxc.mount.entry=%s var/log/juju none defaults,bind 0 0
`

const networkTemplate = `
lxc.network.type = %s
lxc.network.link = %s
lxc.network.flags = up
`

func networkConfigTemplate(networkType, networkLink string) string {
	return fmt.Sprintf(networkTemplate, networkType, networkLink)
}

func generateNetworkConfig(network *NetworkConfig) string {
	if network == nil {
		logger.Warningf("network unspecified, using default networking config")
		network = DefaultNetworkConfig()
	}
	switch network.networkType {
	case physicalNetwork:
		return networkConfigTemplate("phys", network.device)
	default:
		logger.Warningf("Unknown network config type %q: using bridge", network.networkType)
		fallthrough
	case bridgeNetwork:
		return networkConfigTemplate("veth", network.device)
	}
}

func writeLxcConfig(network *NetworkConfig, directory, logdir string) (string, error) {
	networkConfig := generateNetworkConfig(network)
	//configFilename := filepath.Join(directory, "config.lxc")
    configFilename := filepath.Join(directory, "lxc.conf")
	configContent := fmt.Sprintf(localConfig, networkConfig, logdir)
	if err := ioutil.WriteFile(configFilename, []byte(configContent), 0644); err != nil {
		return "", err
	}
	return configFilename, nil
}

func writeUserData(
	directory, machineId, nonce string,
	tools *tools.Tools,
	environConfig *config.Config,
	stateInfo *state.Info,
	apiInfo *api.Info,
) (string, error) {
	userData, err := cloudInitUserData(machineId, nonce, tools, environConfig, stateInfo, apiInfo)
	if err != nil {
		logger.Errorf("failed to create user data: %v", err)
		return "", err
	}
	userDataFilename := filepath.Join(directory, "cloud-init")
	if err := ioutil.WriteFile(userDataFilename, userData, 0644); err != nil {
		logger.Errorf("failed to write user data: %v", err)
		return "", err
	}
	return userDataFilename, nil
}

func cloudInitUserData(
	machineId, nonce string,
	tools *tools.Tools,
	environConfig *config.Config,
	stateInfo *state.Info,
	apiInfo *api.Info,
) ([]byte, error) {
	machineConfig := &cloudinit.MachineConfig{
		MachineId:            machineId,
		MachineNonce:         nonce,
		MachineContainerType: instance.LXC,
		StateInfo:            stateInfo,
		APIInfo:              apiInfo,
		DataDir:              "/var/lib/juju",
		Tools:                tools,
	}
	if err := environs.FinishMachineConfig(machineConfig, environConfig, constraints.Value{}); err != nil {
		return nil, err
	}
	cloudConfig, err := cloudinit.New(machineConfig)
	if err != nil {
		return nil, err
	}

	// Run apt-config to fetch proxy settings from host. If no proxy
	// settings are configured, then we don't set up any proxy information
	// on the container.
	proxyConfig, err := utils.AptConfigProxy()
	if err != nil {
		return nil, err
	}
	if proxyConfig != "" {
		var proxyLines []string
		for _, line := range strings.Split(proxyConfig, "\n") {
			line = strings.TrimSpace(line)
			if m := aptHTTPProxyRE.FindStringSubmatch(line); m != nil {
				cloudConfig.SetAptProxy(m[1])
			} else {
				proxyLines = append(proxyLines, line)
			}
		}
		if len(proxyLines) > 0 {
			cloudConfig.AddFile(
				"/etc/apt/apt.conf.d/99proxy-extra",
				strings.Join(proxyLines, "\n"),
				0644)
		}
	}

	// Run ifconfig to get the addresses of the internal container at least
	// logged in the host.
	cloudConfig.AddRunCmd("ifconfig")

	data, err := cloudConfig.Render()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// uniqueDirectory returns "path/name" if that directory doesn't exist.  If it
// does, the method starts appending .1, .2, etc until a unique name is found.
func uniqueDirectory(path, name string) (string, error) {
	dir := filepath.Join(path, name)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return dir, nil
	}
	for i := 1; ; i++ {
		dir := filepath.Join(path, fmt.Sprintf("%s.%d", name, i))
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return dir, nil
		} else if err != nil {
			return "", err
		}
	}
}
