// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package provisioner_test

import (
	//"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//"time"

	gc "launchpad.net/gocheck"

	"launchpad.net/juju-core/agent/tools"
	"launchpad.net/juju-core/constraints"
    "github.com/Gusabi/judo/container/dock"
	//"launchpad.net/juju-core/container/lxc/mock"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/instance"
	//"launchpad.net/juju-core/juju/osenv"
	jujutesting "launchpad.net/juju-core/juju/testing"
	"launchpad.net/juju-core/state"
	coretesting "launchpad.net/juju-core/testing"
	jc "launchpad.net/juju-core/testing/checkers"
	"launchpad.net/juju-core/version"
	//"launchpad.net/juju-core/worker/provisioner"
	"github.com/Gusabi/judo/worker/provisioner"
)

type dockSuite struct {
	coretesting.LoggingSuite
	dock.TestSuite
	//events chan mock.Event
}

type dockBrokerSuite struct {
	dockSuite
	broker provisioner.Broker
}

var _ = gc.Suite(&dockBrokerSuite{})

func (s *dockSuite) SetUpSuite(c *gc.C) {
	s.LoggingSuite.SetUpSuite(c)
	s.TestSuite.SetUpSuite(c)
}

func (s *dockSuite) TearDownSuite(c *gc.C) {
	s.TestSuite.TearDownSuite(c)
	s.LoggingSuite.TearDownSuite(c)
}

func (s *dockSuite) SetUpTest(c *gc.C) {
	s.LoggingSuite.SetUpTest(c)
	s.TestSuite.SetUpTest(c)
	//s.events = make(chan mock.Event)
	//go func() {
		//for event := range s.events {
			//c.Output(3, fmt.Sprintf("dock event: <%s, %s>", event.Action, event.InstanceId))
		//}
	//}()
	//s.TestSuite.Factory.AddListener(s.events)
}

func (s *dockSuite) TearDownTest(c *gc.C) {
	//close(s.events)
	s.TestSuite.TearDownTest(c)
	s.LoggingSuite.TearDownTest(c)
}

func (s *dockBrokerSuite) SetUpTest(c *gc.C) {
	s.dockSuite.SetUpTest(c)
	tools := &tools.Tools{
		Version: version.MustParseBinary("2.3.4-foo-bar"),
		URL:     "http://tools.testing.invalid/2.3.4-foo-bar.tgz",
	}
	s.broker = provisioner.NewDockBroker(coretesting.EnvironConfig(c), tools)
}

func (s *dockBrokerSuite) startInstance(c *gc.C, machineId string) instance.Instance {
	stateInfo := jujutesting.FakeStateInfo(machineId)
	apiInfo := jujutesting.FakeAPIInfo(machineId)

    //series := "ubuntu:12.04"
    series := "base:latest"
	nonce := "fake-nonce"
	cons := constraints.Value{}
	dock, _, err := s.broker.StartInstance(machineId, nonce, series, cons, stateInfo, apiInfo)
	c.Assert(err, gc.IsNil)
	return dock
}

func (s *dockBrokerSuite) TestStartInstance(c *gc.C) {
	machineId := "1/dock/0"
	dock := s.startInstance(c, machineId)
	c.Assert(dock.Id(), gc.Equals, instance.Id("juju-machine-1-dock-0"))
	c.Assert(s.dockContainerDir(dock), jc.IsDirectory)
	s.assertInstances(c, dock)
	// Uses default network config
    // Docker has its own for now
    /*
	 *dockConfContents, err := ioutil.ReadFile(filepath.Join(s.ContainerDir, string(dock.Id()), "lxc.conf"))
	 *c.Assert(err, gc.IsNil)
	 *c.Assert(string(dockConfContents), jc.Contains, "lxc.network.type = veth")
	 *c.Assert(string(dockConfContents), jc.Contains, "lxc.network.link = docker0")
     */
}

