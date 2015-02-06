package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
				"Debug":         true,
				"StoragePath":   filepath.Join(utils.GetMachineDir()),
				"TlsCaCert":     filepath.Join(utils.GetMachineDir(), "ca.pem"),
				"TlsCaKey":      filepath.Join(utils.GetMachineDir(), "key.pem"),
				"TlsClientCert": filepath.Join(utils.GetMachineClientCertDir(), "cert.pem"),
				"TlsClientKey":  filepath.Join(utils.GetMachineClientCertDir(), "key.pem"),
			},
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
	if json.Unmarshal(data, c); err != nil {
		log.Fatalf("Error unmarshalling %s: %s", c.configFilePath, err)
	}
	return nil
}

func (c *ClientConfigStore) Save() error {
	log.Info("Saving file!")
	log.Info(c)
	b, err := json.MarshalIndent(c.configValues, "", "\t")
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
	key string
}

func (e ErrInvalidType) Error() string {
	return fmt.Sprintf("Invalid type for attempted key access at key: %s", e.key)
}

func (c *ClientConfigStore) get(key string) (interface{}, error) {
	var (
		ok  bool
		val interface{}
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

func (c *ClientConfigStore) GetBool(key string) (bool, error) {
	var (
		assertedVal bool
		ok          bool
	)
	val, err := c.get(key)
	if err != nil {
		return false, err
	}
	assertedVal, ok = val.(bool)
	if !ok {
		return false, ErrInvalidType{key: key}
	}
	return assertedVal, nil
}

func (c *ClientConfigStore) GetString(key string) (string, error) {
	var (
		assertedVal string
		ok          bool
	)
	val, err := c.get(key)
	if err != nil {
		return "", err
	}
	assertedVal, ok = val.(string)
	if !ok {
		return "", ErrInvalidType{key: key}
	}
	return assertedVal, nil
}

func (c *ClientConfigStore) GetInt(key string) (int, error) {
	var (
		assertedVal int
		ok          bool
	)
	val, err := c.get(key)
	if err != nil {
		return 0, err
	}
	assertedVal, ok = val.(int)
	if !ok {
		return 0, ErrInvalidType{key: key}
	}
	return assertedVal, nil
}

func (c *ClientConfigStore) Set(key string, value interface{}) error {
	var (
		nestedVal map[string]interface{}
		ok        bool
	)
	keys := strings.Split(key, ".")
	nestedVal = c.configValues[keys[0]].(map[string]interface{})
	for i, nestedKey := range keys {
		// on last element of array, set the value
		// for that key
		if i+1 == len(keys) {
			nestedVal[nestedKey] = value
			// done
			return nil
		}
		nestedVal, ok = nestedVal[nestedKey].(map[string]interface{})
		if !ok {
			return ErrKeyNotFound{key: key}
		}
		fmt.Println(nestedVal)
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
