package driverfactory

import (
	"fmt"

	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/drivers/azure"
	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/drivers/exoscale"
	"github.com/docker/machine/drivers/generic"
	"github.com/docker/machine/drivers/google"
	"github.com/docker/machine/drivers/hyperv"
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/drivers/openstack"
	"github.com/docker/machine/drivers/rackspace"
	"github.com/docker/machine/drivers/softlayer"
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/drivers/vmwarevcloudair"
	"github.com/docker/machine/drivers/vmwarevsphere"
	"github.com/docker/machine/libmachine/drivers"
)

func NewDriver(driverName, hostName, artifactPath string) (drivers.Driver, error) {
	var (
		driver drivers.Driver
	)

	switch driverName {
	case "virtualbox":
		driver = virtualbox.NewDriver(hostName, artifactPath)
	case "digitalocean":
		driver = digitalocean.NewDriver(hostName, artifactPath)
	case "amazonec2":
		driver = amazonec2.NewDriver(hostName, artifactPath)
	case "azure":
		driver = azure.NewDriver(hostName, artifactPath)
	case "exoscale":
		driver = exoscale.NewDriver(hostName, artifactPath)
	case "generic":
		driver = generic.NewDriver(hostName, artifactPath)
	case "google":
		driver = google.NewDriver(hostName, artifactPath)
	case "hyperv":
		driver = hyperv.NewDriver(hostName, artifactPath)
	case "openstack":
		driver = openstack.NewDriver(hostName, artifactPath)
	case "rackspace":
		driver = rackspace.NewDriver(hostName, artifactPath)
	case "softlayer":
		driver = softlayer.NewDriver(hostName, artifactPath)
	case "vmwarefusion":
		driver = vmwarefusion.NewDriver(hostName, artifactPath)
	case "vmwarevcloudair":
		driver = vmwarevcloudair.NewDriver(hostName, artifactPath)
	case "vmwarevsphere":
		driver = vmwarevsphere.NewDriver(hostName, artifactPath)
	case "none":
		driver = &none.Driver{}
	default:
		return nil, fmt.Errorf("Driver %q not recognized", driverName)
	}

	return driver, nil
}
