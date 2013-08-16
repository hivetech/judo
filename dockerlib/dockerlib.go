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
    //"github.com/dotcloud/docker/utils"
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

func showLocalImages(srv *docker.Server, pattern string) error {
    if images, err := srv.Images(true, pattern); err != nil {
        fmt.Errorf("Error, handles it dude\n")
    } else {
        fmt.Printf("Found %d images\n", len(images))
        for _, image := range images {
            fmt.Println(image.Repository)
            fmt.Println(image.Tag)
            fmt.Println(image.ID)
            fmt.Println(image.VirtualSize)
            fmt.Println(image.Created)
            fmt.Println("=============================")
        }
    }
    return nil
}

func List(srv *docker.Server) {
    //flHost := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    //flHost = utils.ParseHost(docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT, flHost)
    runtime, _ := docker.NewRuntime("/var/lib/docker/graph", false, []string{"127.0.0.1"})
    //containers, _ := srv.Containers(true, false)

    containers := runtime.List()
    fmt.Printf("Found %d containers\n", len(containers))

    //container := srv.runtime.Get("base:latest")
}

/*
 *func UseCmdApi() {
 *    prot := ""
 *    addr := "http://127.0.0.1"
 *    cli := NewDockerCli(os.Stdin, os.Stdout, os.Stderr, proto, addr)
 *}
 */

 func createContainer(srv *docker.Server) error {
    //config := &docker.Config{}
    template := "base"
    name := "godocker"
    args := []string{"-h", name, template, "/bin/bash", "-c", "ls"}
    //config, hostconfig, cmd, err := docker.ParseRun(args, nil)
    //config, _, _, _ := docker.ParseRun(args, nil)
    config, _, _, _ := docker.ParseRun(args, nil)

    if shortid, err := srv.ContainerCreate(config); err != nil {
        return fmt.Errorf("Could not create container %s", name)
    }
    return nil
 }

func main() {
    //flHost := docker.ListOpts{fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)}
    flHost := docker.ListOpts{fmt.Sprintf("http://%s:%s", docker.DEFAULTHTTPHOST, docker.DEFAULTHTTPPORT)}
    srv, err := docker.NewServer("/var/lib/docker/graph", false, true, flHost)
    if err != nil {
        fmt.Println("** Error creating server")
        os.Exit(-1)
    }
    api := srv.DockerVersion()
    fmt.Printf("Docker version: %s\n", api.Version)

    //showLocalImages(srv, "ubuntu")
    //showImages(srv, "base")
    //List(srv)

    if err := createContainer(srv); err != nil {
        fmt.Println("Fuck...")
    }
}
