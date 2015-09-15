package rpcdriver

import (
	"encoding/json"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/version"
)

// TODO: Do not shut down the server completely in the case of a panic.

type RpcFlags struct {
	Values map[string]interface{}
}

func (r RpcFlags) String(key string) string {
	val, ok := r.Values[key].(string)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to string for key %s", key)
	}
	return val
}

func (r RpcFlags) StringSlice(key string) []string {
	val, ok := r.Values[key].([]string)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to string slice for key %s", key)
	}
	return val
}

func (r RpcFlags) Int(key string) int {
	val, ok := r.Values[key].(int)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to int for key %s", key)
	}
	return val
}

func (r RpcFlags) Bool(key string) bool {
	val, ok := r.Values[key].(bool)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to bool for key %s", key)
	}
	return val
}

type RpcServerDriver struct {
	ActualDriver drivers.Driver
	CloseCh      chan bool
}

// Gets slapped on all methods which call actual driver functions, because
// there if there is a panic it will shut down the server and FUBAR all future
// incoming requests.
func Recover() {
	if r := recover(); r != nil {
		log.Errorf("Panic! at the Goroutine: Recovered. Details: %s", r)
	}
}

func (r *RpcServerDriver) Close(_, _ *struct{}) error {
	r.CloseCh <- true
	return nil
}

func (r *RpcServerDriver) GetVersion(_ *struct{}, reply *int) error {
	*reply = version.ApiVersion
	return nil
}

func (r *RpcServerDriver) GetConfigRaw(_ *struct{}, reply *[]byte) error {
	driverData, err := json.Marshal(r.ActualDriver)
	if err != nil {
		return err
	}

	*reply = driverData

	return nil
}

func (r *RpcServerDriver) GetCreateFlags(_ *struct{}, reply *[]mcnflag.Flag) error {
	defer Recover()
	*reply = r.ActualDriver.GetCreateFlags()
	return nil
}

func (r *RpcServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	return json.Unmarshal(data, &r.ActualDriver)
}

func (r *RpcServerDriver) Create(_, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.Create()
}

func (r *RpcServerDriver) DriverName(_ *struct{}, reply *string) error {
	defer Recover()
	*reply = r.ActualDriver.DriverName()
	return nil
}

func (r *RpcServerDriver) GetIP(_ *struct{}, reply *string) error {
	defer Recover()
	ip, err := r.ActualDriver.GetIP()
	*reply = ip
	return err
}

func (r *RpcServerDriver) GetMachineName(_ *struct{}, reply *string) error {
	defer Recover()
	*reply = r.ActualDriver.GetMachineName()
	return nil
}

func (r *RpcServerDriver) GetSSHHostname(_ *struct{}, reply *string) error {
	defer Recover()
	hostname, err := r.ActualDriver.GetSSHHostname()
	*reply = hostname
	return err
}

func (r *RpcServerDriver) GetSSHKeyPath(_ *struct{}, reply *string) error {
	defer Recover()
	*reply = r.ActualDriver.GetSSHKeyPath()
	return nil
}

// GetSSHPort returns port for use with ssh
func (r *RpcServerDriver) GetSSHPort(_ *struct{}, reply *int) error {
	defer Recover()
	port, err := r.ActualDriver.GetSSHPort()
	*reply = port
	return err
}

func (r *RpcServerDriver) GetSSHUsername(_ *struct{}, reply *string) error {
	defer Recover()
	*reply = r.ActualDriver.GetSSHUsername()
	return nil
}

func (r *RpcServerDriver) GetURL(_ *struct{}, reply *string) error {
	defer Recover()
	info, err := r.ActualDriver.GetURL()
	*reply = info
	return err
}

func (r *RpcServerDriver) GetState(_ *struct{}, reply *state.State) error {
	defer Recover()
	s, err := r.ActualDriver.GetState()
	*reply = s
	return err
}

func (r *RpcServerDriver) Kill(_ *struct{}, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.Kill()
}

func (r *RpcServerDriver) LocalArtifactPath(file *string, reply *string) error {
	defer Recover()
	*reply = r.ActualDriver.LocalArtifactPath(*file)
	return nil
}

func (r *RpcServerDriver) GlobalArtifactPath(_ *struct{}, reply *string) error {
	defer Recover()
	*reply = r.ActualDriver.GlobalArtifactPath()
	return nil
}

func (r *RpcServerDriver) PreCreateCheck(_ *struct{}, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.PreCreateCheck()
}

func (r *RpcServerDriver) Remove(_ *struct{}, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.Remove()
}

func (r *RpcServerDriver) Restart(_ *struct{}, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.Restart()
}

func (r *RpcServerDriver) SetConfigFromFlags(flags *drivers.DriverOptions, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.SetConfigFromFlags(*flags)
}

func (r *RpcServerDriver) Start(_ *struct{}, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.Start()
}

func (r *RpcServerDriver) Stop(_ *struct{}, _ *struct{}) error {
	defer Recover()
	return r.ActualDriver.Stop()
}
