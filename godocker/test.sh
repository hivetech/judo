#!/bin/sh

cp godocker.go ${HOME}/.go/src/github.com/judo/godocker
#rm ${HOME}/.go/pkg/linux_amd64/github.com/judo/

sudo sh -c "
    export GOMAXPROCS=\"$GOMAXPROCS\"
    export GOPATH=\"$GOPATH\"
    export GOROOT=\"$GOROOT\"
    export PATH=\"$PATH\"
    /opt/go/bin/go test $*
"