// Currently we do not support custom lxc.conf
/*
 *func (s *dockBrokerSuite) TestStartInstanceWithBridgeEnviron(c *gc.C) {
 *    defer coretesting.PatchEnvironment(osenv.JujuLxcBridge, "br0")()
 *    machineId := "1/dock/0"
 *    dock := s.startInstance(c, machineId)
 *    c.Assert(dock.Id(), gc.Equals, instance.Id("juju-machine-1-dock-0"))
 *    c.Assert(s.dockContainerDir(dock), jc.IsDirectory)
 *    s.assertInstances(c, dock)
 *    // Uses default network config
 *    //dockConfContents, err := ioutil.ReadFile(filepath.Join(s.ContainerDir, string(dock.Id()), "lxc.conf"))
 *    //c.Assert(err, gc.IsNil)
 *    //c.Assert(string(dockConfContents), jc.Contains, "lxc.network.type = veth")
 *    //c.Assert(string(dockConfContents), jc.Contains, "lxc.network.link = br0")
 *}
 */

func (s *dockBrokerSuite) TestStopInstance(c *gc.C) {
    //FIXME Failed to umount fs
    dock0 := s.startInstance(c, "1/dock/0")
    dock1 := s.startInstance(c, "1/dock/1")
    dock2 := s.startInstance(c, "1/dock/2")

    err := s.broker.StopInstances([]instance.Instance{dock0})
    c.Assert(err, gc.IsNil)
    s.assertInstances(c, dock1, dock2)
    c.Assert(s.dockContainerDir(dock0), jc.DoesNotExist)
    c.Assert(s.dockRemovedContainerDir(dock0), jc.IsDirectory)

    err = s.broker.StopInstances([]instance.Instance{dock1, dock2})
    c.Assert(err, gc.IsNil)
    s.assertInstances(c)
}

func (s *dockBrokerSuite) TestAllInstances(c *gc.C) {
    dock0 := s.startInstance(c, "1/dock/0")
    dock1 := s.startInstance(c, "1/dock/1")
    s.assertInstances(c, dock0, dock1)

    err := s.broker.StopInstances([]instance.Instance{dock1})
    c.Assert(err, gc.IsNil)
    dock2 := s.startInstance(c, "1/dock/2")
    s.assertInstances(c, dock0, dock2)
    //mais pourquoi ici putain ya pas un stop après le start du 2 bordel????
    err = s.broker.StopInstances([]instance.Instance{dock2})
    c.Assert(err, gc.IsNil)
}

func (s *dockBrokerSuite) assertInstances(c *gc.C, inst ...instance.Instance) {
	results, err := s.broker.AllInstances()
	c.Assert(err, gc.IsNil)
	coretesting.MatchInstances(c, results, inst...)
}

func (s *dockBrokerSuite) dockContainerDir(inst instance.Instance) string {
	return filepath.Join(s.ContainerDir, string(inst.Id()))
}

func (s *dockBrokerSuite) dockRemovedContainerDir(inst instance.Instance) string {
	return filepath.Join(s.RemovedDir, string(inst.Id()))
}

type dockProvisionerSuite struct {
	CommonProvisionerSuite
	dockSuite
	machineId string
	//events    chan mock.Event
}

var _ = gc.Suite(&dockProvisionerSuite{})

func (s *dockProvisionerSuite) SetUpSuite(c *gc.C) {
	s.CommonProvisionerSuite.SetUpSuite(c)
	s.dockSuite.SetUpSuite(c)
}

func (s *dockProvisionerSuite) TearDownSuite(c *gc.C) {
	s.dockSuite.TearDownSuite(c)
	s.CommonProvisionerSuite.TearDownSuite(c)
}

