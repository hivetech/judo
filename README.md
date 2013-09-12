judo
====

An attempt to extend [juju software](https://juju.ubuntu.com/), initially with
[docker provider](https://github.com/dotcloud/docker).
As a quickview, juju provide a simple-fast-efficient way to deploy cloud services and 
connect them together. Docker give you a powerful container where services ll be deployed.
This project is connected with the "hivelab" one (could also be know as Spop), a charm-like 
application dedicated to team-project-development and continuous integratrion with docker 
on the juju structure.
Judo implement also an alternative to cloud-init with [ansible](https://github.com/ansible/ansible).

Installation
------------

**This package is not yet ready for use, but you're very welcome to fork/download for contribution**

```
$ export GOPATH=/somewhere/to/install  # If not already set
$ go get github.com/Gusabi/judo
```

USAGE
-----

Same way than configuring your juju : 
** juju version: 1.13.2  **
```
$ juju init
$ juju switch hive
$ juju bootstrap
$ #and that's it, now you can deploy 
$ #aand that's it, use etcd as centralized configuration server

```

Technical Overview
------------------
* Set docker go-API next to lxc one
* Introduce ansible as "initializer", next to cloud-init
* Build the "hive" provisionner (as local, ec2...) which provide by default the ansible/docker workflow

Versioning Semantics
--------------------

[From Carl Boettiger](http://carlboettiger.info/)

Releases will be numbered with the following semantic versioning format:

major.minor.patch

And constructed with the following guidelines:

* Breaking backward compatibility bumps the major (and resets the minor 
  and patch)
* New additions without breaking backward compatibility bumps the minor 
  (and resets the patch)
* Bug fixes and misc changes bumps the patch
