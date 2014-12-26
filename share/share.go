package share

import (
	machinedrivers "github.com/docker/machine/drivers"
	"github.com/docker/machine/share/drivers"
)

type ShareDriver interface {
	// In order to mount a share, a "contract" must be fulfilled.
	// For instance, to mount a VirtualBox shared folder the Guest
	// Additions must be installed in the guest machine (and it must
	// be running using the VirtualBox driver), to use a rsync share
	// the rsync binary must be present on the guest _and_ the host,
	// and so on.
	ContractFulfilled() (bool, drivers.ContractFailure, error)
	Create() error
	Remove() error
	DriverName() string
	Push() error
	Pull() error
	Sync() error
	OnHostStart() error
	OnHostStop() error
	OnHostRestart() error
	OnHostKill() error

	// todo: just expect provisioning through cloudinit?
	Provision() error
}

func ListShares() error {
	return nil
}

func GetShare(localDir string, d machinedrivers.Driver) (ShareDriver, error) {
	return drivers.Rsync{}, nil
}

func NewShare(localDir string, driverType string) error {
	return nil
}
