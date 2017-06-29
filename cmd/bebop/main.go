// Bebop is a simple discussion board / forum web application.
// The bebop command is a command-line tool that manages the bebop web server.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	logger       = log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	useEnvConfig = flag.Bool("e", false, "use environment variables as config")
)

func main() {
	flag.Usage = help
	flag.Parse()

	cmds := map[string]func(){
		"start":        startServer,
		"init":         initConfig,
		"gen-key":      genKey,
		"admins":       printAdmins,
		"add-admin":    addAdmin,
		"remove-admin": removeAdmin,
		"help":         help,
	}

	if cmdFunc, ok := cmds[flag.Arg(0)]; ok {
		cmdFunc()
	} else {
		help()
		os.Exit(2)
	}
}

func help() {
	fmt.Fprintln(os.Stderr, `Usage:
	bebop start                      - start the server
	bebop init                       - create an initial configuration file
	bebop gen-key                    - generate a random 32-byte hex-encoded key
	bebop admins                     - show the admin list
	bebop add-admin <username>       - add a user to the admin list
	bebop remove-admin <username>    - remove a user from the admin list
	bebop help                       - show this message
Use -e flag to read configuration from environment variables instead of a file. E.g.:
	bebop -e start
	bebop -e admins`)
}
