// Copyright 2013 Xavier Bruhiere
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

// Package to test and learn about docker api and juju integration
//
// (https://github.com/dotcloud/docker).
//
package dockerlib

import (
	"fmt"
    "os"
    "strings"
    "github.com/dotcloud/docker"
    "github.com/dotcloud/docker/utils"
)

func FindContainer(srv *docker.Server, command, image string) docker.APIContainers {
    fmt.Printf("Searching for container running %s in %s\n", command, image)

    get_size := false
    all := false
    limit := -1  // i.e. no limit, dude
    containers := srv.Containers(all, get_size, limit, "", "")
    for _, box := range containers {
        if box.Image == image && box.Command == command {
            return box
        }
    }
    return docker.APIContainers{ID: ""}
}

func StartContainer(srv *docker.Server, machineId, series string) (string, error) {
    //command := "/bin/bash -c apt-get install cloud-init && cloud-init init"
    command := "while true; do sleep 300; done"
    // Note -h option is left to default, i.e. its id, as the container id is its name
    //FIXME -u ubuntu makes container to exit immediatly with code 1
    //args := []string{"run", "-d", "-u", "ubuntu", series, command}
    args := []string{"run", "-d", series, "/bin/bash", "-c", command}
    if err := Execute(args); err != nil {
        return "", fmt.Errorf("Stop container %s\n", err)
    }

    // Fetching back the id
    //image := series + ":12.04"
    image := series
    container := FindContainer(srv, "/bin/bash -c " + command, image)
    if container.ID == "" {
        return "", fmt.Errorf("Container not found, creation might failed")
    }
    name := container.ID  // Again, short id ?
    return name, nil
}

func StopContainer(id string) error {
    if err := Execute([]string{"stop", id}); err != nil {
        return fmt.Errorf("Stop container %s\n", err)
    }
    if err := Execute([]string{"rm", id}); err != nil {
        return fmt.Errorf("Destroy container: %s\n", err)
    }
    return nil
}

func ListContainers(srv *docker.Server) ([]string, error) {
    result := []string{}
    //Note: Returned variable are instanciated in the signature above and automatically returned
    get_size := true
    all := true  // We're only searching for running containers
    limit := -1   // i.e. no limit, dude
    containers := srv.Containers(all, get_size, limit, "", "")

    // --- manager trick
    prefix := ""
    // --- -------------
	managerPrefix := ""
	if prefix != "" {
		managerPrefix = fmt.Sprintf("%s-", prefix)
        return result, fmt.Errorf("managerPrefix is not supported with Dokcer (%s)", prefix)
	}

	for _, container := range containers {
		// Filter out those not starting with our name.
		name := container.ID  // Or TruncateID ?
		if !strings.HasPrefix(name, managerPrefix) {
			continue
		}
		if !strings.Contains(container.Status, "Exit") {
			result = append(result, name)
		}
	}
    return result, nil
}

func Execute(args []string) error {
    //protoAddrParts := strings.SplitN(flHosts[0], "://", 2)
    if err:= docker.ParseCommands("unix", docker.DEFAULTUNIXSOCKET, args...); err != nil {
        return fmt.Errorf("** Error docker.ParseCommands: %s\n", err)
    }
    return nil
}

func dockerlib_mock() {
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%s", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        fmt.Println("** Error creating server")
        os.Exit(-1)
    }

    fmt.Printf("Using docker version %s (api %.2f)\n", docker.VERSION, docker.APIVERSION)

    // Start a container
    id, _ := StartContainer(srv, "machine-1", "ubuntu")
    fmt.Printf("Container created has id %s", id)

    // List() test
    boxes, _ := ListContainers(srv)
    for i, box := range boxes {
        fmt.Printf("Appends lxc instance: %s\n", box[i])
    }

    // Stop the same container
    err = StopContainer(id)
    if err != nil {
        fmt.Errorf("Stop container: %s", err)
    }
}
