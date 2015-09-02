package commands

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/state"

	"github.com/docker/machine/cli"
)

func cmdSsh(c *cli.Context) {
	args := c.Args()
	name := args.First()

	if name == "" {
		fatal("Error: Please specify a machine name.")
	}

	store := getStore(c)
	host, err := loadHost(store, name)
	if err != nil {
		fatal(err)
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		fatal(err)
	}

	if currentState != state.Running {
		fatalf("Error: Cannot run SSH command: Host %q is not running", host.Name)
	}

	if len(c.Args()) == 1 {
		err := host.CreateSSHShell()
		if err != nil {
			fatal(err)
		}
	} else {
		output, err := host.RunSSHCommand(strings.Join(c.Args().Tail(), " "))
		if err != nil {
			fatal(err)
		}

		fmt.Print(output)
	}

}
