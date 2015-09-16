package commands

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/docker/machine/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/swarm"
)

var (
	ErrDriverNotRecognized = errors.New("Driver not recognized.")
	sharedCreateFlags      = []cli.Flag{
		cli.StringFlag{
			Name: "driver, d",
			Usage: fmt.Sprintf(
				"Driver to create machine with.",
			),
			Value: "none",
		},
		cli.StringFlag{
			Name:   "engine-install-url",
			Usage:  "Custom URL to use for engine installation",
			Value:  "https://get.docker.com",
			EnvVar: "MACHINE_DOCKER_INSTALL_URL",
		},
		cli.StringSliceFlag{
			Name:  "engine-opt",
			Usage: "Specify arbitrary flags to include with the created engine in the form flag=value",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "engine-insecure-registry",
			Usage: "Specify insecure registries to allow with the created engine",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "engine-registry-mirror",
			Usage: "Specify registry mirrors to use",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "engine-label",
			Usage: "Specify labels for the created engine",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "engine-storage-driver",
			Usage: "Specify a storage driver to use with the engine",
		},
		cli.StringSliceFlag{
			Name:  "engine-env",
			Usage: "Specify environment variables to set in the engine",
			Value: &cli.StringSlice{},
		},
		cli.BoolFlag{
			Name:  "swarm",
			Usage: "Configure Machine with Swarm",
		},
		cli.StringFlag{
			Name:   "swarm-image",
			Usage:  "Specify Docker image to use for Swarm",
			Value:  "swarm:latest",
			EnvVar: "MACHINE_SWARM_IMAGE",
		},
		cli.BoolFlag{
			Name:  "swarm-master",
			Usage: "Configure Machine to be a Swarm master",
		},
		cli.StringFlag{
			Name:  "swarm-discovery",
			Usage: "Discovery service to use with Swarm",
			Value: "",
		},
		cli.StringFlag{
			Name:  "swarm-strategy",
			Usage: "Define a default scheduling strategy for Swarm",
			Value: "spread",
		},
		cli.StringSliceFlag{
			Name:  "swarm-opt",
			Usage: "Define arbitrary flags for swarm",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "swarm-host",
			Usage: "ip/socket to listen on for Swarm master",
			Value: "tcp://0.0.0.0:3376",
		},
		cli.StringFlag{
			Name:  "swarm-addr",
			Usage: "addr to advertise for Swarm (default: detect and use the machine IP)",
			Value: "",
		},
	}
)

func cmdCreateInner(c *cli.Context) {
	name := c.Args().First()
	driverName := c.String("driver")
	certInfo := getCertPathInfoFromContext(c)

	store := &persist.Filestore{
		Path:             c.GlobalString("storage-path"),
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	}

	if name == "" {
		cli.ShowCommandHelp(c, "create")
		log.Fatal("You must specify a machine name")
	}

	if err := validateSwarmDiscovery(c.String("swarm-discovery")); err != nil {
		log.Fatalf("Error parsing swarm discovery: %s", err)
	}

	// TODO: Fix hacky JSON solution
	bareDriverData, err := json.Marshal(&drivers.BaseDriver{
		MachineName:  name,
		ArtifactPath: c.GlobalString("storage-path"),
	})
	if err != nil {
		log.Fatalf("Error attempting to marshal bare driver data: %s", err)
	}

	driver, err := newPluginDriver(driverName, bareDriverData)
	if err != nil {
		log.Fatalf("Error loading driver %q: %s", driverName, err)
	}

	h, err := store.NewHost(driver)
	if err != nil {
		log.Fatalf("Error getting new host: %s", err)
	}

	h.HostOptions = &host.HostOptions{
		AuthOptions: &auth.AuthOptions{
			CertDir:          mcndirs.GetMachineCertDir(),
			CaCertPath:       certInfo.CaCertPath,
			CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
			ClientCertPath:   certInfo.ClientCertPath,
			ClientKeyPath:    certInfo.ClientKeyPath,
			ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), name, "server.pem"),
			ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), name, "server-key.pem"),
		},
		EngineOptions: &engine.EngineOptions{
			ArbitraryFlags:   c.StringSlice("engine-opt"),
			Env:              c.StringSlice("engine-env"),
			InsecureRegistry: c.StringSlice("engine-insecure-registry"),
			Labels:           c.StringSlice("engine-label"),
			RegistryMirror:   c.StringSlice("engine-registry-mirror"),
			StorageDriver:    c.String("engine-storage-driver"),
			TlsVerify:        true,
			InstallURL:       c.String("engine-install-url"),
		},
		SwarmOptions: &swarm.SwarmOptions{
			IsSwarm:        c.Bool("swarm"),
			Image:          c.String("swarm-image"),
			Master:         c.Bool("swarm-master"),
			Discovery:      c.String("swarm-discovery"),
			Address:        c.String("swarm-addr"),
			Host:           c.String("swarm-host"),
			Strategy:       c.String("swarm-strategy"),
			ArbitraryFlags: c.StringSlice("swarm-opt"),
		},
	}

	exists, err := store.Exists(h.Name)
	if err != nil {
		log.Fatalf("Error checking if host exists: %s", err)
	}
	if exists {
		log.Fatal(mcnerror.ErrHostAlreadyExists{h.Name})
	}

	// driverOpts is the actual data we send over the wire to set the
	// driver parameters (an interface fulfilling drivers.DriverOptions,
	// concrete type rpcdriver.RpcFlags).
	mcnFlags := driver.GetCreateFlags()
	driverOpts := getDriverOpts(c, mcnFlags)

	if err := h.Driver.SetConfigFromFlags(driverOpts); err != nil {
		log.Fatalf("Error setting machine configuration from flags provided: %s", err)
	}

	if err := libmachine.Create(store, h); err != nil {
		log.Fatal(err)
	}

	if err := saveHost(store, h); err != nil {
		log.Fatalf("Error attempting to save store: %s", err)
	}

	info := fmt.Sprintf("%s env %s", c.App.Name, name)
	log.Infof("To see how to connect Docker to this machine, run: %s", info)
}

