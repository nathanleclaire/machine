package drivers

import (
	machinedrivers "github.com/docker/machine/drivers"
)

type Rsync struct {
	LocalDir  string
	RemoteDir string
	Machine   *machinedrivers.Driver
}

func (r Rsync) DriverName() string {
	return "rsync"
}

func (r Rsync) ContractFulfilled() (bool, ContractFailure, error) {
	return false, RsyncMissingLocally, nil
}

func (r Rsync) Create() error {
	return nil
}

func (r Rsync) Remove() error {
	return nil
}

func (r Rsync) Type() string {
	return ""
}

func (r Rsync) Push() error {
	return nil
}

func (r Rsync) Pull() error {
	return nil
}

func (r Rsync) Sync() error {
	return nil
}

func (r Rsync) OnHostStart() error {
	return nil
}

func (r Rsync) OnHostStop() error {
	return nil
}

func (r Rsync) OnHostRestart() error {
	return nil
}

func (r Rsync) OnHostKill() error {
	return nil
}

func (r Rsync) Provision() error {
	return nil
}
