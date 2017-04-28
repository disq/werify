package rpc

import "time"

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
	// Err is the error value as a primitive
	Err string
}

// OperationInput is the input struct for the operation functionality
type OperationInput struct {
	CommonInput

	// Forward determines if we are running these operations on the currect context or forwarding them down to other hosts.
	// This also makes the current call async, in OperationOutput only Id will be returned.
	Forward bool

	// Ops is a map of operations, map key is the given unique name
	Ops map[string]Operation
}

// OperationOutput is the output struct for the operation functionality
type OperationOutput struct {
	// Handle is a unique id to check the results using OperationStatusCheckInput
	Handle string

	// Results is a map of results per given unique name per server identifier
	Results map[ServerIdentifier]map[string]OperationResult

	// StartedAt is the start time
	StartedAt time.Time

	// EndedAt shows if the operation is still running or ended
	EndedAt *time.Time
}

// OperationStatusCheckInput is the input struct to check status of an operation
type OperationStatusCheckInput struct {
	CommonInput

	// Handle is the unique id of the operation to check results for
	Handle string
}

// OperationStatusCheckOutput is the output struct for the operation functionality
type OperationStatusCheckOutput OperationOutput