func (s *dockProvisionerSuite) SetUpTest(c *gc.C) {
	s.CommonProvisionerSuite.SetUpTest(c)
	s.dockSuite.SetUpTest(c)
	// Write the tools file.
	toolsDir := tools.SharedToolsDir(s.DataDir(), version.Current)
	c.Assert(os.MkdirAll(toolsDir, 0755), gc.IsNil)
	urlPath := filepath.Join(toolsDir, "downloaded-url.txt")
	err := ioutil.WriteFile(urlPath, []byte("http://testing.invalid/tools"), 0644)
	c.Assert(err, gc.IsNil)

	// The dock provisioner actually needs the machine it is being created on
	// to be in state, in order to get the watcher.
	m, err := s.State.AddMachine(config.DefaultSeries, state.JobHostUnits)
	c.Assert(err, gc.IsNil)
	s.machineId = m.Id()

	//s.events = make(chan mock.Event, 25)
	//s.Factory.AddListener(s.events)
}

/*
 *func (s *dockProvisionerSuite) expectStarted(c *gc.C, machine *state.Machine) string {
 *    event := <-s.events
 *    c.Assert(event.Action, gc.Equals, mock.Started)
 *    err := machine.Refresh()
 *    c.Assert(err, gc.IsNil)
 *    s.waitInstanceId(c, machine, instance.Id(event.InstanceId))
 *    return event.InstanceId
 *}
 */

/*
 *func (s *dockProvisionerSuite) expectStopped(c *gc.C, instId string) {
 *    event := <-s.events
 *    c.Assert(event.Action, gc.Equals, mock.Stopped)
 *    c.Assert(event.InstanceId, gc.Equals, instId)
 *}
 */

/*
 *func (s *dockProvisionerSuite) expectNoEvents(c *gc.C) {
 *    select {
 *    case event := <-s.events:
 *        c.Fatalf("unexpected event %#v", event)
 *    case <-time.After(coretesting.ShortWait):
 *        return
 *    }
 *}
 */

func (s *dockProvisionerSuite) TearDownTest(c *gc.C) {
	//close(s.events)
	s.dockSuite.TearDownTest(c)
	s.CommonProvisionerSuite.TearDownTest(c)
}

func (s *dockProvisionerSuite) newDockProvisioner() *provisioner.Provisioner {
	return provisioner.NewProvisioner(provisioner.DOCK, s.State, s.machineId, s.DataDir())
}

/*
 *func (s *dockProvisionerSuite) TestProvisionerStartStop(c *gc.C) {
 *    p := s.newDockProvisioner()
 *    c.Assert(p.Stop(), gc.IsNil)
 *}
 */

/*
 *func (s *dockProvisionerSuite) TestDoesNotStartEnvironMachines(c *gc.C) {
 *    p := s.newDockProvisioner()
 *    defer stop(c, p)
 *
 *    // Check that an instance is not provisioned when the machine is created.
 *    _, err := s.State.AddMachine(config.DefaultSeries, state.JobHostUnits)
 *    c.Assert(err, gc.IsNil)
 *
 *    s.expectNoEvents(c)
 *}
 */

func (s *dockProvisionerSuite) addContainer(c *gc.C) *state.Machine {
	params := state.AddMachineParams{
		ParentId:      s.machineId,
		ContainerType: instance.DOCK,
		Series:        config.DefaultSeries,
		Jobs:          []state.MachineJob{state.JobHostUnits},
	}
	container, err := s.State.AddMachineWithConstraints(&params)
	c.Assert(err, gc.IsNil)
	return container
}

/*
 *func (s *dockProvisionerSuite) TestContainerStartedAndStopped(c *gc.C) {
 *    p := s.newDockProvisioner()
 *    defer stop(c, p)
 *
 *    container := s.addContainer(c)
 *
 *    instId := s.expectStarted(c, container)
 *
 *    // ...and removed, along with the machine, when the machine is Dead.
 *    c.Assert(container.EnsureDead(), gc.IsNil)
 *    s.expectStopped(c, instId)
 *    s.waitRemoved(c, container)
 *}
 */
