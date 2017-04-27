package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/disq/werify"
	wrpc "github.com/disq/werify/rpc"
)

var errorNop = errors.New("No operation was performed")

const defaultTimeoutClientToServer = 10 * time.Second

// fail prints the error and then aborts the program
func fail(err error, command *string) {
	if command != nil {
		fmt.Printf("Error running %s: %s\n", *command, err.Error())
	} else {
		fmt.Printf("Error: %s\n", err.Error())
	}
	os.Exit(1)
}

func printUsageLine() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [COMMAND [PARAMS...]]\n\nAvailable options:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nAvailable commands:\n")

	// Sort commands by wrpc.Commands.Order
	orders := make([]int, 0)
	byOrder := make(map[int][]string)
	for k, v := range wrpc.Commands {
		byOrder[v.Order] = append(byOrder[v.Order], k)
	}
	for k := range byOrder {
		orders = append(orders, k)
	}
	sort.Sort(sort.IntSlice(orders))

	// Iterate over Order values and over each value with the same Order, printing the command and description
	for _, o := range orders {
		for _, c := range byOrder[o] {
			fmt.Fprintf(os.Stderr, "  %15s  %s\n", c, wrpc.Commands[c].Description)
		}
	}

	fmt.Fprintf(os.Stderr, "\nCommands can also be specified from stdin using \"-\".\n")
}

func main() {
	env := flag.String("env", werify.DefaultEnv, "Env tag")
	connect := flag.String("connect", fmt.Sprintf("localhost:%d", werify.DefaultPort), "Connect to werifyd")
	timeout := flag.Duration("timeout", defaultTimeoutClientToServer, "Connect timeout")

	flag.Usage = printUsageLine
	flag.Parse()

	_, cancelFunc := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGPIPE)
		<-ch
		//fmt.Println("Got signal, cleaning up...")
		cancelFunc()
	}()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	c := &client{
		env:     *env,
		server:  *connect,
		timeout: *timeout,
	}

	err := c.connect()
	if err != nil {
		fail(err, nil)
	}

	if flag.Arg(0) == "-" {
		parseArgsFromFile(c, os.Stdin)
	} else {
		err := parseArgs(c, flag.Args())
		if err != nil {
			fail(err, nil)
		}
	}
}

func parseArgsFromFile(c *client, f *os.File) {
	processed := 0
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		args := strings.Split(line, " ")
		err := parseArgs(c, args)
		if err != nil && err != errorNop {
			fail(err, &line)
		}
		if err != errorNop {
			processed++
		}
	}
	if err := s.Err(); err != nil {
		fail(err, nil)
	}
	fmt.Printf("Commands processed: %d\n", processed)
}

func parseArgs(c *client, args []string) error {
	if len(args) == 0 {
		return errors.New("Empty command specified")
	}
	command := strings.TrimSpace(args[0])
	if command == "" || command[0] == '#' {
		return errorNop
	}

	return c.parseCommand(command, args[1:])
}
