package main

import (
    "fmt"
    "github.com/garyburd/redigo/redis"
)

func main() {
    c, err := redis.Dial("tcp", ":6379")
    if err != nil {
        fmt.Errorf("** Error connecting to redis server: %v", err)
    }
    defer c.Close()

    response, _ := redis.Int64(c.Do("DECR", "default_ssh_port"))
    fmt.Printf("Port: %d\n", response)

    //c.Send("GET", "hello")
    //c.Flush()
    //v, err := redis.String(c.Receive())
    //fmt.Printf("%s", v)
    //for _, val := range v {
        //fmt.Printf(val)
    //}
}
