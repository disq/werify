package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

	log.Printf("Connected to %s", c.server)

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
		log.Printf("Result: %t", out.Ok)

	case "del":
		out := wrpc.RemoveHostOutput{}
		err := c.conn.Call(rpcCmd, wrpc.RemoveHostInput{CommonInput: ci, Endpoint: wrpc.Endpoint(args[0])}, &out)
		if err != nil {
			return err
		}
		log.Printf("Result: %t", out.Ok)

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
			log.Printf("Active hosts (%d)", len(out.ActiveHosts))
			for _, e := range out.ActiveHosts {
				log.Print(e)
			}
		}
		if command == "list" || command == "listinactive" {
			log.Printf("Inactive hosts (%d)", len(out.InactiveHosts))
			for _, e := range out.InactiveHosts {
				log.Print(e)
			}
		}
		log.Print("end of list")

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

		// TODO: make it async? Get an identifier, run another command to read so-far-collected results and status
		err = c.conn.Call(rpcCmd, in, &out)
		if err != nil {
			return err
		}

		if out.Handle != "" {
			log.Printf("Operation submitted, the handle is %s. Run ./werifyctl get %s to check progress.", out.Handle, out.Handle)
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
				log.Printf("Host:%s Operation:%s Error:%s", id, name, result.Err)
			} else {
				log.Printf("Host:%s Operation:%s Success:%t", id, name, result.Success)
			}
		}
	}

	if o.EndedAt != nil {
		log.Printf("Operation ended, took %v", o.EndedAt.Sub(o.StartedAt))
	} else {
		log.Printf("Operation still running...")
	}
}
