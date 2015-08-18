package driverfactory

import (
	"fmt"

	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/drivers/virtualbox"
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
	case "none":
		driver = &none.Driver{}
	default:
		return nil, fmt.Errorf("Driver %q not recognized", driverName)
	}

	return driver, nil
}
