// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hive_test

import (
	stdtesting "testing"

	gc "launchpad.net/gocheck"

	"launchpad.net/juju-core/environs"
	//"launchpad.net/juju-core/provider"
	"github.com/Gusabi/judo/provider"
	"github.com/Gusabi/judo/provider/hive"
	"launchpad.net/juju-core/testing"
)

func TestLocal(t *stdtesting.T) {
	gc.TestingT(t)
}

type localSuite struct {
	testing.LoggingSuite
}

var _ = gc.Suite(&localSuite{})

func (*localSuite) TestProviderRegistered(c *gc.C) {
	provider, error := environs.Provider(provider.Hive)
	c.Assert(error, gc.IsNil)
	c.Assert(provider, gc.DeepEquals, &hive.Provider)
}
