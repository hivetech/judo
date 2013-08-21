#!/bin/bash


#TODO Remove doublons, not all stopped
function remove_containers() {
    if [ -n "$1" ]; then
        echo "Removing container $1"
        docker rm $1
    else
        echo "Removing exited container"
        docker ps -a | grep 'Exit' |  awk '{print $1}' | xargs docker rm
    fi
}

function remove_images() {
    if [ -n "$1" ]; then
        echo "Removing image $1"
        docker rmi $1
    else
        echo "Removing unused images"
        docker images | grep '<none>' |  awk '{print $3}'  | xargs docker rmi
    fi
}

remove_containers
remove_images
