// Package rpc contains the I/O signatures for the RPC calls
package rpc

// AddHostInput is the input struct for the add host functionality
type AddHostInput struct {
	CommonInput
	Endpoint Endpoint
}

// AddHostOutput is the output struct for the add host functionality
type AddHostOutput struct {
	Ok bool
}

// RemoveHostInput is the input struct for the remove host functionality
type RemoveHostInput struct {
	CommonInput
	Endpoint Endpoint
}

// RemoveHostOutput is the output struct for the remove host functionality
type RemoveHostOutput struct {
	Ok bool
}

// ListHostsInput is the input struct for the list hosts functionality
type ListHostsInput struct {
	CommonInput
	ListActive   bool
	ListInactive bool
}

// ListHostsOutput is the output struct for the list hosts functionality
type ListHostsOutput struct {
	ActiveHosts   []Endpoint
	InactiveHosts []Endpoint
}

// RefreshInput is the input struct for refresh hosts/start healthcheck functionality
type RefreshInput struct {
	CommonInput
}

// RefreshOutput is the output struct for refresh hosts/start healthcheck functionality
type RefreshOutput struct {
	Ok bool
}
