package share

type ContractFailure int

const (
	None state = iota
	MachineDriverNotEligible
	RsyncMissingLocally
	RsyncMissingOnRemote
)
