package host

import (
	"encoding/json"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
)

type RawHost struct {
	Driver *json.RawMessage
}

type HostV2 struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *HostOptions
	Name          string
}

func MigrateHostV2ToHostV3(hostV2 *HostV2, data []byte) *Host {
	// Migrate to include RawDriver so that driver plugin will work
	// smoothly.
	rawHost := &RawHost{}
	if err := json.Unmarshal(data, &rawHost); err != nil {
		log.Warn("Could not unmarshal raw host for RawDriver information: %s", err)
	}

	h := &Host{
		ConfigVersion: 2,
		DriverName:    hostV2.DriverName,
		Name:          hostV2.Name,
		HostOptions:   hostV2.HostOptions,
		RawDriver:     *rawHost.Driver,
	}

	return h
}
