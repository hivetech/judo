package godocker_test

import (
	. "launchpad.net/gocheck"
    "os"
	"os/user"
	"testing"

    "github.com/judo/godocker"
)

// Hook up gocheck into  "go test" runner
func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct{}

var _ = Suite(&ConfigSuite{})

type LXCSuite struct {
	factory godocker.ContainerFactory
}

var _ = Suite(&LXCSuite{godocker.Factory()})

func (s *LXCSuite) SetUpSuite(c *C) {
	u, err := user.Current()
	c.Assert(err, IsNil)
	if u.Uid != "0" {
		// Has to be run as root!
		c.Skip("tests must run as root")
	}
}

func (s *LXCSuite) TestCreateDestroy(c *C) {
    // Test clean creation and destroying of a container.
    lc := s.factory.New("godocker")
    c.Assert(lc.IsConstructed(), Equals, false)
    home := godocker.ContainerHome(lc)
    _, err := os.Stat(home)
    c.Assert(err, ErrorMatches, "stat .*: no such file or directory")

    err = lc.Create("", "ubuntu:12.04")
    c.Assert(err, IsNil)
    c.Assert(lc.IsConstructed(), Equals, true)
    defer func() {
        err = lc.Destroy()
        c.Assert(err, IsNil)
        _, err = os.Stat(home)
        c.Assert(err, ErrorMatches, "stat .*: no such file or directory")
    }()
    fi, err := os.Stat(godocker.ContainerHome(lc))
    c.Assert(err, IsNil)
    c.Assert(fi.IsDir(), Equals, true)
}

func (s *LXCSuite) TestCreateTwice(c *C) {
    // Test that a container cannot be created twice.
    lc1 := s.factory.New("godocker")
    c.Assert(lc1.IsConstructed(), Equals, false)
    err := lc1.Create("", "ubuntu:12.04")
    c.Assert(err, IsNil)
    c.Assert(lc1.IsConstructed(), Equals, true)
    defer func() {
        c.Assert(lc1.Destroy(), IsNil)
    }()
    lc2 := s.factory.New("godocker")
    err = lc2.Create("", "ubuntu:12.04")
    c.Assert(err, ErrorMatches, "container .* is already created")
}


func (s *LXCSuite) TestCreateIllegalTemplate(c *C) {
    // Test that a container creation fails correctly in
    // case of an illegal template.
    lc := s.factory.New("godocker")
    c.Assert(lc.IsConstructed(), Equals, false)
    err := lc.Create("", "name-of-a-not-existing-template-for-godocker")
    c.Assert(err, ErrorMatches, `No base image found for container .*`)
    //c.Assert(err, ErrorMatches, `error executing "lxc-create": No config file specified, .*`)
    c.Assert(lc.IsConstructed(), Equals, false)
}

/*
 *func (s *LXCSuite) TestDestroyNotCreated(c *C) {
 *    // Test that a non-existing container can't be destroyed.
 *    lc := s.factory.New("godocker")
 *    c.Assert(lc.IsConstructed(), Equals, false)
 *    err := lc.Destroy()
 *    c.Assert(err, ErrorMatches, "container .* is not yet created")
 *}
 */

/*
 *func contains(lcs []godocker.Container, lc godocker.Container) bool {
 *    for _, clc := range lcs {
 *        if clc.Name() == lc.Name() {
 *            return true
 *        }
 *    }
 *    return false
 *}
 */

/*
 *func (s *LXCSuite) TestList(c *C) {
 *    // Test the listing of created containers.
 *    lcs, err := s.factory.List()
 *    oldLen := len(lcs)
 *    c.Assert(err, IsNil)
 *    c.Assert(oldLen >= 0, Equals, true)
 *    lc := s.factory.New("godocker")
 *    c.Assert(lc.IsConstructed(), Equals, false)
 *    c.Assert(lc.Create("", "ubuntu"), IsNil)
 *    c.Assert(lc.IsConstructed(), Equals, true)
 *    defer func() {
 *        c.Assert(lc.Destroy(), IsNil)
 *    }()
 *    lcs, _ = s.factory.List()
 *    newLen := len(lcs)
 *    c.Assert(newLen == oldLen+1, Equals, true)
 *    c.Assert(contains(lcs, lc), Equals, true)
 *}
 */

