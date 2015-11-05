package drivers

import (
	"sync"

	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

var globalLock = &sync.Mutex{}

// synchronizedDriverDecorator is a decorator around a Driver that will lock on a mutex
// before each function call.
// This comes in handy to protect from drivers that can't be called from multiple go routines in parallel.
type synchronizedDriverDecorator struct {
	Delegate Driver
	lock     sync.Locker
}

// SynchronizeGlobal wraps a driver to synchronize each function on a global shared lock.
func SynchronizeGlobal(driver Driver) Driver {
	return SynchronizeOnLock(driver, globalLock)
}

// SynchronizeOnLock wraps a driver to synchronize each function on a given lock.
func SynchronizeOnLock(driver Driver, lock sync.Locker) Driver {
	return &synchronizedDriverDecorator{
		Delegate: driver,
		lock:     lock,
	}
}

// Create a host using the driver's config
func (d *synchronizedDriverDecorator) Create() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.Create()
}

// DriverName returns the name of the driver as it is registered
func (d *synchronizedDriverDecorator) DriverName() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.DriverName()
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *synchronizedDriverDecorator) GetCreateFlags() []mcnflag.Flag {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetCreateFlags()
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *synchronizedDriverDecorator) GetIP() (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetIP()
}

// GetMachineName returns the name of the machine
func (d *synchronizedDriverDecorator) GetMachineName() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetMachineName()
}

// GetSSHHostname returns hostname for use with ssh
func (d *synchronizedDriverDecorator) GetSSHHostname() (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetSSHHostname()
}

// GetSSHKeyPath returns key path for use with ssh
func (d *synchronizedDriverDecorator) GetSSHKeyPath() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetSSHKeyPath()
}

// GetSSHPort returns port for use with ssh
func (d *synchronizedDriverDecorator) GetSSHPort() (int, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetSSHPort()
}

// GetSSHUsername returns username for use with ssh
func (d *synchronizedDriverDecorator) GetSSHUsername() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetSSHUsername()
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *synchronizedDriverDecorator) GetURL() (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetURL()
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *synchronizedDriverDecorator) GetState() (state.State, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.GetState()
}

// Kill stops a host forcefully
func (d *synchronizedDriverDecorator) Kill() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.Kill()
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *synchronizedDriverDecorator) PreCreateCheck() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.PreCreateCheck()
}

// Remove a host
func (d *synchronizedDriverDecorator) Remove() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.Remove()
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *synchronizedDriverDecorator) Restart() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.Restart()
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *synchronizedDriverDecorator) SetConfigFromFlags(opts DriverOptions) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.SetConfigFromFlags(opts)
}

// Start a host
func (d *synchronizedDriverDecorator) Start() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.Start()
}

// Stop a host gracefully
func (d *synchronizedDriverDecorator) Stop() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.Delegate.Stop()
}
