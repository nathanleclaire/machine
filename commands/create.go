package commands

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/codegangsta/cli"
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
)

func cmdCreate(c *cli.Context) {
	driverName := c.String("driver")
	name := c.Args().First()

	certInfo := getCertPathInfoFromContext(c)

	store := &persist.Filestore{
		Path:             c.GlobalString("storage-path"),
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	}

	hostOptions := &host.HostOptions{
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

	// TODO: So much flag manipulation and voodoo here, it seems to be
	// asking for trouble.
	//
	// driverOpts is the actual data we send over the wire to set the
	// driver parameters (an interface fulfilling drivers.DriverOptions,
	// concrete type rpcdriver.RpcFlags).
	//
	// mcnFlags is the data we get back over the wire (type
	// mcnflag.Flag) to indicate which parameters are available.
	mcnFlags := driver.GetCreateFlags()
	driverOpts := getDriverOpts(c, mcnFlags)

	// This bit will actually make "create" display the correct flags based
	// on the requested driver.
	if driverName != "none" {
		// TODO: Fix this, it doesn't work
		cliFlags, err := convertMcnFlagsToCliFlags(mcnFlags)
		if err != nil {
			log.Fatalf("Error trying to convert provided driver flags to cli flags: %s", err)
		}
		c.Command = addDriverFlagsToCommand(cliFlags, c.Command)
	}

	if name == "" {
		cli.ShowCommandHelp(c, "create")
		log.Fatal("You must specify a machine name")
	}

	if err := validateSwarmDiscovery(c.String("swarm-discovery")); err != nil {
		log.Fatalf("Error parsing swarm discovery: %s", err)
	}

	h, err := store.NewHost(driver)
	if err != nil {
		log.Fatalf("Error getting new host: %s", err)
	}

	h.HostOptions = hostOptions

	exists, err := store.Exists(h.Name)
	if err != nil {
		log.Fatalf("Error checking if host exists: %s", err)
	}
	if exists {
		log.Fatal(mcnerror.ErrHostAlreadyExists{h.Name})
	}

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

	for _, f := range mcnflags {
		driverOpts.Values[f.Name] = f.Value

		// Hardcoded logic for boolean... :(
		if f.Value == nil {
			driverOpts.Values[f.Name] = false
		}
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

func addDriverFlagsToCommand(cliFlags []cli.Flag, cmd cli.Command) cli.Command {
	cmd.Flags = append(sharedCreateFlags, cliFlags...)
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
