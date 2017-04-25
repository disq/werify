package rpc

// Type SetIdentifierInput is the input struct for the set identifier functionality
type SetIdentifierInput struct {
	CommonInput
	Identifier string
}

// Type SetIdentifierOutput is the output struct for the set identifier functionality
type SetIdentifierOutput struct {
	Ok bool
}
