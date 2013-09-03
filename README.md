judo
====

An attempt to extend juju software (https://juju.ubuntu.com/), initially with
docker provider (https://github.com/dotcloud/docker).

Installation
------------

**This package is not yet ready for use, but you're very welcome to fork/download for contribution**

```
$ export GOPATH=/somewhere/to/install  # If not already set
$ go get github.com/Gusabi/judo
```

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
