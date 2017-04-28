package rpc

// CommandConfig configures a cli command
type CommandConfig struct {
	// Order is the listing order on cli help
	Order int

	// NumArgs is the # of arguments the command expects
	NumArgs int

	// Description is the cli help string
	Description string

	// RpcMethod is the method to call
	RpcMethod string
}

// RunOperationRpcCommand is the name of the Run Operation RPC command
const RunOperationRpcCommand = "RunOperation"

// Commands is the map of all cli commands. Key is the command name in cli.
var Commands = map[string]CommandConfig{
	"add":          {1, 1, "Adds a host to werifyd", "AddHost"},
	"del":          {2, 1, "Removes a host from werifyd", "RemoveHost"},
	"list":         {3, 0, "Lists hosts in werifyd", "ListHost"},
	"listactive":   {4, 0, "Lists active hosts in werifyd", "ListHost"},
	"listinactive": {5, 0, "Lists inactive hosts in werifyd", "ListHost"},
	"operation":    {6, 1, "Runs operations from file on werifyd", RunOperationRpcCommand},
	"get":          {7, 1, "Get status of operation with handle", "OperationStatusCheck"},
	"refresh":      {8, 0, "Start health check on all hosts", "Refresh"},
}
