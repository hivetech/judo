// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

// golxc - Go package to interact with Linux Containers (LXC).
//
// https://launchpad.net/golxc/
//
package dockerinterface

import (
    "github.com/dotcloud/docker"
)

// ContainerFactory represents the methods used to create Containers.
type ContainerFactory interface {
	// New returns a container instance which can then be used for operations
	// like Create(), Start(), Stop() or Destroy().
	New(string) docker.Container

	// List returns all the existing containers on the system.
	List() ([]docker.Container, error)
}
