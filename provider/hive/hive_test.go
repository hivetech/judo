// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hive_test

import (
	stdtesting "testing"

	gc "launchpad.net/gocheck"

	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/provider"
	"launchpad.net/juju-core/provider/hive"
	"launchpad.net/juju-core/testing"
)

func TestHive(t *stdtesting.T) {
	gc.TestingT(t)
}

type hiveSuite struct {
	testing.LoggingSuite
}

var _ = gc.Suite(&hiveSuite{})

func (*hiveSuite) TestProviderRegistered(c *gc.C) {
	provider, error := environs.Provider(provider.Hive)
	c.Assert(error, gc.IsNil)
	c.Assert(provider, gc.DeepEquals, &hive.Provider)
}
