package main

import (
	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(new(amazonec2.Driver))
}
