// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package provisioner

import (
	"os"

	"launchpad.net/loggo"

	"launchpad.net/juju-core/agent/tools"
	"launchpad.net/juju-core/constraints"
	"github.com/Gusabi/judo/container/dock"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/instance"
	"launchpad.net/juju-core/juju/osenv"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/state/api"
)

var dockLogger = loggo.GetLogger("juju.provisioner.dock")

var _ Broker = (*dockBroker)(nil)

func NewdockBroker(config *config.Config, tools *tools.Tools) Broker {
	return &dockBroker{
		manager: dock.NewContainerManager(dock.ManagerConfig{Name: "juju"}),
		config:  config,
		tools:   tools,
	}
}

type dockBroker struct {
	manager dock.ContainerManager
	config  *config.Config
	tools   *tools.Tools
}

func (broker *dockBroker) StartInstance(machineId, machineNonce string, series string, cons constraints.Value, info *state.Info, apiInfo *api.Info) (instance.Instance, *instance.HardwareCharacteristics, error) {
	dockLogger.Infof("starting dock container for machineId: %s", machineId)

	// Default to using the host network until we can configure.
    //Note: osenv.JujuLxcBridge ok for docker use ?
	bridgeDevice := os.Getenv(osenv.JujuLxcBridge)
	if bridgeDevice == "" {
		bridgeDevice = dock.DefaultDockerBridge
	}
	network := dock.BridgeNetworkConfig(bridgeDevice)

	inst, err := broker.manager.StartContainer(machineId, series, machineNonce, network, broker.tools, broker.config, info, apiInfo)
	if err != nil {
		dockLogger.Errorf("failed to start container: %v", err)
		return nil, nil, err
	}
	dockLogger.Infof("started dock container for machineId: %s, %s", machineId, inst.Id())
	return inst, nil, nil
}

// StopInstances shuts down the given instances.
func (broker *dockBroker) StopInstances(instances []instance.Instance) error {
	// TODO: potentially parallelise.
	for _, instance := range instances {
		dockLogger.Infof("stopping dock container for instance: %s", instance.Id())
		if err := broker.manager.StopContainer(instance); err != nil {
			dockLogger.Errorf("container did not stop: %v", err)
			return err
		}
	}
	return nil
}

// AllInstances only returns running containers.
func (broker *dockBroker) AllInstances() (result []instance.Instance, err error) {
	return broker.manager.ListContainers()
}