/*
 *func (s *LXCSuite) TestClone(c *C) {
 *    // Test the cloning of an existing container.
 *    lc1 := s.factory.New("godocker")
 *    c.Assert(lc1.IsConstructed(), Equals, false)
 *    c.Assert(lc1.Create("", "ubuntu"), IsNil)
 *    c.Assert(lc1.IsConstructed(), Equals, true)
 *    defer func() {
 *        c.Assert(lc1.Destroy(), IsNil)
 *    }()
 *    lcs, _ := s.factory.List()
 *    oldLen := len(lcs)
 *    lc2, err := lc1.Clone("godockerclone")
 *    c.Assert(err, IsNil)
 *    c.Assert(lc2.IsConstructed(), Equals, true)
 *    defer func() {
 *        c.Assert(lc2.Destroy(), IsNil)
 *    }()
 *    lcs, _ = s.factory.List()
 *    newLen := len(lcs)
 *    c.Assert(newLen == oldLen+1, Equals, true)
 *    c.Assert(contains(lcs, lc1), Equals, true)
 *    c.Assert(contains(lcs, lc2), Equals, true)
 *}
 */

/*
 *func (s *LXCSuite) TestStartStop(c *C) {
 *    // Test starting and stopping a container.
 *    lc := s.factory.New("godocker")
 *    c.Assert(lc.IsConstructed(), Equals, false)
 *    c.Assert(lc.Create("", "ubuntu"), IsNil)
 *    defer func() {
 *        c.Assert(lc.Destroy(), IsNil)
 *    }()
 *    c.Assert(lc.Start("", ""), IsNil)
 *    c.Assert(lc.IsRunning(), Equals, true)
 *    c.Assert(lc.Stop(), IsNil)
 *    c.Assert(lc.IsRunning(), Equals, false)
 *}
 */

/*
 *func (s *LXCSuite) TestStartNotCreated(c *C) {
 *    // Test that a non-existing container can't be started.
 *    lc := s.factory.New("godocker")
 *    c.Assert(lc.IsConstructed(), Equals, false)
 *    c.Assert(lc.Start("", ""), ErrorMatches, "container .* is not yet created")
 *}
 */

/*
 *func (s *LXCSuite) TestStopNotRunning(c *C) {
 *    // Test that a not running container can't be stopped.
 *    lc := s.factory.New("godocker")
 *    c.Assert(lc.IsConstructed(), Equals, false)
 *    c.Assert(lc.Create("", "ubuntu"), IsNil)
 *    defer func() {
 *        c.Assert(lc.Destroy(), IsNil)
 *    }()
 *    c.Assert(lc.Stop(), IsNil)
 *}
 */

/*
 *func (s *LXCSuite) TestWait(c *C) {
 *    // Test waiting for one of a number of states of a container.
 *    // ATTN: Using a not reached state blocks the test until timeout!
 *    lc := s.factory.New("godocker")
 *    c.Assert(lc.IsConstructed(), Equals, false)
 *    c.Assert(lc.Wait(), ErrorMatches, "no states specified")
 *    c.Assert(lc.Wait(godocker.StateStopped), IsNil)
 *    c.Assert(lc.Wait(godocker.StateStopped, godocker.StateRunning), IsNil)
 *    c.Assert(lc.Create("", "ubuntu"), IsNil)
 *    defer func() {
 *        c.Assert(lc.Destroy(), IsNil)
 *    }()
 *    go func() {
 *        c.Assert(lc.Start("", ""), IsNil)
 *    }()
 *    c.Assert(lc.Wait(godocker.StateRunning), IsNil)
 *}
 */
