package hive_client

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "launchpad.net/goyaml"
    "io/ioutil"
    "github.com/codegangsta/cli"
    //"github.com/coreos/go-etcd/etcd"
)

// https://github.com/coreos/etcd/blob/master/store/store.go
func hive_client() {
/*
 *    c := etcd.NewClient() // default binds to http://0.0.0.0:4001
 *
 *    res, _ := c.Set("test", "bar", 0)
 *    fmt.Printf("set response: %+v\n", res)
 *
 *    values, _ := c.Get("test")
 *    for i, res := range values { // .. and print them out
 *        fmt.Printf("[%d] get response: %+v\n", i, res)
 *        fmt.Printf("Action: %v\n", res.Action)
 *        fmt.Printf("Key: %v\n", res.Key)
 *    }
 *
 *    res, _ = c.Delete("foo")
 *    fmt.Printf("delete response: %+v\n", res)
 */

    app := cli.NewApp()
    app.Name = "Hivy"
    app.Usage = "Juju wrapper for hive management"
    app.Version = "0.0.1"

    app.Flags = []cli.Flag {
        cli.StringFlag{"project", "", "Project name"},
        cli.StringFlag{"username", "", "Project's owner"},
    }

    app.Action = func(c *cli.Context) {
        conf, err := ReadProjectConfig(c.String("project"), c.String("username"))
        if err != nil {
            panic(err);
        }
        fmt.Printf("Config: %v\n", conf)
        fmt.Printf("Editor: %s\n", conf.Cell["editor"])
        fmt.Printf("Database: %s\n", conf.Services["db"])
        //FIXME How interface index should be prcessed ?
        //fmt.Printf("2nd plugin: %s\n", conf.Cell["plugins"][1])

        if err := deploy(conf); err != nil {
            panic(err)
        }
    }

    app.Run(os.Args)
}

type Config struct {
    UserName string
    Charm string
    StoreType string
    Series   string
    Version  string
    Services map[string]string
    Cell     map[string]interface{}
}

func deploy(conf *Config) error {
    juju_bin    := "/home/xavier/dev/goworkspace/bin/juju"
    repo_path   := "/home/xavier/dev/goworkspace/src/github.com/Gusabi/judo/charms"
    deploy_args := fmt.Sprintf("--repository=%s %s:%s/%s %s-%s",
        repo_path, conf.StoreType, conf.Series, conf.Charm, conf.UserName, conf.Charm)

    target_name := fmt.Sprintf("%s-%s", conf.UserName, conf.Charm)
    target_repr := fmt.Sprintf("%s:%s/%s", conf.StoreType, conf.Series, conf.Charm)
    fmt.Printf("Deployment: %s deploy %v\n", juju_bin, deploy_args)
    cmd := exec.Command(juju_bin, "deploy", "--repository", repo_path, target_repr, target_name)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("Running juju deploy %v: %v", deploy_args, err)
    }
    return nil
}

func ReadProjectConfig(project string, username string) (*Config, error) {
    envPath := filepath.Join("/home/git/airport", project, "hive.yml")
	data, err := ioutil.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
            return nil, fmt.Errorf("No project conf found: %s\n", envPath)
		}
		return nil, err
	}
	c, err := ReadProjectConfigBytes(data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q: %v", envPath, err)
	}
    c.UserName = username
    //TODO Automatic detection
    c.StoreType = "local"
	return c, nil
}

func ReadProjectConfigBytes(data []byte) (*Config, error) {
    var raw *Config
    err := goyaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

    //TODO Default values

	return raw, nil
}
