// Copyright 2013 Canonical Ltd.  This software is licensed under the
// GNU Lesser General Public License version 3 (see the file COPYING).

// This is an example of how to use GWACL to interact with the Azure storage
// services API.
//
// Note that it is provided "as-is" and contains very little error handling.
// Real code should handle errors.

package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "launchpad.net/gwacl"
    "os"
)

var account string
var key string
var filename string
var container string
var prefix string
var blobname string
var blobtype string
var size int
var pagerange string
var acl string

func checkContainerAndFilename() error {
    if container == "" || filename == "" {
        return fmt.Errorf("Must supply container and filename")
    }
    return nil
}

var VALID_CMDS string = "<listcontainers|containeracl|list|getblob|addblock|deleteblob|putblob|putpage>"

func badOperationError() error {
    return fmt.Errorf("Must specify one of %s", VALID_CMDS)
}

func getParams() (string, error) {
    flag.StringVar(&account, "account", "", "Storage account name")
    flag.StringVar(&key, "key", "", "A valid storage account key (base64 encoded), defaults to the empty string (i.e. anonymous access)")
    flag.StringVar(&container, "container", "", "Name of the container to use")
    flag.StringVar(&filename, "filename", "", "File containing blob/page to upload/download")
    flag.StringVar(&prefix, "prefix", "", "Prefix to match when listing blobs")
    flag.StringVar(&blobname, "blobname", "", "Name of blob in container")
    flag.StringVar(&blobtype, "blobtype", "", "Type of blob, 'page' or 'block'")
    flag.IntVar(&size, "size", 0, "Size of blob to create for a page 'putblob'")
    flag.StringVar(&pagerange, "pagerange", "", "When uploading to a page blob, this specifies what range in the blob. Use the format 'start-end', e.g. -pagerange 1024-2048")
    flag.StringVar(&acl, "acl", "", "When using 'containeracl', specify an ACL type")
    flag.Parse()

    if account == "" {
        return "", fmt.Errorf("Must supply account parameter")
    }

    if len(flag.Args()) != 1 {
        return "", badOperationError()
    }

    switch flag.Arg(0) {
    case "listcontainers":
        return "listcontainers", nil
    case "containeracl":
        if container == "" {
            return "", fmt.Errorf("Must supply container with containeracl")
        }
        if key == "" {
            return "", fmt.Errorf("Must supply key with containeracl")
        }
        if acl != "container" && acl != "blob" && acl != "private" {
            return "", fmt.Errorf("Usage: containeracl -container=<container> <container|blob|private>")
        }
        return "containeracl", nil
    case "list":
        if container == "" {
            return "", fmt.Errorf("Must supply container with 'list'")
        }
        return "list", nil
    case "getblob":
        if container == "" || filename == "" {
            return "", fmt.Errorf("Must supply container and filename with 'list'")
        }
        return "getblob", nil
    case "addblock":
        if key == "" {
            return "", fmt.Errorf("Must supply key with addblock")
        }
        err := checkContainerAndFilename()
        if err != nil {
            return "", err
        }
        return "addblock", nil
    case "deleteblob":
        if key == "" {
            return "", fmt.Errorf("Must supply key with deleteblob")
        }
        err := checkContainerAndFilename()
        if err != nil {
            return "", err
        }
        return "deleteblob", nil
    case "putblob":
        if key == "" {
            return "", fmt.Errorf("Must supply key with putblob")
        }
        if blobname == "" || blobtype == "" || container == "" || size == 0 {
            return "", fmt.Errorf("Must supply container, blobname, blobtype and size with 'putblob'")
        }
        return "putblob", nil
    case "putpage":
        if key == "" {
            return "", fmt.Errorf("Must supply key with putpage")
        }
        if blobname == "" || container == "" || pagerange == "" || filename == "" {
            return "", fmt.Errorf("Must supply container, blobname, pagerange and filename")
        }
        return "putpage", nil
    }

    return "", badOperationError()
}

