package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"time"

	wrpc "github.com/disq/werify/rpc"
)

type client struct {
	env     string
	server  string
	timeout time.Duration

	conn *rpc.Client
}

func (c *client) connect() error {
	connection, err := net.DialTimeout("tcp", c.server, c.timeout)
	if err != nil {
		return err
	}

	//fmt.Printf("Connected to %s\n", c.server)

	c.conn = rpc.NewClient(connection)
	return nil
}

func (c *client) parseCommand(command string, args []string) error {
	cmdCfg, ok := wrpc.Commands[command]
	if !ok {
		return fmt.Errorf("Unknown command %s", command)
	}
	if cmdCfg.NumArgs != len(args) {
		return fmt.Errorf("Invalid number of arguments for %s: Expected %d but got %d", command, cmdCfg.NumArgs, len(args))
	}

	rpcCmd := wrpc.BuildMethod(cmdCfg.RpcMethod)
	ci := c.newCommonInput()

	switch command {
	case "add":
		out := wrpc.AddHostOutput{}
		err := c.conn.Call(rpcCmd, wrpc.AddHostInput{CommonInput: ci, Endpoint: wrpc.Endpoint(args[0])}, &out)
		if err != nil {
			return err
		}
		if out.Ok {
			fmt.Printf("Added host %s\n", args[0])
		} else {
			fmt.Printf("Could not add host %s\n", args[0])
		}

	case "del":
		out := wrpc.RemoveHostOutput{}
		err := c.conn.Call(rpcCmd, wrpc.RemoveHostInput{CommonInput: ci, Endpoint: wrpc.Endpoint(args[0])}, &out)
		if err != nil {
			return err
		}
		if out.Ok {
			fmt.Printf("Removed host %s\n", args[0])
		} else {
			fmt.Printf("Could not remove host %s\n", args[0])
		}

	case "refresh":
		out := wrpc.RefreshOutput{}
		err := c.conn.Call(rpcCmd, wrpc.RefreshInput{CommonInput: ci}, &out)
		if err != nil {
			return err
		}
		if out.Ok {
			fmt.Println("Initiated refresh")
		} else {
			fmt.Println("Could not initiate refresh")
		}

	// FIXME: repeating ugly code below
	case "list":
		fallthrough
	case "listactive":
		fallthrough
	case "listinactive":
		out := wrpc.ListHostsOutput{}
		in := wrpc.ListHostsInput{CommonInput: ci, ListActive: true, ListInactive: true}
		if command == "listactive" {
			in.ListInactive = false
		} else if command == "listinactive" {
			in.ListActive = false
		}
		err := c.conn.Call(rpcCmd, in, &out)
		if err != nil {
			return err
		}

		if command == "list" || command == "listactive" {
			fmt.Printf("Active hosts (%d)\n", len(out.ActiveHosts))
			for _, e := range out.ActiveHosts {
				fmt.Println(e)
			}
		}
		if command == "list" || command == "listinactive" {
			fmt.Printf("Inactive hosts (%d)\n", len(out.InactiveHosts))
			for _, e := range out.InactiveHosts {
				fmt.Println(e)
			}
		}
		fmt.Println("End of list")

	case "operation":
		b, err := ioutil.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("Reading %s: %s", args[0], err.Error())
		}

		in := wrpc.OperationInput{
			CommonInput: ci,
			Forward:     true,
		}

		err = json.Unmarshal(b, &in.Ops)
		if err != nil {
			return fmt.Errorf("Parsing %s: %s", args[0], err.Error())
		}

		out := wrpc.OperationOutput{}

		err = c.conn.Call(rpcCmd, in, &out)
		if err != nil {
			return err
		}

		if out.Handle != "" {
			fmt.Printf("Operation submitted. To check progress, run: ./werifyctl get %s\n", out.Handle)
		} else {
			c.displayOperation(out)
		}

	case "get":
		out := wrpc.OperationStatusCheckOutput{}
		err := c.conn.Call(rpcCmd, wrpc.OperationStatusCheckInput{CommonInput: ci, Handle: args[0]}, &out)
		if err != nil {
			return err
		}
		c.displayOperation(wrpc.OperationOutput(out))

	default:
		return fmt.Errorf("Unhandled command %s", command)
	}

	return nil
}

// newCommonInput initializes and returns a CommonInput struct using client's information
func (c *client) newCommonInput() wrpc.CommonInput {
	return wrpc.CommonInput{
		EnvTag: c.env,
	}
}

func (c *client) displayOperation(o wrpc.OperationOutput) {
	for id, res := range o.Results {
		for name, result := range res {
			if result.Err != "" {
				fmt.Printf("Host:%s Operation:%s Error:%s\n", id, name, result.Err)
			} else {
				fmt.Printf("Host:%s Operation:%s Success:%t\n", id, name, result.Success)
			}
		}
	}

	if o.EndedAt != nil {
		fmt.Printf("Operation ended, took %v\n", o.EndedAt.Sub(o.StartedAt))
	} else {
		fmt.Printf("Operation still running...\n")
	}
}
