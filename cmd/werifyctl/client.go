package main

import (
	"errors"
	"fmt"
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
		return fmt.Errorf("Unhandled command %s", command)
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
	default:
		return errors.New("Unhandled command")
	}

	return nil
}

// newCommonInput initializes and returns a CommonInput struct using client's information
func (c *client) newCommonInput() wrpc.CommonInput {
	return wrpc.CommonInput{
		EnvTag: c.env,
	}
}
