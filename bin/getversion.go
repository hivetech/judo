//package dockerlib_test
package main

import (
    "fmt"
    //"testing"
    "judo/godocker"
)

//func TestVersion(t *testing.T} {
//}

func main() {
    version := godocker.GetVersion()
    fmt.Printf("Version: %s", version)
}
