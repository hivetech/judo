// Copyright 2013 Xavier Bruhiere
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

// Package to test and learn about docker api and juju integration
//
// (https://github.com/dotcloud/docker).
//
package main

import (
	"fmt"
    "os"
    "strings"
    "time"
    "github.com/dotcloud/docker"
    "github.com/dotcloud/docker/utils"
)

func StartContainer(machineId, series string) (string, error) {
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
    time.Sleep(time.Second * 10)
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%s", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        fmt.Println("** Error creating server")
        os.Exit(-1)
    }

    get_size := false
    all := true
    limit := -1  // i.e. no limit, dude
    containers := srv.Containers(all, get_size, limit, "", "")
    for _, box := range containers {
        //TODO Add time criteria
        fmt.Printf("%s vs %s\n", box.Image, series)
        if box.Image == series {
            fmt.Printf("Found container: %s\n", box.ID)
            return box.ID, nil
        }
    }
    return "", fmt.Errorf("Container not found, creation might failed")
}

func StopContainer(id string) error {
    fmt.Printf("Stoping container %s\n", id)
    if err := Execute([]string{"stop", id}); err != nil {
        return fmt.Errorf("Stop container %s\n", err)
    }
    time.Sleep(time.Second * 5)
    if err := Execute([]string{"rm", id}); err != nil {
        return fmt.Errorf("Destroy container: %s\n", err)
    }
    return nil
}

func ListContainers() ([]string, error) {
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%s", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        fmt.Println("** Error creating server")
        os.Exit(-1)
    }

    result := []string{}
    //Note: Returned variable are instanciated in the signature above and automatically returned
    get_size := true
    all := false   // We're only searching for running containers
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
        fmt.Printf("Checking container %s\n", container.ID)
        fmt.Printf("Created: %s, ", time.Unix(container.Created, 0))
        t := time.Now()
        fmt.Printf("%d seconds from now\n\n", t.Unix() - container.Created)
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

func main() {
    fmt.Printf("Using docker version %s (api %.2f)\n", docker.VERSION, docker.APIVERSION)

    // Start a container
    id, _ := StartContainer("machine-1", "base:latest")

    if id != "" {
        fmt.Printf("Container created has id %s\n", id)
        // List() test
        boxes, _ := ListContainers()
        for i, box := range boxes {
            fmt.Printf("Appends lxc instance: %s\n", box[i])
        }

        // Stop the same container
        if err := StopContainer(id); err != nil {
            fmt.Errorf("Stop container: %s", err)
        }
    } else {
        fmt.Println("No container id was fetched back")
    }
}
