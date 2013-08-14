// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see COPYING and COPYING.LESSER file for details.

// godocker - Go package to interact with Linux Containers (LXC) through docker
//
// (https://github.com/dotcloud/docker).
//
package godocker

import (
	"fmt"
    "github.com/dotcloud/docker"
)

func GetVersion() string {
    srv := docker.Server{}

    //srv, err := docker.server.NewServer()
    //runtime := mkRuntime()
    //srv := &Server{runtime: runtime}

    api := srv.DockerVersion()

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

     return api.Version
}
