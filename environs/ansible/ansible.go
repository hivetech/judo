// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package ansible

import (
	"encoding/base64"
	"fmt"
    "os"
    "os/exec"
    "path/filepath"
    "io/ioutil"

	"launchpad.net/goyaml"

	agenttools "launchpad.net/juju-core/agent/tools"
	corecloud "launchpad.net/juju-core/cloudinit"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/environs/cloudinit"
	"launchpad.net/juju-core/names"
	"launchpad.net/juju-core/utils"
)

const (
    DockerCmd string = "/usr/sbin/sshd -D"
)

type playbookConfig struct {
    target, password, path string
    port int64
    host string
    ansible_hosts string
    playbook string
}

func NewPlaybook(machine_name string, ssh_port int64, password string, machine_dir string) *playbookConfig {
    return &playbookConfig{
        target: machine_name,
        port: ssh_port,
        password: password,
        path: machine_dir,
        host: "127.0.0.1",
        ansible_hosts: "/etc/ansible/hosts",
        playbook: "/var/lib/juju/ansible/cloudinit.yml",
    }
}

func SuitItUp(conf playbookConfig) error {
    permission := fmt.Sprintf("%s ansible_ssh_host=%s ansible_ssh_port=%d ansible_ssh_pass=%s\n", 
        conf.target, conf.host, conf.port, conf.password)
    fd, err := os.OpenFile(conf.ansible_hosts, os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }
    defer fd.Close()
    if _, err = fd.WriteString(permission); err != nil {
        panic(err)
    }
    //FIXME Still return /usr/bin/ansible-playbook
    //ansible_bin, err := exec.LookPath("ansible-playbook")
    //if err != nil {
        //return fmt.Errorf("command not found: ansible-playbook")
    //}
    ansible_bin := "/usr/local/bin/ansible-playbook"
    extra_vars := fmt.Sprintf("hosts=%s config_vars=%s", conf.target, filepath.Join(conf.path, "cloud-init"))
    cmd := exec.Command(ansible_bin, conf.playbook, "-u", "root", "--extra-vars", extra_vars)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("** Executing cloudinit playbook: %v", err)
    }
    return nil
}

func StartJujudMachine(target string, machine_tag string, machine_id string) error {
    //TODO Use juju-core/names here to get machinetag from Id
    command := fmt.Sprintf("export JUJU_PROVIDER_TYPE=%s && /var/lib/juju/tools/%s/jujud machine --log-file /var/log/juju/%s.log --data-dir /var/lib/juju --machine-id %s --debug >> /var/log/juju/%s.log 2>&1 &", "hive", machine_tag, machine_tag, machine_id, machine_tag)

    ansible_bin := "/usr/local/bin/ansible"
    cmd := exec.Command(ansible_bin, target, "-u", "root", "-m", "shell", "-a", command)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("** Executing jujud machine starter: %v", err)
    }
    return nil
}

type AnsibleMachineConfig struct {
    cloudinit.MachineConfig
}

