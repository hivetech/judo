// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hive

import (
	"io/ioutil"
	"os"
	"path/filepath"

	gc "launchpad.net/gocheck"

	"launchpad.net/juju-core/testing"
)

type prereqsSuite struct {
	testing.LoggingSuite
	tmpdir  string
	oldpath string
}

var _ = gc.Suite(&prereqsSuite{})

const lsbrelease = `#!/bin/sh
echo $JUJUTEST_LSB_RELEASE_ID
`

func init() {
	// Set the paths to mongod and lxc-ls to
	// something we know exists. This allows
	// all of the non-prereqs tests to pass
	// even when mongodb and lxc-ls can't be
	// found.
	mongodPath = "/bin/true"
	dockerPath = "/bin/true"
}

func (s *prereqsSuite) SetUpTest(c *gc.C) {
	s.LoggingSuite.SetUpTest(c)
	s.tmpdir = c.MkDir()
	s.oldpath = os.Getenv("PATH")
	mongodPath = filepath.Join(s.tmpdir, "mongod")
	dockerPath = filepath.Join(s.tmpdir, "docker")
	os.Setenv("PATH", s.tmpdir)
	os.Setenv("JUJUTEST_LSB_RELEASE_ID", "Ubuntu")
	err := ioutil.WriteFile(filepath.Join(s.tmpdir, "lsb_release"), []byte(lsbrelease), 0777)
	c.Assert(err, gc.IsNil)
}

func (s *prereqsSuite) TearDownTest(c *gc.C) {
	os.Setenv("PATH", s.oldpath)
	mongodPath = "/bin/true"
	dockerPath = "/bin/true"
	s.LoggingSuite.TearDownTest(c)
}

func (*prereqsSuite) TestSupportedOS(c *gc.C) {
	defer func(old string) {
		goos = old
	}(goos)
	goos = "windows"
	err := VerifyPrerequisites()
	c.Assert(err, gc.ErrorMatches, "Unsupported operating system: windows(.|\n)*")
}

func (s *prereqsSuite) TestMongoPrereq(c *gc.C) {
	err := VerifyPrerequisites()
	c.Assert(err, gc.ErrorMatches, "(.|\n)*MongoDB server must be installed(.|\n)*")
	c.Assert(err, gc.ErrorMatches, "(.|\n)*apt-get install mongodb-server(.|\n)*")

	os.Setenv("JUJUTEST_LSB_RELEASE_ID", "NotUbuntu")
	err = VerifyPrerequisites()
	c.Assert(err, gc.ErrorMatches, "(.|\n)*MongoDB server must be installed(.|\n)*")
	c.Assert(err, gc.Not(gc.ErrorMatches), "(.|\n)*apt-get install(.|\n)*")

	err = ioutil.WriteFile(mongodPath, nil, 0777)
	c.Assert(err, gc.IsNil)
	err = ioutil.WriteFile(filepath.Join(s.tmpdir, "docker"), nil, 0777)
	c.Assert(err, gc.IsNil)
	err = VerifyPrerequisites()
	c.Assert(err, gc.IsNil)
}

func (s *prereqsSuite) TestDockerPrereq(c *gc.C) {
	err := ioutil.WriteFile(mongodPath, nil, 0777)
	c.Assert(err, gc.IsNil)

	err = VerifyPrerequisites()
	//c.Assert(err, gc.ErrorMatches, "(.|\n)*Linux Containers \\(LXC\\) userspace tools must be\ninstalled(.|\n)*")
	c.Assert(err, gc.ErrorMatches, "(.|\n)*Docker, the LXC containers manager, must be\ninstalled(.|\n)*")
    c.Assert(err, gc.ErrorMatches, "(.|\n)*curl https://get.docker.io | sudo sh -x(.|\n)*")

	os.Setenv("JUJUTEST_LSB_RELEASE_ID", "NotUbuntu")
	err = VerifyPrerequisites()
	c.Assert(err, gc.ErrorMatches,  "(.|\n)*Docker, the LXC containers manager, must be\ninstalled(.|\n)*")
	c.Assert(err, gc.Not(gc.ErrorMatches), "(.|\n)*curl(.|\n)*")

	err = ioutil.WriteFile(dockerPath, nil, 0777)
	c.Assert(err, gc.IsNil)
	err = VerifyPrerequisites()
	c.Assert(err, gc.IsNil)
}
