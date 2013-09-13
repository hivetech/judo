# Makefile for juju-core.
# vim:ft=make

LOGS?=/tmp/make.logs
PROJECT=github.com/Gusabi/judo
JUJU_PATH?=${GOPATH}/src/launchpad.net/juju-core
DOCKER_PATH?=${GOPATH}/src/github.com/dotcloud/docker

# Run tests.
check:
	go test $(PROJECT)/...

# Reformat the source files.
format:
	go fmt $(PROJECT)/...

# Install packages required to develop Juju and run tests.
install:
	#sudo apt-get install build-essential bzr zip git-core mercurial distro-info-data golang-go
	sudo sh -c "curl http://get.docker.io/gpg | apt-key add -"
	# Add the Docker repository to your apt sources list.
	sudo sh -c "echo deb https://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list"
	sudo apt-get install -y linux-image-extra-`uname -r` 2>&1 >> ${LOGS}
	apt-get update 2>&1 >> ${LOGS}
	sudo apt-get install -y build-essential bzr zip git-core mercurial distro-info-data redis 2>&1 >> ${LOGS}
	sudo apt-get install lxc-docker 2>&1 >> ${LOGS}

	go get -u github.com/dotcloud/docker
	go get -u github.com/garyburd/redigo/redis
	#FIXME go get -u launchpad.net/juju-core
	@echo
	@echo "Make sure you have MongoDB installed.  See the README file."
	@if [ -z "$(GOPATH)" ]; then \
		echo; \
		echo "You need to set up a GOPATH.  See the README file."; \
	fi

# Invoke gofmt's "simplify" option to streamline the source code.
simplify:
	find "$(GOPATH)/src/$(PROJECT)/" -name \*.go | xargs gofmt -w -s

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

	cp cmd/juju/bootstrap* ${JUJU_PATH}/cmd/juju

	cp -r environs/ansible ${JUJU_PATH}/environs
	#cp environs/cloudinit/cloudinit.go ${JUJU_PATH}/environs/cloudinit

	#cp version/version* ${JUJU_PATH}/version/

	cp agent/agent* ${JUJU_PATH}/agent/
	cp environs/config/config* ${JUJU_PATH}/environs/config/

	@echo "Preparing ansible"
	cp ansible/ansible.cfg /etc/ansible
	test -d /var/lib/juju || mkdir /var/lib/juju
	cp -r ansible /var/lib/juju
	cp bin/* ${GOPATH}/bin

	@echo "Patching docker network file"
	cp docker/network.go ${DOCKER_PATH}

	chown -R xavier ${JUJU_PATH}

.PHONY: build check format install-dependencies simplify