func base64yaml(m *config.Config) string {
	data, err := goyaml.Marshal(m.AllAttrs())
	if err != nil {
		// can't happen, these values have been validated a number of times
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func New(cfg *cloudinit.MachineConfig) (*corecloud.Config, error) {
	c := corecloud.New()
	if err := verifyConfig(cfg); err != nil {
		return nil, err
	}
    acfg := AnsibleMachineConfig{*cfg}
    return Configure(&acfg, c)
}

func Configure(cfg *AnsibleMachineConfig, c *corecloud.Config) (*corecloud.Config, error) {
    //Note: should be useless
    c.SetAttr("authorized_keys", cfg.AuthorizedKeys)

    //FIXME Figure out a way to add multiple package with ansible
	//c.AddPackage("git")
	// Perfectly reasonable to install lxc on environment instances and kvm
	// containers.
	//if cfg.MachineContainerType != instance.LXC && cfg.MachineContainerType != instance.DOCK {
		//c.AddPackage("lxc")
	//}

    c.SetAttr("juju_bin", cfg.jujuTools())
    c.SetAttr("juju_dl_path", cfg.Tools.URL)

	// TODO (thumper): work out how to pass the logging config to the children
	debugFlag := ""
	// TODO: disable debug mode by default when the system is stable.
	if true {
		debugFlag = " --debug"
	}

	// We add the machine agent's configuration info
	// before running bootstrap-state so that bootstrap-state
	// has a chance to rerwrite it to change the password.
	// It would be cleaner to change bootstrap-state to
	// be responsible for starting the machine agent itself,
	// but this would not be backwardly compatible.
	machineTag := names.MachineTag(cfg.MachineId)
    //agentConf, err := cfg.addAgentInfo(c, machineTag)
    //if err != nil {
        //return nil, err
    //}
    //acfg, _ := cfg.AgentConfig(machineTag)
    //c.SetAttr("machine_tag", machineTag)
    //c.SetAttr("oldpassword", acfg.GetOldPassword())
	//var password string
	//if cfg.StateInfo == nil {
		//password = cfg.APIInfo.Password
	//} else {
		//password = cfg.StateInfo.Password
	//}
    //acfg.WriteCommands()
    //c.SetAttr("oldpassword", acfg.GetOldPassword())
    //c.SetAttr("password", password)

    //FIXME Always "error"
    //if len(cfg.StateInfo.Addrs) == 0 {
        //c.SetAttr("server_addrs", "error")
    //} else {
        //c.SetAttr("server_addrs", cfg.StateInfo.Addrs[0])
    //}

    //c.SetAttr("cacert", string(cfg.StateInfo.CACert))
    //c.SetAttr("cacert", cfg.StateInfo.CACert)

    //FIXME No example of what this shit actually does
    /*
	 *if cfg.StateServer {
	 *    if cfg.NeedMongoPPA() {
	 *        c.AddAptSource("ppa:juju/stable", "1024R/C8068B11")
	 *    }
	 *    c.AddPackage("mongodb-server")
	 *    certKey := string(cfg.StateServerCert) + string(cfg.StateServerKey)
	 *    c.AddFile(cfg.dataFile("server.pem"), certKey, 0600)
	 *    if err := cfg.addMongoToBoot(c); err != nil {
	 *        return nil, err
	 *    }
	 *    // We temporarily give bootstrap-state a directory
	 *    // of its own so that it can get the state info via the
	 *    // same mechanism as other jujud commands.
	 *    acfg, err := cfg.addAgentInfo(c, "bootstrap")
	 *    if err != nil {
	 *        return nil, err
	 *    }
	 *    c.AddScripts(
	 *        fmt.Sprintf("echo %s > %s", shquote(cfg.StateInfoURL), cloudinit.BootstrapStateURLFile),
	 *        cfg.jujuTools()+"/jujud bootstrap-state"+
	 *            " --data-dir "+shquote(cfg.DataDir)+
	 *            " --env-config "+shquote(base64yaml(cfg.Config))+
	 *            " --constraints "+shquote(cfg.Constraints.String())+
	 *            debugFlag,
	 *        "rm -rf "+shquote(acfg.Dir()),
	 *    )
	 *}
     */

	if err := cfg.addMachineAgentToBoot(c, machineTag, cfg.MachineId, debugFlag); err != nil {
		return nil, err
	}

	// general options
    c.SetAttr("update_cache", true)
    c.SetAttr("upgrade", "yes")
    //FIXME Should logs be handled the same way ?
    /*
     *c.SetAptUpgrade(true)
     *c.SetOutput(corecloud.OutAll, "| tee -a /var/log/cloud-init-output.log", "")
     */

    //TODO use it for api and state
    c.SetAttr("machine_tag", machineTag)
    c.SetAttr("data_dir", cfg.DataDir)

    // Agent configuration
    // Raring configuration style
    //acfg, _ := cfg.AgentConfig(machineTag)
    agent_c := corecloud.New()

    agent_c.SetAttr("data_dir", cfg.DataDir)
    //FIXME oldpassword
    agent_c.SetAttr("oldpassword", cfg.APIInfo.Password)
    //agent_c.SetAttr("oldpassword", acfg.GetOldPassword())
    agent_c.SetAttr("machinenonce", cfg.MachineNonce)
    //FIXME old_api_password
    agent_c.SetAttr("oldapipassword", "")

    // API server
    var api_info = map[string]interface{}{
        "addrs": []string{"10.0.3.1:17070"},
        "password": cfg.APIInfo.Password,
        "cacert": []byte("CA CERT\n" + string(cfg.APIInfo.CACert)),
        "tag": machineTag,
    }
    agent_c.SetAttr("apiinfo", api_info)
    //FIXME c.SetAttr("api_addrs", cfg.APIInfo.Addrs)
    //c.SetAttr("api_cert", cfg.APIInfo.CACert)
    //FIXME c.SetAttr("api_machine_tag", cfg.StateInfo.Tag)

    // State server
    //if cfg.StateServer {
    var state_info = map[string]interface{}{
        "addrs": []string{"10.0.3.1:37017"},
        "password": cfg.APIInfo.Password,
        "cacert": []byte("CA CERT\n" + string(cfg.APIInfo.CACert)),
        "tag": machineTag,
    }
    agent_c.SetAttr("stateinfo", state_info)
    //FIXME c.SetAttr("state_addrs", cfg.StateInfo.Addrs)
    //c.SetAttr("state_password", cfg.StateInfo.Password)
    //c.SetAttr("state_cert", cfg.StateInfo.CACert)
    //FIXME c.SetAttr("state_cacert", []byte("CA CERT\n" + string(cfg.StateInfo.CACert)))
    //FIXME c.SetAttr("state_machine_tag", cfg.StateInfo.Tag)
    //FIXME old_password
    //}

    if err := generateAgentConfig(agent_c, machineTag); err != nil {
        return c, err
    }

	return c, nil
}

func generateAgentConfig(agent_c *corecloud.Config, tag string) (error) {
    data, err := agent_c.Render()
    if err != nil {
        return err
    }
	agentFilename := filepath.Join("/tmp", tag + "-agent.conf")
	if err := ioutil.WriteFile(agentFilename, data, 0644); err != nil {
		return fmt.Errorf("failed to write agent data: %v", err)
	}
    return nil
}

/*
 *func (cfg *MachineConfig) dataFile(name string) string {
 *    return filepath.Join(cfg.DataDir, name)
 *}
 */

/*
 *func (cfg *MachineConfig) agentConfig(tag string) (agent.Config, error) {
 *    // TODO for HAState: the stateHostAddrs and apiHostAddrs here assume that
 *    // if the machine is a stateServer then to use localhost.  This may be
 *    // sufficient, but needs thought in the new world order.
 *    var password string
 *    if cfg.StateInfo == nil {
 *        password = cfg.APIInfo.Password
 *    } else {
 *        password = cfg.StateInfo.Password
 *    }
 *    var configParams = agent.AgentConfigParams{
 *        DataDir:        cfg.DataDir,
 *        Tag:            tag,
 *        Password:       password,
 *        Nonce:          cfg.MachineNonce,
 *        StateAddresses: cfg.stateHostAddrs(),
 *        APIAddresses:   cfg.apiHostAddrs(),
 *        CACert:         cfg.StateInfo.CACert,
 *    }
 *    if cfg.StateServer {
 *        return agent.NewStateMachineConfig(
 *            agent.StateMachineConfigParams{
 *                AgentConfigParams: configParams,
 *                StateServerCert:   cfg.StateServerCert,
 *                StateServerKey:    cfg.StateServerKey,
 *                StatePort:         cfg.StatePort,
 *                APIPort:           cfg.APIPort,
 *            })
 *    }
 *    return agent.NewAgentConfig(configParams)
 *}
 */

func (cfg *AnsibleMachineConfig) addMachineAgentToBoot(c *corecloud.Config, tag, machineId, logConfig string) error {
	// Make the agent run via a symbolic link to the actual tools
	// directory, so it can upgrade itself without needing to change
	// the upstart script.
	toolsDir := agenttools.ToolsDir(cfg.DataDir, tag)
	// TODO(dfc) ln -nfs, so it doesn't fail if for some reason that the target already exists
	//c.AddScripts(fmt.Sprintf("ln -s %v %s", cfg.Tools.Version, shquote(toolsDir)))
    version_struct := cfg.Tools.Version
    c.SetAttr("juju_version", version_struct.String())
    c.SetAttr("tools_dir", toolsDir)
    c.SetAttr("provider", cfg.MachineEnvironment["JUJU_PROVIDER_TYPE"])
    //c.SetAttr("machine_tag", tag)
    c.SetAttr("machine_id", machineId)

	return nil
}

/*
 *func (cfg *MachineConfig) addMongoToBoot(c *corecloud.Config) error {
 *    dbDir := filepath.Join(cfg.DataDir, "db")
 *    c.AddScripts(
 *        "mkdir -p "+dbDir+"/journal",
 *        // Otherwise we get three files with 100M+ each, which takes time.
 *        "dd bs=1M count=1 if=/dev/zero of="+dbDir+"/journal/prealloc.0",
 *        "dd bs=1M count=1 if=/dev/zero of="+dbDir+"/journal/prealloc.1",
 *        "dd bs=1M count=1 if=/dev/zero of="+dbDir+"/journal/prealloc.2",
 *    )
 *
 *    conf := upstart.MongoUpstartService("juju-db", cfg.DataDir, dbDir, cfg.StatePort)
 *    cmds, err := conf.InstallCommands()
 *    if err != nil {
 *        return fmt.Errorf("cannot make cloud-init upstart script for the state database: %v", err)
 *    }
 *    c.AddScripts(cmds...)
 *    return nil
 *}
 */

func (cfg *AnsibleMachineConfig) jujuTools() string {
    return agenttools.SharedToolsDir(cfg.DataDir, cfg.Tools.Version)
}

/*
 *func (cfg *AnsibleMachineConfig) stateHostAddrs() []string {
 *    var hosts []string
 *    if cfg.StateServer {
 *        hosts = append(hosts, fmt.Sprintf("localhost:%d", cfg.StatePort))
 *    }
 *    if cfg.StateInfo != nil {
 *        hosts = append(hosts, cfg.StateInfo.Addrs...)
 *    }
 *    return hosts
 *}
 *
 *func (cfg *AnsibleMachineConfig) apiHostAddrs() []string {
 *    var hosts []string
 *    if cfg.StateServer {
 *        hosts = append(hosts, fmt.Sprintf("localhost:%d", cfg.APIPort))
 *    }
 *    if cfg.APIInfo != nil {
 *        hosts = append(hosts, cfg.APIInfo.Addrs...)
 *    }
 *    return hosts
 *}
 */

/*
 *func (cfg *MachineConfig) NeedMongoPPA() bool {
 *    series := cfg.Tools.Version.Series
 *    // 11.10 and earlier are not supported.
 *    // 13.04 and later ship a compatible version in the archive.
 *    return series == "precise" || series == "quantal"
 *}
 */

/*
 *func shquote(p string) string {
 *    return utils.ShQuote(p)
 *}
 */

type requiresError string

func (e requiresError) Error() string {
	return "invalid machine configuration: missing " + string(e)
}

//TODO Refactore with cloudinit.VerifyConfig
func verifyConfig(cfg *cloudinit.MachineConfig) (err error) {
	defer utils.ErrorContextf(&err, "invalid machine configuration")
	if !names.IsMachine(cfg.MachineId) {
		return fmt.Errorf("invalid machine id")
	}
	if cfg.DataDir == "" {
		return fmt.Errorf("missing var directory")
	}
	if cfg.Tools == nil {
		return fmt.Errorf("missing tools")
	}
	if cfg.Tools.URL == "" {
		return fmt.Errorf("missing tools URL")
	}
	if cfg.StateInfo == nil {
		return fmt.Errorf("missing state info")
	}
	if len(cfg.StateInfo.CACert) == 0 {
		return fmt.Errorf("missing CA certificate")
	}
	if cfg.APIInfo == nil {
		return fmt.Errorf("missing API info")
	}
	if len(cfg.APIInfo.CACert) == 0 {
		return fmt.Errorf("missing API CA certificate")
	}
	if cfg.StateServer {
		if cfg.Config == nil {
			return fmt.Errorf("missing environment configuration")
		}
		if cfg.StateInfo.Tag != "" {
			return fmt.Errorf("entity tag must be blank when starting a state server")
		}
		if cfg.APIInfo.Tag != "" {
			return fmt.Errorf("entity tag must be blank when starting a state server")
		}
		if len(cfg.StateServerCert) == 0 {
			return fmt.Errorf("missing state server certificate")
		}
		if len(cfg.StateServerKey) == 0 {
			return fmt.Errorf("missing state server private key")
		}
		if cfg.StatePort == 0 {
			return fmt.Errorf("missing state port")
		}
		if cfg.APIPort == 0 {
			return fmt.Errorf("missing API port")
		}
	} else {
		if len(cfg.StateInfo.Addrs) == 0 {
			return fmt.Errorf("missing state hosts")
		}
		if cfg.StateInfo.Tag != names.MachineTag(cfg.MachineId) {
			return fmt.Errorf("entity tag must match started machine")
		}
		if len(cfg.APIInfo.Addrs) == 0 {
			return fmt.Errorf("missing API hosts")
		}
		if cfg.APIInfo.Tag != names.MachineTag(cfg.MachineId) {
			return fmt.Errorf("entity tag must match started machine")
		}
	}
	if cfg.MachineNonce == "" {
		return fmt.Errorf("missing machine nonce")
	}
	return nil
}
