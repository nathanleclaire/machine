package swarm

const (
	DockerImage              = "swarm:latest"
	DiscoveryServiceEndpoint = "https://discovery-stage.hub.docker.com/v1"
)

type SwarmOptions struct {
	IsSwarm   bool
	Discovery string
	Master    bool
	Host      string
	Addr      string
}
