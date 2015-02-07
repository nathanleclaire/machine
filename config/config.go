package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/utils"
)

const (
	clientConfigFile = "machine-client-config.json"
)

type Config interface {
	Load(file string) error
	Save() error
}

type ConfigHierarchy struct {
	EnvVar  string
	JsonKey string
	Default interface{}
}

type ClientConfigStore struct {
	configFilePath string
	configValues   map[string]interface{}
}

func NewClientConfigStore() (ClientConfigStore, error) {
	// defaults
	c := ClientConfigStore{
		configFilePath: filepath.Join(utils.GetDockerDir(), clientConfigFile),
		configValues: map[string]interface{}{
			"Core": map[string]interface{}{
				"Debug":         "false",
				"StoragePath":   filepath.Join(utils.GetMachineDir()),
				"TlsCaCert":     filepath.Join(utils.GetMachineDir(), "ca.pem"),
				"TlsCaKey":      filepath.Join(utils.GetMachineDir(), "key.pem"),
				"TlsClientCert": filepath.Join(utils.GetMachineClientCertDir(), "cert.pem"),
				"TlsClientKey":  filepath.Join(utils.GetMachineClientCertDir(), "key.pem"),
			},
			// TODO: This type is ugly, but I can't think of a way to implement it that isn't.
			//       We could use a struct and reflection, but in considering that it seems more trouble
			//       than it's worth.
			"Drivers": map[string]map[string]interface{}{
				"DigitalOcean": map[string]interface{}{
					"AccessToken": "",
				},
			},
		},
	}
	if err := c.Load(); err != nil {
		return c, err
	}
	return c, nil
}

func (c *ClientConfigStore) Load() error {
	if _, err := os.Stat(c.configFilePath); os.IsNotExist(err) {
		// not a problem - just an empty config
		return nil
	}
	data, err := ioutil.ReadFile(c.configFilePath)
	if err != nil {
		switch err {
		case os.ErrPermission:
			log.Fatalf("Error reading %s: Insufficient permissions", c.configFilePath)
		default:
			log.Fatalf("Unrecognized error reading %s: %s", c.configFilePath, err)
		}
	}
	if json.Unmarshal(data, &c.configValues); err != nil {
		log.Fatalf("Error unmarshalling %s: %s", c.configFilePath, err)
	}
	return nil
}

func (c *ClientConfigStore) Save() error {
	b, err := json.MarshalIndent(c.configValues, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(c.configFilePath, b, 0600)
	if err != nil {
		return err
	}
	return nil
}

type ErrKeyNotFound struct {
	key string
}

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("Key not found: %s", e.key)
}

type ErrInvalidType struct {
	key             string
	conversionError error
}

func (e ErrInvalidType) Error() string {
	return fmt.Sprintf("Invalid type for attempted key access at key: %s\nError: %s", e.key, e.conversionError)
}

type ErrKeyNotSupported struct {
	key string
}

func (e ErrKeyNotSupported) Error() string {
	return fmt.Sprintf("Tried to set a key which is not supported: %s", e.key)
}

func (c *ClientConfigStore) get(key string) (interface{}, error) {
	var (
		val interface{}
		ok  bool
	)
	nestedVal := c.configValues
	keys := strings.Split(key, ".")
	for i, nestedKey := range keys {
		if i+1 == len(keys) {
			val, ok = nestedVal[nestedKey]
			if !ok {
				return nil, ErrKeyNotFound{key: key}
			}
			break
		}
		nestedVal, ok = nestedVal[nestedKey].(map[string]interface{})
		if !ok {
			return nil, ErrKeyNotFound{key: key}
		}
	}
	return val, nil
}

func (c *ClientConfigStore) GetString(key string) (string, error) {
	val, err := c.get(key)
	if err != nil {
		return "", err
	}
	assertedVal, ok := val.(string)
	if !ok {
		return "", ErrInvalidType{key: key}
	}
	return assertedVal, nil
}

func (c *ClientConfigStore) GetBool(key string) (bool, error) {
	val, err := c.GetString(key)
	if err != nil {
		return false, err
	}
	convertedVal, err := strconv.ParseBool(val)
	if err != nil {
		return false, ErrInvalidType{
			key:             key,
			conversionError: err,
		}
	}
	return convertedVal, nil
}

func (c *ClientConfigStore) GetInt(key string) (int64, error) {
	val, err := c.GetString(key)
	if err != nil {
		return 0, err
	}
	convertedVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, ErrInvalidType{
			key:             key,
			conversionError: err,
		}
	}
	return convertedVal, nil
}

func (c *ClientConfigStore) Set(key string, value interface{}) error {
	var (
		nestedVal map[string]interface{}
		ok        bool
	)
	keys := strings.Split(key, ".")
	nestedVal = c.configValues
	for i, nestedKey := range keys {
		// on last element of array, set the value
		// for that key
		if i+1 == len(keys) {
			// Check for existence of key.
			// All supported keys should be present by default.
			_, ok = nestedVal[nestedKey]
			if !ok {
				return ErrKeyNotSupported{key: key}
			}
			nestedVal[nestedKey] = value
			// done
			return nil
		}
		nestedVal, ok = nestedVal[nestedKey].(map[string]interface{})
		if !ok {
			return ErrKeyNotFound{key: key}
		}
	}
	return ErrKeyNotFound{key: key}
}

// The hierarchy of configuration options flows like
// this, in order of most preferred to least preferred:
//
// Command Line Flags => Environment Variables => defaults.json => Hardcoded Defaults
func GetConfigValue(key string) interface{} {
	return nil
}
