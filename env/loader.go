package env

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/google/logger"
	"github.com/peterbourgon/ff"
)

func Load() {
	fs := flag.NewFlagSet("demo-hello", flag.ExitOnError)
	var (
		name = fs.String("name", "hello-svc", "the name of the app (default is 'hello-svc')")
		port = fs.Int("port", 8484, "the listen port (default is '8484')")
		url  = fs.String("remote", "localhost:8485/hello", "the url of a remote service (default is 'another-svc:8485/hello')")
	)

	// Parse flags with or without a config file
	filename := os.Getenv("HELLO_CONFIG_LOCATION")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		ff.Parse(fs, os.Args[1:],
			ff.WithEnvVarPrefix("HELLO"))
	} else {
		logger.Infof("Using config filename: %s", filename)
		ff.Parse(fs, os.Args[1:],
			ff.WithConfigFile(filename),
			ff.WithConfigFileParser(PropertiesParser),
			ff.WithEnvVarPrefix("HELLO"))
	}

	NAME = *name
	PORT = *port
	REMOTE_URL = *url
}

// PlainParser is a parser for config files in an extremely simple format. Each
// line is tokenized as a single key/value pair. The first equal-delimited
// token in the line is interpreted as the flag name, and all remaining tokens
// are interpreted as the value. Any leading hyphens on the flag name are
// ignored.
func PropertiesParser(r io.Reader, set func(name, value string) error) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue // skip empties
		}

		if line[0] == '#' {
			continue // skip comments
		}

		var (
			name  string
			value string
			index = strings.IndexRune(line, '=')
		)
		if index > 0 {
			name, value = strings.TrimSpace(line[:index]), strings.TrimSpace(line[index+1:])
		}

		if i := strings.IndexRune(value, '#'); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		if err := set(name, value); err != nil {
			return err
		}
	}
	return nil
}
