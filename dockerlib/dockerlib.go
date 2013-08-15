// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

// godocker - Go package to interact with Linux Containers (LXC) through docker
//
// (https://github.com/dotcloud/docker).
//
package main

import (
	"fmt"
    "github.com/dotcloud/docker"
)

func main() {
    //srv := docker.Server{}
    //api := srv.DockerVersion()

    //srv, err := docker.server.NewServer()
    //runtime := mkRuntime()
    //srv := &Server{runtime: runtime}

    //FIXME Need sudo on my machine
    /*
     *if images, err := srv.ImagesSearch("ubuntu"); err != nil {
     *    fmt.Println("Error, handles it dude")
     *} else {
     *    for _, image := range images {
     *        fmt.Println(image)
     *    }
     *}
     */
    srv, _ := docker.NewServer("/var/lib/docker/graph", false, true, nil)
    //images, _ := srv.Images(true, template)
    images, _ := srv.ImagesSearch("ubuntu")
    fmt.Printf("Found %d images", len(images))

    //for img := range images {
        //fmt.Printf("Description: %s", img.Description)
        //fmt.Printf("Name: %s", img.Name)
    //}
}
