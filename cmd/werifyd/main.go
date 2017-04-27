package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/disq/werify"
	wrpc "github.com/disq/werify/rpc"
)

func main() {
	env := flag.String("env", werify.DefaultEnv, "Env tag")
	port := flag.Int("port", werify.DefaultPort, "Listen on port")
	numWorkers := flag.Int("w", runtime.NumCPU(), "Number of workers per operation")

	flag.Parse()

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGPIPE)
		<-ch
		log.Print("Got signal, cleaning up...")
		cancelFunc()
	}()

	rand.Seed(time.Now().UnixNano())

	s := &Server{
		context:    ctx,
		env:        *env,
		numWorkers: *numWorkers,
		opBuffer:   make(map[string]wrpc.OperationOutput),
	}

	err := rpc.RegisterName(wrpc.ProtoVersion, s)
	if err != nil {
		log.Fatalf("Registering RPC server: %s", err.Error())
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Could not bind: %d", err.Error())
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	rpc.Accept(listener)
}
