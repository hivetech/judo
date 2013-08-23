// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package hive

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"launchpad.net/juju-core/version"
)

var notLinuxError = errors.New("The local provider is currently only available for Linux")

const installMongodUbuntu = "MongoDB server must be installed to enable the local provider:"
const aptAddRepositoryJujuStable = `
    sudo apt-add-repository ppa:juju/stable   # required for MongoDB SSL support
    sudo apt-get update`
const aptGetInstallMongodbServer = `
    sudo apt-get install mongodb-server`

const installMongodGeneric = `
MongoDB server must be installed to enable the local provider.
Please consult your operating system distribution's documentation
for instructions on installing the MongoDB server. Juju requires
a MongoDB server built with SSL support.
`

const installDockerUbuntu = `
Docker, the LXC containers manager, must be
installed to enable the hive provider:
  
    curl https://get.docker.io | sudo sh -x
`

const installDockerGeneric = `
Docker, the LXC containers manager, must be
installed to enable the hive provider.
Please consult your operating system distribution's
documentation for instructions on installing Docker.`

const errUnsupportedOS = `Unsupported operating system: %s
The hive provider is currently only available for Linux`

// mongodPath is the path to "mongod", the MongoDB server.
// This is a variable only to support unit testing.
var mongodPath = "/usr/bin/mongod"

// lxclsPath is the path to "lxc-ls", an LXC userspace tool
// we check the presence of to determine whether the
// tools are installed. This is a variable only to support
// unit testing.
var dockerPath = "docker"

// The operating system the process is running in.
// This is a variable only to support unit testing.
var goos = runtime.GOOS

// VerifyPrerequisites verifies the prerequisites of
// the local machine (machine 0) for running the local
// provider.
func VerifyPrerequisites() error {
	if goos != "linux" {
		return fmt.Errorf(errUnsupportedOS, goos)
	}
	if err := verifyMongod(); err != nil {
		return err
	}
	return verifyDocker()
}

func verifyMongod() error {
	if _, err := os.Stat(mongodPath); err != nil {
		if os.IsNotExist(err) {
			return wrapMongodNotExist(err)
		} else {
			return err
		}
	}
	// TODO(axw) verify version/SSL capabilities
	return nil
}

func verifyDocker() error {
	_, err := exec.LookPath(dockerPath)
	if err != nil {
		return wrapDockerNotFound(err)
	}
	return nil
}

func isUbuntu() bool {
	out, err := exec.Command("lsb_release", "-i", "-s").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "Ubuntu"
}

func wrapMongodNotExist(err error) error {
	if isUbuntu() {
		series := version.Current.Series
		args := []interface{}{err, installMongodUbuntu}
		format := "%v\n%s\n%s"
		if series == "precise" || series == "quantal" {
			format += "%s"
			args = append(args, aptAddRepositoryJujuStable)
		}
		args = append(args, aptGetInstallMongodbServer)
		return fmt.Errorf(format, args...)
	}
	return fmt.Errorf("%v\n%s", err, installMongodGeneric)
}

func wrapDockerNotFound(err error) error {
	if isUbuntu() {
		return fmt.Errorf("%v\n%s", err, installDockerUbuntu)
	}
	return fmt.Errorf("%v\n%s", err, installDockerGeneric)
}
