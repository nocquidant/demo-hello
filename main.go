package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/google/logger"
	"github.com/nocquidant/demo-hello/api"
	"github.com/nocquidant/demo-hello/env"
	"github.com/peterbourgon/ff"
	"github.com/satori/go.uuid"
)

func main() {
	logger.Init("hello", true, false, ioutil.Discard)

	fs := flag.NewFlagSet("demo-hello", flag.ExitOnError)
	var (
		name = fs.String("name", "hello-svc", "the name of the app (default is 'hello-svc')")
		port = fs.Int("port", 8484, "the listen port (default is '8484')")
		url  = fs.String("remote", "localhost:8485/hello", "the url of a remote service (default is 'another-svc:8485/hello')")
	)

	// Use env variable like 'HELLO_NAME=dummy'
	ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("HELLO"))

	env.NAME = *name
	env.PORT = *port
	env.REMOTE_URL = *url

	// Set a UUID for the running instance
	env.INSTANCE_ID = uuid.NewV4().String()

	logger.Info("Environment used...")
	logger.Infof(" - env.version: %s\n", env.VERSION)
	logger.Infof(" - env.build: %s\n", env.GITCOMMIT)
	logger.Infof(" - env.name: %s\n", env.NAME)
	logger.Infof(" - env.port: %d\n", env.PORT)
	logger.Infof(" - env.remoteUrl: %s\n", env.REMOTE_URL)
	logger.Infof(" - env.instanceId: %s\n", env.INSTANCE_ID)

	logger.Infof("HTTP service: %s, is running using port: %d\n", env.NAME, env.PORT)
	logger.Info("Available GET endpoints are: '/health', '/hello', '/info' and '/request'")

	http.HandleFunc("/", api.HandlerHealth)
	http.HandleFunc("/health", api.HandlerHealth)
	http.HandleFunc("/hello", api.HandlerHello)
	http.HandleFunc("/remote", api.HandlerRemote)

	http.ListenAndServe(":"+strconv.Itoa(env.PORT), nil)
}
