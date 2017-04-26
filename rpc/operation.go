package rpc

type OperationType string
type OperationArgument string

// Operation is a single operation
type Operation struct {
	OpType   OperationType     `json:"type"`
	PathArg  OperationArgument `json:"path,omitempty"`
	CheckArg OperationArgument `json:"check,omitempty"`
}

// OperationResult is a result of a single operation
type OperationResult struct {
	Success bool
	Err     error
}

// Type OperationInput is the input struct for the operation functionality
type OperationInput struct {
	CommonInput

	// Forward determines if we are running these operations on the currect context or forwarding them down to other hosts
	Forward bool

	// Ops is a map of operations, map key is the given unique name
	Ops map[string]Operation
}

// Type OperationOutput is the output struct for the operation functionality
type OperationOutput struct {
	// Results is a map of results per given unique name per server identifier
	Results map[ServerIdentifier]map[string]OperationResult
}

// WorkerOperation is Operation with unique name attached, for the worker pool
type WorkerOperation struct {
	Name string
	Operation
}

// WorkerOperationResult is OperationResult with unique name attached, for the worker pool
type WorkerOperationResult struct {
	Name string
	OperationResult
}
