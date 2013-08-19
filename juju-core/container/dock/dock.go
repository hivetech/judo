// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package dock

import (
	"fmt"
	"io/ioutil"
	"os"
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
    srv docker.Server
    uri string
}

// NewContainerManager returns a manager object that can start and stop docker
// containers. The containers that are created are namespaced by the name
// parameter.
func NewContainerManager(conf ManagerConfig) ContainerManager {
	logdir := "/var/log/juju"
	if conf.LogDir != "" {
		logdir = conf.LogDir
	}

    autorestart := true
    enablecors := false
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%d", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = dockerutils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    server, err := docker.NewServer("/var/lib/docker/graph", autorestart, enablecors, flHosts)
    if err != nil {
        fmt.Println("** Error creating server")
        return nil
    }
    api := server.DockerVersion()
    fmt.Printf("Docker version: %s\n", api.Version)

    if conf.Name == "" {
        fmt.Errorf("Custm manager name not supported with docker provider, setting to empty (%s)\n", conf.Name)
        conf.Name = ""
    }
    return &containerManager{name: conf.Name, logdir: logdir, srv: *server, uri: flHosts[0]}
}

func (manager *containerManager) execute(args []string) error {
    protoAddrParts := strings.SplitN(manager.uri, "://", 2)
    if err:= docker.ParseCommands(protoAddrParts[0], protoAddrParts[1], args...); err != nil {
        return fmt.Errorf("** Error docker.ParseCommands: %s\n", err)
    }
    return nil
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
    command := "apt-get update && apt-get install cloud-init && cloud-init -f /mnt/cloud-config"
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
    // Fetching back the id
    // Note: Image has to be specified with its tag
    container := docker.APIContainers{ID: ""}
    containers := manager.srv.Containers(false, false, -1, "", "")
    // Note: Use the date instead ?
    for _, box := range containers {
        if box.Image == series && box.Command == "/bin/bash -c " + command {
            container = box
        }
    }
    if container.ID == "" {
        return nil, fmt.Errorf("Container not found, creation might have failed")
    }
    // Committing the container let us set its image name to, well, name
    logger.Tracef("Commit the container")
    if err := manager.execute([]string{"commit", container.ID, name}); err != nil {
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
	containerLogFile := filepath.Join(dockerContainerDir, container.ID, container.ID + "-json.log")
	if err := os.Symlink(containerLogFile, directory + "console.log"); err != nil {
	    return nil, err
	}
    logger.Tracef("Container logs linked to juju container directory")

    return &lxcInstance{id: name, dockerid: container.ID}, nil
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
    get_size := true
    all := true  // We're only searching for running containers
    limit := -1   // i.e. no limit, dude
    containers := manager.srv.Containers(all, get_size, limit, "", "")

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
            result = append(result, &lxcInstance{id: container.Image, dockerid: container.ID})
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
