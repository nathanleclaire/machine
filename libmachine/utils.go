package libmachine

import (
	"path/filepath"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/utils"
)

// validates host config and modifies if needed
// this is used for configuration updates
func validateHost(host *Host) *Host {
	certInfo := getCertInfoFromHost(host)

	if host.HostConfig == nil {
		host.HostConfig = &HostOptions{}
	}
	if host.HostConfig.EngineConfig == nil {
		host.HostConfig.EngineConfig = &engine.EngineOptions{}
	}

	if host.HostConfig.SwarmConfig == nil {
		host.HostConfig.SwarmConfig = &swarm.SwarmOptions{
			Address:   "",
			Discovery: host.SwarmDiscovery,
			Host:      host.SwarmHost,
			Master:    host.SwarmMaster,
		}
	}

	if host.HostConfig.AuthConfig == nil {
		host.HostConfig.AuthConfig = &auth.AuthOptions{
			StorePath:            host.StorePath,
			CaCertPath:           certInfo.CaCertPath,
			CaCertRemotePath:     "",
			ServerCertPath:       certInfo.ServerCertPath,
			ServerKeyPath:        certInfo.ServerKeyPath,
			ClientKeyPath:        certInfo.ClientKeyPath,
			ServerCertRemotePath: "",
			ServerKeyRemotePath:  "",
			PrivateKeyPath:       certInfo.CaKeyPath,
			ClientCertPath:       certInfo.ClientCertPath,
		}
	}

	return host
}

// validates host metadata and modifies if needed
// this is used for configuration updates
func validateHostMetadata(m *HostMetadata) *HostMetadata {
	if m.HostConfig.EngineConfig == nil {
		m.HostConfig.EngineConfig = &engine.EngineOptions{}
	}

	if m.HostConfig.AuthConfig == nil {
		m.HostConfig.AuthConfig = &auth.AuthOptions{
			StorePath:            m.StorePath,
			CaCertPath:           m.CaCertPath,
			CaCertRemotePath:     "",
			ServerCertPath:       m.ServerCertPath,
			ServerKeyPath:        m.ServerKeyPath,
			ClientKeyPath:        "",
			ServerCertRemotePath: "",
			ServerKeyRemotePath:  "",
			PrivateKeyPath:       m.PrivateKeyPath,
			ClientCertPath:       m.ClientCertPath,
		}
	}

	return m
}

func getCertInfoFromHost(h *Host) CertPathInfo {
	// setup cert paths
	caCertPath := h.CaCertPath
	caKeyPath := h.PrivateKeyPath
	clientCertPath := h.ClientCertPath
	clientKeyPath := h.ClientKeyPath

	if caCertPath == "" {
		caCertPath = filepath.Join(utils.GetMachineCertDir(), "ca.pem")
	}

	if caKeyPath == "" {
		caKeyPath = filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	}

	if clientCertPath == "" {
		clientCertPath = filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	}

	if clientKeyPath == "" {
		clientKeyPath = filepath.Join(utils.GetMachineCertDir(), "key.pem")
	}

	return CertPathInfo{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
	}
}
