package plugin

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"
)

// DriverPlugin interface wraps the underlying mechanics of starting a driver plugin
// server and then figuring out where it can be dialed.
type DriverPlugin interface {
	Serve(driverName string) error
	Address() (string, error)
}

type LocalBinaryPlugin struct {
	BinaryPath string
	addrCh     chan string
	errCh      chan error
}

var (
	defaultTimeout = 10 * time.Second
)

func (lbp *LocalBinaryPlugin) execServer() (chan string, chan error) {
	errCh := make(chan error)
	addrCh := make(chan string)

	go func() {
		cmd := exec.Command(lbp.BinaryPath)

		pluginStdout, err := cmd.StdoutPipe()
		if err != nil {
			errCh <- fmt.Errorf("Error getting cmd stdout pipe: %s", err)
		}

		pluginStderr, err := cmd.StderrPipe()
		if err != nil {
			errCh <- fmt.Errorf("Error getting cmd stderr pipe: %s", err)
		}

		lineReader := bufio.NewReader(pluginStdout)
		errScanner := bufio.NewScanner(pluginStderr)

		if err := cmd.Start(); err != nil {
			errCh <- fmt.Errorf("Error starting plugin binary: %s", err)
		}

		addr, err := lineReader.ReadString('\n')
		if err != nil {
			errCh <- fmt.Errorf("Error reading plugin address: %s", err)
		}

		addrCh <- strings.TrimSpace(addr)

		// Scan / print the plugin stderr.
		go func() {
			for errScanner.Scan() {
				log.Debug("PLUGIN ERR => ", strings.TrimSpace(errScanner.Text()))
			}
			if err := errScanner.Err(); err != nil {
				log.Warn("Error scanning plugin stderr: %s", err)
			}
		}()

		// TODO: I'm not sold on this approach, it should be up to the
		// DriverPlugin interface to provide this stream
		for {
			line, err := lineReader.ReadString('\n')
			if err != nil {
				log.Warn(err)
			}

			log.Debug("PLUGIN OUT => ", strings.TrimSpace(line))
		}
	}()

	return addrCh, errCh
}

func (lbp *LocalBinaryPlugin) lookPath(binaryName string) (string, error) {
	// TODO: Add config file where paths can be hardcoded
	return exec.LookPath(binaryName)
}

func (lbp *LocalBinaryPlugin) Serve(driverName string) error {
	log.Debugf("Launching plugin server for driver %s", driverName)

	binaryPath, err := lbp.lookPath(fmt.Sprintf("docker-machine-%s", driverName))
	if err != nil {
		return fmt.Errorf("Error trying to locate plugin binary: %s", err)
	}

	log.Debugf("Found binary path at %s", binaryPath)

	lbp.BinaryPath = binaryPath

	lbp.addrCh, lbp.errCh = lbp.execServer()

	return nil
}

func (lbp *LocalBinaryPlugin) Address() (string, error) {
	select {
	case addr := <-lbp.addrCh:
		log.Debugf("Plugin server listening at address %s", addr)
		return addr, nil
	case err := <-lbp.errCh:
		return "", fmt.Errorf("Error reading address from plugin binary: %s", err)
	case <-time.After(defaultTimeout):
		return "", fmt.Errorf("Failed to dial the plugin server in %s", defaultTimeout)
	}
}
