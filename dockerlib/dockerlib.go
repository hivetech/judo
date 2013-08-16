// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

// godocker - Go package to interact with Linux Containers (LXC) through docker
//
// (https://github.com/dotcloud/docker).
//
package main

import (
	"fmt"
    "os"
    "github.com/dotcloud/docker"
    "github.com/dotcloud/docker/utils"
    //"github.com/dotcloud/docker/auth"
    //"net/http"
)

func showImages(srv *docker.Server, pattern string) error {
    if images, err := srv.ImagesSearch(pattern); err != nil {
        fmt.Errorf("Error, handles it dude\n")
    } else {
        fmt.Printf("Found %d images\n", len(images))
        for _, image := range images {
            fmt.Println(image.Name)
            fmt.Println(image.Description)
            fmt.Println("=============================")
        }
    }
    return nil
}

 func createContainer(srv *docker.Server, base_image string) error {

    //var test_id string
    /*
     *if images, err := srv.Images(true, base_image); err != nil {
     *    fmt.Errorf("Error listing images: %s\n", err)
     *} else {
     *    if len(images) == 0 {
     *        return fmt.Errorf("No image found on this server")
     *    } else {
     *        fmt.Printf("Found %d images\n", len(images))
     *        for _, image := range images {
     *            fmt.Println(image.Repository)
     *            fmt.Println(image.Tag)
     *            fmt.Println(image.ID)
     *            //test_id = image.ID
     *            fmt.Println(image.VirtualSize)
     *            fmt.Println(image.Created)
     *        }
     *    }
     *}
     */


    // 1st try
    //box_id := "7b949cbc8c3dbd6d0c766e3dd114ea3f223250e3a8826375c39eb08e85cde1a1"
    //args := []string{"8dbd9e392a96", "/bin/bash", "-c", "\"while True; do sleep 10; done\""}
    //args := []string{"7b949cbc8c3dbd6d0c766e3dd114ea3f223250e3a8826375c39eb08e85cde1a1", "echo", "run", "-u", "ubntu", "-h", "godocker", "/bin/builder"}
    args := []string{"-u", "ubuntu", "-h", "godocker", base_image, "/bin/bash", "-c", "\"while true; do sleep 10; done\""}
    config, hostconfig, _, err := docker.ParseRun(args, nil)
    if err != nil {
        return fmt.Errorf("ParseRun failed: %s", err)
    }

    //utils.GetResolvConf().(string)
    //config.Dns = []string{"/etc/resolv.conf"}

    // Homemade
    /*
     *config := docker.Config{
     *    Hostname: "godocker",
     *    PortSpecs: []string{},
     *    User: "ubuntu",
     *    Tty: false,
     *    NetworkDisabled: false,
     *    OpenStdin: false,
     *    Memory: 0,
     *    CpuShares: 0,
     *    AttachStdin: false,
     *    AttachStderr: false,
     *    AttachStdout: false,
     *    Env: []string{},
     *    Cmd: []string{"/bin/bash", "-c", "ls"},
     *    Dns: nil,
     *    Image: template,
     *    Volumes: nil,
     *    VolumesFrom: "",
     *    Entrypoint: nil,
     *    Privileged: false,
     *}
     */

    shortid, err := srv.ContainerCreate(config)
    if err != nil {
        return fmt.Errorf("Could not create container %s\n", err)
    }
    fmt.Printf("New container created with Id: %s\n", shortid)
    if err := srv.ContainerStart(shortid, hostconfig); err != nil {
        return fmt.Errorf("Errorr starting container: %s", err)
    }
    fmt.Println("Container started successfully")

    //if _, err := srv.ContainerWait(shortid); err != nil {
        //return fmt.Errorf("Errorr waiting container: %s", err)
    //}
    //fmt.Println("We waited successfully")

    container, err := srv.ContainerInspect(shortid)
    if err != nil {
        return fmt.Errorf("Error inspecting container %s", shortid)
    }
    //newhostconfig, _ := container.ReadHostConfig()
    //fmt.Println(newhostconfig)
    //if err := container.Run(); err != nil {
        //return fmt.Errorf("Error container.Run: %s", err)
    //}

    //fmt.Printf("Long container id: %s\n", container.ID)

    containers := srv.Containers(true, false, -1, "", "")
    for _, box := range containers {
        if box.ID == container.ID {
            fmt.Printf("===============\n")
            fmt.Printf("Container id: %s\n", box.ID)
            fmt.Printf("Container cmd: %s\n", box.Command)
            fmt.Printf("Container image: %s\n", box.Image)
            fmt.Printf("Container status: %s\n", box.Status)

            //if err:= srv.ContainerStop(box.ID, 1); err != nil {
                //return fmt.Errorf("Error stoping container with id %s", box.ID)
            //}
            //fmt.Println("Container stopped successfully")

            //if err:= srv.ContainerDestroy(box.ID, true); err != nil {
                //return fmt.Errorf("Error destroying container with id %s", box.ID)
            //}
            //fmt.Println("Container destroyed successfully")

            //fmt.Printf("===============\n")
        }
    }

    return nil
 }

func main() {
    //flHosts := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%s", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    //flHosts := docker.ListOpts{fmt.Sprintf("tcp://%s:%s", docker.DEFAULTHTTPHOST, 4242)}
    //flHost := docker.ListOpts{fmt.Sprintf("tcp://:4243")}
    flHosts[0] = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHosts[0])
    srv, err := docker.NewServer("/var/lib/docker", false, false, flHosts)
    if err != nil {
        fmt.Println("** Error creating server")
        os.Exit(-1)
    }

    //runtime := mkRuntime()
    //defer nuke(runtime)

    //srv := &Server{runtime: runtime}

    //showLocalImages(srv, "ubuntu")
    //showImages(srv, "base")
    //List(srv)

    if err := createContainer(srv, "ubuntu"); err != nil {
        fmt.Println(err)
    }
}
