package drivers

type ContractFailure int

const (
	None ContractFailure = iota
	MachineDriverNotEligible
	RsyncMissingLocally
	RsyncMissingOnRemote
)
