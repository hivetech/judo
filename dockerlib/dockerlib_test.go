package dockerlib_test

import (
    "fmt"
    "time"
    "testing"
    "github.com/judo/dockerlib"
    "github.com/dotcloud/docker"
    "github.com/dotcloud/docker/utils"
)

func TestVersion(t *testing.T) {
    t.Log("Using docker version %s (api %.2f)\n", docker.VERSION, docker.APIVERSION)
    if docker.APIVERSION > 1.0 {
        t.Log("OK Docker api version: %.2f", docker.APIVERSION)
    } else {
        t.Errorf("Invalid Docker api version: %.2f", docker.APIVERSION)
        t.Fail()
    }
}

func TestServerConnection(t *testing.T) {
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    _, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        t.Errorf("Create server")
        t.Fail()
    } else {
        t.Log("Server successfully connected")
    }
}

func TestListContainers(t *testing.T) {
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        t.Fatal("Create server")
    } else {
        t.Log("Server successfully connected")
    }

    boxes, err := dockerlib.ListContainers(srv)
    if err != nil {
        t.Fatal("ListContainer failed")
    }
    for i, box := range boxes {
        t.Log("Appends lxc instance: %s\n", box[i])
    }
}

func testStartContainer(t *testing.T) {
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        t.Fatal("Create server")
    } else {
        t.Log("Server successfully connected")
    }

    // Note: tag is necessary
    id, err := dockerlib.StartContainer(srv, "machine-1", "ubuntu:12.04")
    if err != nil {
        t.Errorf("dockerlib.StartContainer: %s", err)
        t.Fail()
    } else if id == "" {
        t.Errorf("No id fetched after container creation")
        t.Fail()
    } else {
        t.Log("OK Container createdi with id %s", id)
    }
}

func TestStopContainer(t *testing.T) {
    flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        t.Fatal("Create server")
    } else {
        t.Log("Server successfully connected")
    }

    id, _ := dockerlib.StartContainer(srv, "machine-1", "ubuntu:12.04")
    // Give the container time to boot
    time.Sleep(2 * time.Second)
    // Stop the same container
    //FIXME No error raised even if docker server fails
    if err := dockerlib.StopContainer(id); err != nil {
        t.Errorf("Stop container: %s", err)
        t.Fail()
    }
}
