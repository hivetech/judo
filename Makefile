# Makefile for juju-core.
# vim:ft=make

LOGS?=/tmp/make.logs
PROJECT=github.com/Gusabi/judo
JUJU_PATH?=${GOPATH}/src/launchpad.net/juju-core

# Default target.  Compile, just to see if it will.
build:
	go build $(PROJECT)/...

# Run tests.
check:
	go test $(PROJECT)/...

# Reformat the source files.
format:
	go fmt $(PROJECT)/...

# Install packages required to develop Juju and run tests.
install-dependencies:
	#sudo apt-get install build-essential bzr zip git-core mercurial distro-info-data golang-go
	sudo apt-get install build-essential bzr zip git-core mercurial distro-info-data
	@echo
	@echo "Make sure you have MongoDB installed.  See the README file."
	@if [ -z "$(GOPATH)" ]; then \
		echo; \
		echo "You need to set up a GOPATH.  See the README file."; \
	fi

# Invoke gofmt's "simplify" option to streamline the source code.
simplify:
	find "$(GOPATH)/src/$(PROJECT)/" -name \*.go | xargs gofmt -w -s

deps:
	test -d ${GOROOT} || apt-get update 2>&1 >> ${LOGS}
	@echo "[make] Installing packages: go lang and its compiler"
	test -d ${GOROOT} || apt-get -y install bzr go gccgo >> ${LOGS}
	@echo "[make] Installing go packages, somewhere in ${GOPATH}..."
	#NOTE Will the following deps will be installed along ?
	#FIXME go get -u launchpad.net/juju-core

	go get -u launchpad.net/gnuflag
	go get -u launchpad.net/gocheck
	#FIXME go get -u launchpad.net/goamz
	go get -u launchpad.net/goyaml
	go get -u launchpad.net/loggo
	go get -u launchpad.net/goose
	go get -u launchpad.net/gwacl
	go get -u launchpad.net/gomaasapi
	go get -u launchpad.net/lpad
	go get -u launchpad.net/tomb
	go get -u launchpad.net/golxc

	go get -u labix.org/v2/mgo

	#FIXME go get -u code.google.com/p/go.crypto
	#FIXME go get -u code.google.com/p/go.net
	
	go get -u github.com/dotcloud/docker

	go get -u github.com/garyburd/redigo/redis

#patch: deps install-dependencies
patch:
	#go get -u launchpad.net/juju-core
	@echo "Updating import headers"
	find . -name \*.go -print | xargs sed -ire "s/github.com\/Gusabi\/judo/launchpad.net\/juju-core/g"

	@echo "Patching juju-core sources"
	cp cmd/jujud/machine.go ${JUJU_PATH}/cmd/jujud

	#FIXME Are dock containers lxc type container ? For now, no, but could be simplified
	cp instance/container* ${JUJU_PATH}/instance
	cp -r provider/* ${JUJU_PATH}/provider

	cp -r container/dock ${JUJU_PATH}/container

	cp worker/provisioner/dock-* ${JUJU_PATH}/worker/provisioner
	cp worker/provisioner/provisioner* ${JUJU_PATH}/worker/provisioner

	cp cmd/juju/bootstrap.go ${JUJU_PATH}/cmd/juju

	cp -r environs/ansible ${JUJU_PATH}/environs

	cp version/version.go ${JUJU_PATH}/version/

	cp agent/agent.go ${JUJU_PATH}/agent/

	@echo "Preparing ansible"
	cp ansible ansible/ansible.cfg /etc/ansible
	cp ansible /var/lib/juju
	cp init-juju-image.sh ${GOPATH}/bin

.PHONY: build check format install-dependencies simplify
