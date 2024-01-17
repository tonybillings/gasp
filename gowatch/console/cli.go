package console

import (
	"errors"
	"flag"
	"os"
	"regexp"
	"strings"
)

var (
	Args CliArgs
)

var (
	socket  = flag.String("s", "localhost:8800", "socket (<address>:<port>) used for comm")
	rootDir = flag.String("r", ".", "root directory of the project, if not the current directory")
)

type CliArgs struct {
	Socket  string
	RootDir string
}

func Init() error {
	parseArgs()
	return validate()
}

func parseArgs() {
	flag.Parse()
	Args.Socket = *socket
	Args.RootDir = *rootDir
	if Args.RootDir != "/" && strings.HasSuffix(Args.RootDir, "/") {
		Args.RootDir = Args.RootDir[:len(Args.RootDir)-1]
	}
}

func validate() error {
	err := validateSocket(Args.Socket)
	if err != nil {
		return err
	}

	err = validateRootDir(Args.RootDir)
	if err != nil {
		return err
	}

	return nil
}

func validateSocket(socket string) error {
	socket = strings.Replace(Args.Socket, "localhost", "127.0.0.1", 1)
	socketMatch, _ := regexp.MatchString("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}:[0-9]{1,5}$", socket)
	if !socketMatch {
		return errors.New("invalid socket value passed to GoWatch (expected <address>:<port>)")
	}
	return nil
}

func validateRootDir(rootDir string) error {
	if rootDir == "." {
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if wd != "/" && strings.HasSuffix(wd, "/") {
		wd = wd[:len(wd)-1]
	}

	if strings.HasPrefix(rootDir, "./") || (strings.HasPrefix(rootDir, wd) && len(rootDir) > len(wd)) {
		return errors.New("the root directory cannot be a sub-directory of the current working directory")
	}

	return nil
}
