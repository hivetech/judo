// Copyright 2013 Canonical Ltd.  This software is licensed under the
// GNU Lesser General Public License version 3 (see the file COPYING).

// Define the various role sizes and their attributes.

package gwacl

// RoleSize is a representation of the data available in the Azure
// documentation here:
// http://msdn.microsoft.com/en-us/library/windowsazure/dn197896.aspx and
// pricing from here:
// http://www.windowsazure.com/en-us/pricing/details/virtual-machines/
// and is correct as of 2013-07-22.
//
// The Disk Space values are only the maxumim permitted; actual space is
// determined by the OS image being used.
type RoleSize struct {
    Name             string
    CpuCores         uint64
    Mem              uint64 // In MB
    OSDiskSpaceCloud uint64 // In MB
    OSDiskSpaceVirt  uint64 // In MB
    MaxDataDisks     uint64 // 1TB each
    Cost             uint64 // USD cents per hour
}

const (
    // MB is the unit in which we specify sizes, so it's 1.
    // But please include it anyway, so that units are always explicit.
    MB  = 1
    GB  = 1024 * MB
    TB  = 1024 * GB
)

var RoleSizes = []RoleSize{
    {
        Name:             "ExtraSmall",
        CpuCores:         1,  // shared
        Mem:              768 * MB,
        OSDiskSpaceCloud: 19 * GB,
        OSDiskSpaceVirt:  20 * GB,
        MaxDataDisks:     1,
        Cost:             2,
    }, {
        Name:             "Small",
        CpuCores:         1,
        Mem:              1.75 * GB,
        OSDiskSpaceCloud: 224 * GB,
        OSDiskSpaceVirt:  70 * GB,
        MaxDataDisks:     2,
        Cost:             6,
    }, {
        Name:             "Medium",
        CpuCores:         2,
        Mem:              3.5 * GB,
        OSDiskSpaceCloud: 489 * GB,
        OSDiskSpaceVirt:  135 * GB,
        MaxDataDisks:     4,
        Cost:             12,
    }, {
        Name:             "Large",
        CpuCores:         4,
        Mem:              7 * GB,
        OSDiskSpaceCloud: 999 * GB,
        OSDiskSpaceVirt:  285 * GB,
        MaxDataDisks:     8,
        Cost:             24,
    }, {
        Name:             "ExtraLarge",
        CpuCores:         8,
        Mem:              14 * GB,
        OSDiskSpaceCloud: 2039 * GB,
        OSDiskSpaceVirt:  65 * GB,
        MaxDataDisks:     16,
        Cost:             48,
    }, {
        Name:             "A6",
        CpuCores:         4,
        Mem:              28 * GB,
        OSDiskSpaceCloud: 999 * GB,
        OSDiskSpaceVirt:  285 * GB,
        MaxDataDisks:     8,
        Cost:             82,
    }, {
        Name:             "A7",
        CpuCores:         8,
        Mem:              56 * GB,
        OSDiskSpaceCloud: 2039 * GB,
        OSDiskSpaceVirt:  65 * GB,
        MaxDataDisks:     16,
        Cost:             164,
    },
}

var RoleNameMap map[string]RoleSize = make(map[string]RoleSize)

func init() {
    for _, rolesize := range RoleSizes {
        RoleNameMap[rolesize.Name] = rolesize
    }
}
