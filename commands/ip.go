package commands

import (
	"github.com/docker/machine/cli"
	"github.com/docker/machine/libmachine/log"
)

func cmdIp(c *cli.Context) {
	if err := runActionWithContext("ip", c); err != nil {
		log.Fatal(err)
	}
}
