#!/bin/sh

cp * ${HOME}/.go/src/launchpad.net/juju-core/container/dock
#rm ${HOME}/.go/pkg/linux_amd64/github.com/judo/

#sudo -E /opt/go/bin/go test
sudo sh -c "
    export GOMAXPROCS=\"$GOMAXPROCS\"
    export GOPATH=\"$GOPATH\"
    export GOROOT=\"$GOROOT\"
    export PATH=\"$PATH\"
    /opt/go/bin/go test $*
"
