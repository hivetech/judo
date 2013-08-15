package dockerlib_test

import (
    "testing"
    "judo/godocker"
)

func TestVersion(t *testing.T) {
    version := godocker.GetVersion()
    t.Log("Version: %s", version)
    if version != "0.5.3-dev" {
        t.Errorf("Invalid docker version: %s\n", version)
        //t.FailNow()
        //t.Fail()
    } else {
        t.Log("OK docker verion: %s", version)
    }
}
