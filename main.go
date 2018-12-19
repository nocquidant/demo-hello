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
	"github.com/nocquidant/demo-hello/parse"
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

	// We may use a config file...
	filename := os.Getenv("HELLO_CONFIG_LOCATION")
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		logger.Infof("Using config filename: %s", filename)
		ff.Parse(fs, os.Args[1:], 
			ff.WithConfigFile(filename),
			ff.WithConfigFileParser(parse.PropertiesParser),
			ff.WithEnvVarPrefix("HELLO"))
	} else {
		ff.Parse(fs, os.Args[1:], 
			ff.WithEnvVarPrefix("HELLO"))
	}

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
	logger.Info("Available GET endpoints are: '/health', '/hello' and '/remote'")

	mux := http.NewServeMux()
	mux.HandleFunc("/", api.HandlerHealth)
	mux.HandleFunc("/health", api.HandlerHealth)
	mux.HandleFunc("/hello", api.HandlerHello)
	mux.HandleFunc("/remote", api.HandlerRemote)

	http.ListenAndServe(":"+strconv.Itoa(env.PORT), mux)
}
