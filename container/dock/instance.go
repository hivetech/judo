// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package dock

import (
	"fmt"

	"launchpad.net/golxc"

	"launchpad.net/juju-core/instance"
)

type dockInstance struct {
	golxc.Container
	id string
}

var _ instance.Instance = (*dockInstance)(nil)

// Id implements instance.Instance.Id.
func (lxc *dockInstance) Id() instance.Id {
	return instance.Id(lxc.id)
}

// Status implements instance.Instance.Status.
func (lxc *dockInstance) Status() string {
	// On error, the state will be "unknown".
	state, _, _ := lxc.Info()
	return string(state)
}

func (lxc *dockInstance) Addresses() ([]instance.Address, error) {
	logger.Errorf("dockInstance.Addresses not implemented")
	return nil, nil
}

// DNSName implements instance.Instance.DNSName.
func (lxc *dockInstance) DNSName() (string, error) {
	return "", instance.ErrNoDNSName
}

// WaitDNSName implements instance.Instance.WaitDNSName.
func (lxc *dockInstance) WaitDNSName() (string, error) {
	return "", instance.ErrNoDNSName
}

// OpenPorts implements instance.Instance.OpenPorts.
func (lxc *dockInstance) OpenPorts(machineId string, ports []instance.Port) error {
	return fmt.Errorf("not implemented")
}

// ClosePorts implements instance.Instance.ClosePorts.
func (lxc *dockInstance) ClosePorts(machineId string, ports []instance.Port) error {
	return fmt.Errorf("not implemented")
}

// Ports implements instance.Instance.Ports.
func (lxc *dockInstance) Ports(machineId string) ([]instance.Port, error) {
	return nil, fmt.Errorf("not implemented")
}

// Add a string representation of the id.
func (lxc *dockInstance) String() string {
	return fmt.Sprintf("lxc:%s", lxc.id)
}
