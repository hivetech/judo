// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package dock

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"launchpad.net/loggo"

	"launchpad.net/juju-core/tools"
	"launchpad.net/juju-core/constraints"
	"launchpad.net/juju-core/environs"
    "launchpad.net/juju-core/environs/cloudinit"
	"launchpad.net/juju-core/environs/ansible"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/instance"
	"launchpad.net/juju-core/names"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/state/api"
	"launchpad.net/juju-core/utils"

    "github.com/dotcloud/docker"
    dockerutils "github.com/dotcloud/docker/utils"

    "github.com/garyburd/redigo/redis"
)

var logger = loggo.GetLogger("juju.container.dock")

var (
    defaultTemplate     = "base"
	containerDir        = "/var/lib/juju/containers"
	removedContainerDir = "/var/lib/juju/removed-containers"
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
        fmt.Errorf("Custom manager name not supported with docker provider, setting to empty (%s)\n", conf.Name)
        conf.Name = ""
    }
    return &containerManager{name: conf.Name, logdir: logdir}
}

func (manager *containerManager) dockerCall(args []string) error {
    //protoAddrParts := strings.SplitN(manager.uri, "://", 2)
    //if err:= docker.ParseCommands("tcp", "127.0.0.1:4243", args...); err != nil {
    if err:= docker.ParseCommands("unix", docker.DEFAULTUNIXSOCKET, args...); err != nil {
        return fmt.Errorf("** Error docker.ParseCommands: %s\n", err)
    }
    return nil
}

func FromNameToId(name string) (string, error) {
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%d", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = dockerutils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        return "", fmt.Errorf("** Error creating server: %v", err)
    }

    containers := srv.Containers(false, false, -1, "", "")

    // Search last created container with image "series"
    for _, container := range containers {
        if container.Image == name + ":latest" {
            return container.ID, nil
        }
    }
    return "", fmt.Errorf("No cointainer found with image %s", name)
}

func getLastContainer(series string) (string, error){
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%d", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = dockerutils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        return "", fmt.Errorf("** Error creating server: %v", err)
    }

    get_size := true
    all := false  // We're only searching for running containers
    limit := -1   // i.e. no limit, dude
    containers := srv.Containers(all, get_size, limit, "", "")

    // Search last created container with image "series"
    target_container := docker.APIContainers{}
    var last_time int64 = -1
    for _, container := range containers {
        //if strings.Contains(series, container.Image) {
        if container.Image == series + ":latest" {
            // Selecting the last one
            if container.Created > last_time {
                last_time = container.Created
                target_container = container
            }
        }
    }

    if last_time == -1 {
        // We found nothing
        return "", fmt.Errorf("No cointainer found with image %s\n", series)
    }
    return target_container.ID, nil
}