var Usage = func() {
    fmt.Fprintf(os.Stderr, "%s [args] %s\n", os.Args[0], VALID_CMDS)
    flag.PrintDefaults()

    fmt.Fprintf(os.Stderr, `
    This is an example of how to interact with the Azure storage service.
    It is not a complete example but it does give a useful way to do do some
    basic operations.

    The -account param must always be supplied and -key must be supplied for
    operations that change things, (get these from the Azure web UI) otherwise
    anonymous access is made.  Additionally there are the following command
    invocation parameters:

    Show existing storage containers:
        listcontainers

    List files in a container:
        -container=<container> list

    Set access on a container:
        -container=<container> -acl <container|blob|private> containeracl

    Get a file from a container (it's returned on stdout):
        -container=<container> -filename=<filename> getblob

    Upload a file to a block blob:
        -container=<container> -filename=<filename> addblock

    Delete a blob:
        -container=<container> -filename=<filename> deleteblob

    Create an empty page blob:
        -container=<container> -blobname=<blobname> -size=<bytes>
                -blobtype="page" putblob

    Upload a file to a page blob's page.  The range parameters must be
    (modulo 512)-(modulo 512 -1), eg: -pagerange=0-511
        -container=<container> -blobname=<blobname> -pagerange=<N-N>
        -filename=<local file> putpage
    `)
}

func dumpError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "ERROR:")
        fmt.Fprintf(os.Stderr, "%s\n", err)
    }
}

func listcontainers(storage gwacl.StorageContext) {
    res, e := storage.ListAllContainers()
    if e != nil {
        dumpError(e)
        return
    }
    for _, c := range res.Containers {
        // TODO: embellish with the other container data
        fmt.Println(c.Name)
    }
}

func containeracl(storage gwacl.StorageContext) {
    err := storage.SetContainerACL(&gwacl.SetContainerACLRequest{
        Container: container,
        Access:    acl,
    })
    dumpError(err)
}

func list(storage gwacl.StorageContext) {
    request := &gwacl.ListBlobsRequest{
        Container: container, Prefix: prefix}
    res, err := storage.ListAllBlobs(request)
    if err != nil {
        dumpError(err)
        return
    }
    for _, b := range res.Blobs {
        fmt.Printf("%s, %s, %s\n", b.ContentLength, b.LastModified, b.Name)
    }
}

func addblock(storage gwacl.StorageContext) {
    var err error
    file, err := os.Open(filename)
    if err != nil {
        dumpError(err)
        return
    }
    defer file.Close()

    err = storage.UploadBlockBlob(container, filename, file)
    if err != nil {
        dumpError(err)
        return
    }
}

func deleteblob(storage gwacl.StorageContext) {
    err := storage.DeleteBlob(container, filename)
    dumpError(err)
}

func getblob(storage gwacl.StorageContext) {
    var err error
    file, err := storage.GetBlob(container, filename)
    if err != nil {
        dumpError(err)
        return
    }
    data, err := ioutil.ReadAll(file)
    if err != nil {
        dumpError(err)
        return
    }
    os.Stdout.Write(data)
}

func putblob(storage gwacl.StorageContext) {
    err := storage.PutBlob(&gwacl.PutBlobRequest{
        Container: container,
        BlobType:  blobtype,
        Filename:  blobname,
        Size:      size,
    })
    dumpError(err)
}

func putpage(storage gwacl.StorageContext) {
    var err error
    file, err := os.Open(filename)
    if err != nil {
        dumpError(err)
        return
    }
    defer file.Close()

    var start, end int
    fmt.Sscanf(pagerange, "%d-%d", &start, &end)

    err = storage.PutPage(&gwacl.PutPageRequest{
        Container:  container,
        Filename:   blobname,
        StartRange: start,
        EndRange:   end,
        Data:       file,
    })
    if err != nil {
        dumpError(err)
        return
    }
}

func main() {
    flag.Usage = Usage
    var err error
    op, err := getParams()
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
        flag.Usage()
        os.Exit(1)
    }

    storage := gwacl.StorageContext{
        Account: account,
        Key:     key,
    }

    switch op {
    case "listcontainers":
        listcontainers(storage)
    case "containeracl":
        containeracl(storage)
    case "list":
        list(storage)
    case "addblock":
        addblock(storage)
    case "deleteblob":
        deleteblob(storage)
    case "getblob":
        getblob(storage)
    case "putblob":
        putblob(storage)
    case "putpage":
        putpage(storage)
    }
}
