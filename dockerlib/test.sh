#!/bin/sh

cp dockerlib.go ${HOME}/.go/src/github.com/judo/dockerlib
#rm ${HOME}/.go/pkg/linux_amd64/github.com/judo/
#sudo -E /opt/go/bin/go test

sudo sh -c "
    export GOMAXPROCS=\"$GOMAXPROCS\"
    export GOPATH=\"$GOPATH\"
    export GOROOT=\"$GOROOT\"
    export PATH=\"$PATH\"
    /opt/go/bin/go test $*
"
