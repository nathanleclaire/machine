package share

import "github.com/codegangsta/cli"

type ShareDriver interface {
	// In order to mount a share, a "contract" must be fulfilled.
	// For instance, to mount a VirtualBox shared folder the Guest
	// Additions must be installed in the guest machine (and it must
	// be running using the VirtualBox driver), to use a rsync share
	// the rsync binary must be present on the guest _and_ the host,
	// and so on.
	ContractFulfilled() (bool, ContractFailure, error)
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

type RegisteredShareDriver struct {
	New            func(storePath string) (Driver, error)
	GetCreateFlags func() []cli.Flag
}

type Share interface {
	LocalPath() string
	RemotePath() string
}
