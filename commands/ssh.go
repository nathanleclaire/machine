package commands

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"

	"github.com/codegangsta/cli"
)

func cmdSsh(c *cli.Context) {
	args := c.Args()
	h := getFirstArgHost(c)
	cmd := ""

	currentState, err := h.Driver.GetState()
	if err != nil {
		log.Fatal(err)
	}

	if currentState != state.Running {
		log.Fatalf("Error: Cannot run SSH command: Host %q is not running", h.Name)
	}

	// Loop through the arguments and parse out a command which relies on
	// flags if it exists, for instance an invocation of the form
	// `docker-machine ssh dev -- df -h` would mandate this, otherwise we
	// will accidentally trigger the codegangsta/cli help text because it
	// thinks we are trying to specify codegangsta flags.
	//
	// TODO: I thought codegangsta/cli supported the flag parsing
	// terminator manually, which would mitigate the need for this kind of
	// hack.  We should investigate.
	for i, arg := range args {
		if arg == "--" {
			cmd = strings.Join(args[i+1:], " ")
			break
		}
	}

	if len(c.Args()) <= 1 {
		err := h.CreateSSHShell()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		cmd = strings.Join(args[1:], " ")
		output, err := h.RunSSHCommand(cmd)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print(output)
	}

}
