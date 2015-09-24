package main

import (
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(new(fakedriver.Driver))
}
