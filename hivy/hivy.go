package main

import (
    "os"
    "fmt"
    "github.com/codegangsta/cli"
    "github.com/coreos/go-etcd/etcd"
    "github.com/coreos/etcd/store"
    "launchpad.net/loggo"
)

/*
 *type Config struct {
 *    UserName string
 *    Charm string
 *    StoreType string
 *    Series   string
 *    Version  string
 *    Services map[string]string
 *    Cell     map[string]interface{}
 *}
 */

// Capture possible logging while syncing and write it on the screen.
var log = loggo.GetLogger("judo.hivy.watchers")

func watch_state(c chan *store.Response) {
    // Response object: https://github.com/coreos/etcd/blob/master/store/store.go
    for ;; {
        // Wait for a store.Response on the channel (here, sent by the watcher)
        res := <- c
        log.Debugf("Watch change: %v\n", res)
        log.Infof("Is directory: %v\n", res.Dir)
        log.Infof("Previous value: %v\n", res.PrevValue)
        log.Infof("Value: %v\n", res.Value)
        log.Infof("Action: %v\n", res.Action)
        log.Infof("Key: %v\n", res.Key)
        log.Criticalf("New Key: %v\n", res.NewKey)
        log.Errorf("Expiration: %v\n", res.Expiration)
        log.Warningf("TTL: %v\n", res.TTL)
        log.Tracef("Index: %v\n", res.Index)

        // 1.TODO Fetch back from etcdrequired values
        // 2.TODO Exec juju with those parameters (cf client/)
    }
}

func main() {
    // Command line flags configuration
    app := cli.NewApp()
    app.Name = "Hivy"
    app.Usage = "Juju wrapper for hive management"
    app.Version = "0.0.1"

    app.Flags = []cli.Flag {
        cli.StringFlag{"user", "robot", "Project's owner"},
        cli.BoolFlag{"verbose", "Verbose mode"},
    }

    // Main function with the cli package
    app.Action = func(c *cli.Context) {
        // Current logger configuration
        log_level := "judo.hivy.watchers=WARNING"
        if c.Bool("verbose") {
            // User wants it more verbose
            log_level = "judo.hivy.watchers=TRACE"
        }
        loggo.ConfigureLoggers(log_level)
        fmt.Println("Logging level:", loggo.LoggerInfo())
        defer loggo.RemoveWriter("judo.hivy.watchers")

        client := etcd.NewClient()  // default binds to http://0.0.0.0:4001

        // Channel used for state callback
        stateChannel := make(chan *store.Response)
        go watch_state(stateChannel)
        defer close(stateChannel)

        // Initialize <user>/state value
        res, _ := client.Set(fmt.Sprintf("%s/state", c.String("user")), "none", 0)
        log.Debugf("Set response: %+v\n", res)

        // The information we are monitoring
        prefix := fmt.Sprintf("/%s/state", c.String("user"))
        // The given stateChannel will make the function to listen forever
        // 3.TODO use a stop channel as fourth argument, to abort properly
        res, err := client.Watch(prefix, 0, stateChannel, nil)
        if err != nil {
            panic(err)
        }
        log.Infof("Done with watching (%+v)\n", res)
    }

    app.Run(os.Args)
}
