package rsync

import (
	_ "github.com/docker/machine/share"
)

func ContractFulfilled() (bool, ContractFailure, error) {
	return false, RsyncMissingLocally, nil
}

func Create() error {
	return nil
}

func Remove() error {
	return nil
}

func Type() string {
	return ""
}

func Push() error {
	return nil
}

func Pull() error {
	return nil
}

func Sync() error {
	return nil
}

func OnHostStart() error {
	return nil
}

func OnHostStop() error {
	return nil
}

func OnHostRestart() error {
	return nil
}

func OnHostKill() error {
	return nil
}

func Provision() error {
	return nil
}