func cmdCreateOuter(c *cli.Context) {
	var (
		driverName string
	)

	// HACK: Parse driver flag first before doing full flag lookup.
	for i, arg := range os.Args {
		if strings.Contains(arg, "-d") {
			// format '--driver foo' or '-d foo'
			if arg == "-d" || arg == "--driver" {
				if i+1 < len(os.Args) {
					driverName = os.Args[i+1]
				}
			}
			// format '--driver=foo' or '-d=foo'
			if strings.HasPrefix(arg, "-d=") || strings.HasPrefix(arg, "--driver=") {
				driverName = strings.Split(arg, "=")[1]
			}
			continue
		}
	}

	// We didn't recognize the driver name.
	if driverName == "" {
		cli.ShowCommandHelp(c, "create")
		return
	}

	name := c.Args().First()

	// TODO: Fix hacky JSON solution
	bareDriverData, err := json.Marshal(&drivers.BaseDriver{
		MachineName: name,
	})
	if err != nil {
		log.Fatalf("Error attempting to marshal bare driver data: %s", err)
	}

	driver, err := newPluginDriver(driverName, bareDriverData)
	if err != nil {
		log.Fatalf("Error loading driver %q: %s", driverName, err)
	}

	defer ClosePluginServers()

	// TODO: So much flag manipulation and voodoo here, it seems to be
	// asking for trouble.
	//
	// mcnFlags is the data we get back over the wire (type mcnflag.Flag)
	// to indicate which parameters are available.
	mcnFlags := driver.GetCreateFlags()

	// This bit will actually make "create" display the correct flags based
	// on the requested driver.
	cliFlags, err := convertMcnFlagsToCliFlags(mcnFlags)
	if err != nil {
		log.Fatalf("Error trying to convert provided driver flags to cli flags: %s", err)
	}
	for i := range c.App.Commands {
		cmd := &c.App.Commands[i]
		if cmd.HasName("create") {
			cmd = addDriverFlagsToCommand(cliFlags, cmd)
		}
	}

	if err := c.App.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getDriverOpts(c *cli.Context, mcnflags []mcnflag.Flag) drivers.DriverOptions {
	// TODO: This function is pretty damn YOLO and would benefit from some
	// sanity checking around types and assertions.
	//
	// But, we need it so that we can actually send the flags for creating
	// a machine over the wire (cli.Context is a no go since there is so
	// much stuff in it).
	driverOpts := rpcdriver.RpcFlags{
		Values: make(map[string]interface{}),
	}

	for _, f := range mcnflags {
		driverOpts.Values[f.Name] = f.Value

		// Hardcoded logic for boolean... :(
		if f.Value == nil {
			driverOpts.Values[f.Name] = false
		}
	}

	for _, name := range c.FlagNames() {
		getter, ok := c.Generic(name).(flag.Getter)
		if !ok {
			// TODO: This is pretty hacky.  StringSlice is the only
			// type so far we have to worry about which is not a
			// Getter, though.
			driverOpts.Values[name] = c.StringSlice(name)
			continue
		}
		driverOpts.Values[name] = getter.Get()
	}

	return driverOpts
}

func convertMcnFlagsToCliFlags(mcnFlags []mcnflag.Flag) ([]cli.Flag, error) {
	cliFlags := []cli.Flag{}
	for _, f := range mcnFlags {
		switch t := f.Value.(type) {
		// TODO: It seems pretty wrong to just default "nil" to this,
		// but cli.BoolFlag doesn't have a "Value" field (false is
		// always the default)
		case nil:
			cliFlags = append(cliFlags, cli.BoolFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
			})
		case int:
			cliFlags = append(cliFlags, cli.IntFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
				Value:  f.Value.(int),
			})
		case string:
			cliFlags = append(cliFlags, cli.StringFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
				Value:  f.Value.(string),
			})
		case []string:
			cliFlags = append(cliFlags, cli.StringSliceFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,

				//TODO: Is this used with defaults? Can we convert the literal []string to cli.StringSlice properly?
				Value: &cli.StringSlice{},
			})
		default:
			log.Warn("Flag is ", f)
			return nil, fmt.Errorf("Flag is unrecognized flag type: %T", t)
		}
	}

	return cliFlags, nil
}

func addDriverFlagsToCommand(cliFlags []cli.Flag, cmd *cli.Command) *cli.Command {
	cmd.Flags = append(sharedCreateFlags, cliFlags...)
	cmd.SkipFlagParsing = false
	cmd.Action = cmdCreateInner
	sort.Sort(ByFlagName(cmd.Flags))

	return cmd
}

func validateSwarmDiscovery(discovery string) error {
	if discovery == "" {
		return nil
	}

	matched, err := regexp.MatchString(`[^:]*://.*`, discovery)
	if err != nil {
		return err
	}

	if matched {
		return nil
	}

	return fmt.Errorf("Swarm Discovery URL was in the wrong format: %s", discovery)
}