func allocateNewSSHPort() (int64, error) {
    c, err := redis.Dial("tcp", ":6379")
    if err != nil {
        return 0, fmt.Errorf("** Error connecting to redis server: %v", err)
    }
    defer c.Close()

    ssh_forwarded_port, err := redis.Int64(c.Do("DECR", "default_ssh_port"))
    if err != nil {
        return 0, fmt.Errorf("** Error asking for fowarded ssh port\n")
    }
    return ssh_forwarded_port, nil
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
	_, err := writeUserData(directory, machineId, nonce, tools, environConfig, stateInfo, apiInfo)
	if err != nil {
		logger.Errorf("failed to write user data: %v", err)
		return nil, err
	}

    /*
     *Note : Docker use a conf.lxc in container.root directory, automatically
     */

    ssh_forwarded_port, err := allocateNewSSHPort()
    if err != nil {
        return nil, fmt.Errorf("** Error asking for new ssh port: %v", err)
    }
    logger.Tracef("Got forwarded ssh port: %d\n", ssh_forwarded_port)
    
    // Default is ansible configuration
    // Prepare ansible for execution. Mainly set default parameters
    playbook_cfg := ansible.NewPlaybook(name, ssh_forwarded_port, "quant", directory)
    command := ansible.DockerCmd
    logger.Tracef("Created playbook configuration\n")
    if environConfig.Initializer() == "cloudinit" {
        command = "cloud-init -f /mnt/cloud-init init "
        logger.Tracef("Using default cloud-init configuration\n")
    }

	logger.Tracef("Create the original container")
	image_name := strings.Split(series, ":")
    //FIXME Lot of hard coded stuff here...
    cmd := exec.Command("/home/xavier/dev/goworkspace/bin/init-juju-image.sh", image_name[0], name)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("Running init-juju-image: %v", err)
    }

	logger.Tracef("Create final container")
    //command := "while true; do sleep 300; done"

	templateParams := []string{
        "run", "-d",  // detach mode
        "-p", fmt.Sprintf("%d:22", ssh_forwarded_port),
        "-h", name,   // default is id, may be fine
        "-v", "/home/xavier/.juju/hive/log:/var/log/juju",
        "-v", directory + ":/mnt",
        //"-u", "ubuntu",  //FIXME makes the container exit with error (no password found for user ubuntu)
		name,
        "/bin/bash", "-c", command,
	}
	// Create the container.
	logger.Tracef("create the container")
    // Note : bash execution outputs the id, maybe appropriate here ?
    if err := manager.dockerCall(templateParams); err != nil {
        return nil, fmt.Errorf("** Create container: %v\n", err)
    }
    // Fetching back the id of last created container
    cid, err := getLastContainer(name); 
    if err != nil {
        return nil, fmt.Errorf("%v\n",err)
    }
    logger.Tracef("Got new container id: %s (%s)\n", cid, name)

	// Make sure that the mount dir has been created.
    //FIXME What is this step about ?
    logger.Tracef("make the mount dir for the shard logs")
    if err := os.MkdirAll(internalLogDir(cid), 0755); err != nil {
        logger.Errorf("failed to create internal /var/log/juju mount dir: %v", err)
        return nil, err
    }
	logger.Tracef("Dock container created")

    //TODO find out docker way to handle restart

	// Start the lxc container with the appropriate settings for grabbing the
	// console output and a log file.
    // Docker use container.root + id + "-json.log", Symlink to directory + "console.log"
	//consoleFile := filepath.Join(directory, "console.log")
    //FIXME Juju seems to retry start when failed, and fails again because this link already exists anyway
	containerLogFile := filepath.Join(dockerContainerDir, cid, cid + "-json.log")
	if err := os.Symlink(containerLogFile, directory + "/console.log"); err != nil {
	    return nil, err
	}
    logger.Tracef("Container logs linked to juju container directory")

    if environConfig.Initializer() == "ansible" {
        if err := ansible.SuitItUp(*playbook_cfg); err != nil {
            return nil, fmt.Errorf("** Deploy ansible: %v", err)
        }
        logger.Tracef("Ansible deployed")
    }

    return &dockInstance{id: name}, nil
}

func (manager *containerManager) StopContainer(instance instance.Instance) error {
	name := string(instance.Id())

	//id := string(instance.DockerId())
    id, err := FromNameToId(name)
    if err != nil {
        return fmt.Errorf("%v", err)
    }
    if err := manager.dockerCall([]string{"stop", id}); err != nil {
		logger.Errorf("failed to stop dock container: %v", err)
		return err
	}
	// Destroy removes the restart symlink for us.
    if err := manager.dockerCall([]string{"rm", id}); err != nil {
		logger.Errorf("failed to destroy dock container: %v", err)
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
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%d", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = dockerutils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        return nil, fmt.Errorf("** Error creating server: %v", err)
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
		//name := container.ID
        name := strings.Split(container.Image, ":")
		if !strings.HasPrefix(name[0], managerPrefix) {
			continue
		}
        // Note : Should be useless thanks to all = false parameter above
        if !strings.Contains(container.Status, "Exit") {
            result = append(result, &dockInstance{id: name[0]})
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

func NetworkConfigTemplate(networkType, networkLink string) string {
	return fmt.Sprintf(networkTemplate, networkType, networkLink)
}

func GenerateNetworkConfig(network *NetworkConfig) string {
	if network == nil {
		logger.Warningf("network unspecified, using default networking config")
		network = DefaultNetworkConfig()
	}
	switch network.networkType {
	case physicalNetwork:
		return NetworkConfigTemplate("phys", network.device)
	default:
		logger.Warningf("Unknown network config type %q: using bridge", network.networkType)
		fallthrough
	case bridgeNetwork:
		return NetworkConfigTemplate("veth", network.device)
	}
}

func writeLxcConfig(network *NetworkConfig, directory, logdir string) (string, error) {
	networkConfig := GenerateNetworkConfig(network)
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
		MachineContainerType: instance.DOCK,
		StateInfo:            stateInfo,
		APIInfo:              apiInfo,
		DataDir:              "/var/lib/juju",
		Tools:                tools,
	}
	if err := environs.FinishMachineConfig(machineConfig, environConfig, constraints.Value{}); err != nil {
		return nil, err
	}

    //TODO apt-proxy support below
    cloudConfig, err := ansible.New(machineConfig)
    if err != nil {
        return nil, err
    }
    if environConfig.Initializer() == "cloudinit" {
        cloudConfig, err := cloudinit.New(machineConfig)

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
    }

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
