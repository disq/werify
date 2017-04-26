package rpc

// ServerIdentifier is actually an Endpoint in a different context.
type ServerIdentifier string

// Type SetIdentifierInput is the input struct for the set identifier functionality
type SetIdentifierInput struct {
	CommonInput
	Identifier ServerIdentifier
}

// Type SetIdentifierOutput is the output struct for the set identifier functionality
type SetIdentifierOutput struct {
	Ok bool
}
