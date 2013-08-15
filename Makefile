#
# Makefile
# xavier, 2013-07-30 07:56
#
# vim:ft=make
#

LOGS?=/tmp/make.logs

all: deps install
	@echo "Done."

install:
	#FIXME Temporary
	@echo "[make] Installing last revision"
	go get -u github.com/Gusabi/judo

deps:
	test -d ${GOROOT} || apt-get update 2>&1 >> ${LOGS}
	@echo "[make] Installing packages: go lang and its compiler"
	test -d ${GOROOT} || apt-get -y install bzr go gccgo >> ${LOGS}
	@echo "[make] Installing go packages, somewhere in ${GOPATH}..."
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

tests:
	# Run as root
	cd godocker && test.sh
