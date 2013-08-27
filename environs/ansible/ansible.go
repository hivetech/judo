// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package ansible

import (
	"encoding/base64"
	"fmt"
	//"path/filepath"

	"launchpad.net/goyaml"

	//"launchpad.net/juju-core/agent"
	agenttools "launchpad.net/juju-core/agent/tools"
	corecloud "launchpad.net/juju-core/cloudinit"
	//"launchpad.net/juju-core/constraints"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/environs/cloudinit"
	//"launchpad.net/juju-core/instance"
	//"launchpad.net/juju-core/log/syslog"
	"launchpad.net/juju-core/names"
	//"launchpad.net/juju-core/state"
	//"launchpad.net/juju-core/state/api"
	//coretools "launchpad.net/juju-core/tools"
	//"launchpad.net/juju-core/upstart"
	"launchpad.net/juju-core/utils"
)

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

     c.SetAttr("data_dir", cfg.DataDir)

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
    c.SetAttr("machine_tag", machineTag)
	var password string
	if cfg.StateInfo == nil {
		password = cfg.APIInfo.Password
	} else {
		password = cfg.StateInfo.Password
	}
    c.SetAttr("oldpassword", password)
    //c.SetAttr("machine_nonce", agentConf.Nonce())
    c.SetAttr("machine_nonce", cfg.MachineNonce)
    //c.SetAttr("cacert", agentConf.StateServerCert)
    //c.SetAttr("cacert", string(cfg.StateInfo.CACert))

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
    //FIXME Should logs be handled the same way ?
    /*
     *c.SetAptUpgrade(true)
     *c.SetOutput(corecloud.OutAll, "| tee -a /var/log/cloud-init-output.log", "")
     */
	return c, nil
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
    //version_struct := cfg.Tools.Version
    //c.SetAttr("juju_version", version_struct.String())
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
