package plugin

import (
	"bufio"
	"fmt"
	"io"
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
	Close() error
}

type LocalBinaryPlugin struct {
	Addr       string
	BinaryPath string
	addrCh     chan string
	addrErrCh  chan error
	stopCh     chan bool
}

var (
	defaultTimeout = 10 * time.Second
)

func attachStream(scanner *bufio.Scanner, streamOutCh chan<- string) {
	for scanner.Scan() {
		streamOutCh <- strings.TrimSpace(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			log.Warnf("Unexpected error scanning stream: %s", err)
		} else {
			return
		}
	}
}

func (lbp *LocalBinaryPlugin) execServer() (chan string, chan error) {
	// Channels for communicating results of exec-ing the RPC server.
	addrErrCh := make(chan error)
	addrCh := make(chan string)

	// Channels for sending messages from the plugin's STDOUT or STDERR.
	stdOutCh := make(chan string)
	stdErrCh := make(chan string)

	// Channel for communicating when the reading of the plugin's output
	// should stop.  This is part of the teardown procedure for a plugin.
	lbp.stopCh = make(chan bool)

	go func() {
		cmd := exec.Command(lbp.BinaryPath)

		pluginStdout, err := cmd.StdoutPipe()
		if err != nil {
			addrErrCh <- fmt.Errorf("Error getting cmd stdout pipe: %s", err)
		}

		pluginStderr, err := cmd.StderrPipe()
		if err != nil {
			addrErrCh <- fmt.Errorf("Error getting cmd stderr pipe: %s", err)
		}

		outScanner := bufio.NewScanner(pluginStdout)
		errScanner := bufio.NewScanner(pluginStderr)

		if err := cmd.Start(); err != nil {
			addrErrCh <- fmt.Errorf("Error starting plugin binary: %s", err)
		}

		outScanner.Scan()
		addr := outScanner.Text()
		if err := outScanner.Err(); err != nil {
			addrErrCh <- fmt.Errorf("Error reading plugin address: %s", err)
		}

		addrCh <- strings.TrimSpace(addr)

		close(addrCh)
		close(addrErrCh)

		// Scan / print the plugin stderr.
		// TODO: I'm not sold on this approach, it should be up to the
		// DriverPlugin interface to provide this stream
		go attachStream(errScanner, stdErrCh)
		go attachStream(outScanner, stdOutCh)

		for {
			select {
			case out := <-stdOutCh:
				log.Debug("PLUGIN OUT => ", out)
			case err := <-stdErrCh:
				log.Debug("PLUGIN ERR => ", err)
			case _ = <-lbp.stopCh:
				// TODO: This is still not very safe (sharing
				// these structures between goroutines), figure
				// out a better way.
				pluginStdout.Close()
				pluginStderr.Close()
				close(lbp.stopCh)
				close(stdErrCh)
				close(stdOutCh)
				return
			}
		}
	}()

	return addrCh, addrErrCh
}

func (lbp *LocalBinaryPlugin) Serve(driverName string) error {
	log.Debugf("Launching plugin server for driver %s", driverName)

	binaryPath, err := exec.LookPath(fmt.Sprintf("docker-machine-%s", driverName))
	if err != nil {
		return fmt.Errorf("Error trying to locate plugin binary: %s", err)
	}

	log.Debugf("Found binary path at %s", binaryPath)

	lbp.BinaryPath = binaryPath

	lbp.addrCh, lbp.addrErrCh = lbp.execServer()

	return nil
}

func (lbp *LocalBinaryPlugin) Address() (string, error) {
	if lbp.Addr == "" {
		select {
		case addr := <-lbp.addrCh:
			lbp.Addr = addr
			log.Debugf("Plugin server listening at address %s", addr)
			return addr, nil
		case err := <-lbp.addrErrCh:
			return "", fmt.Errorf("Error reading address from plugin binary: %s", err)
		case <-time.After(defaultTimeout):
			return "", fmt.Errorf("Failed to dial the plugin server in %s", defaultTimeout)
		}
	}
	return lbp.Addr, nil
}

func (lbp *LocalBinaryPlugin) Close() {
	lbp.stopCh <- true
}
