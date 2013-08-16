#!/bin/sh

sudo sh -c "
    export GOMAXPROCS=\"$GOMAXPROCS\"
    export GOPATH=\"$GOPATH\"
    export GOROOT=\"$GOROOT\"
    export PATH=\"$PATH\"
    /opt/go/bin/go run dockerlib.go
"
