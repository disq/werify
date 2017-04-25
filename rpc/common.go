package rpc

// ProtoVersion is poor man's protocol versioning system
const ProtoVersion = "werify.v1"

// type CommonInput is included in all RPC inputs
type CommonInput struct {
	EnvTag string
}

func BuildMethod(m string) string {
	return ProtoVersion + "." + m
}
