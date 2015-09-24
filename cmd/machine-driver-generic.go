package main

import (
	"github.com/docker/machine/drivers/generic"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(new(generic.Driver))
}
