package rpc

// ProtoVersion is poor man's protocol versioning system
const ProtoVersion = "werify.v1"

// CommonInput is included in all RPC inputs
type CommonInput struct {
	EnvTag string
}

// BuildMethod prepends the ProtoVersion to the rpc method name
func BuildMethod(rpcMethod string) string {
	return ProtoVersion + "." + rpcMethod
}
