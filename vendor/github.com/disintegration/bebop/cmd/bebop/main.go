// Bebop is a simple discussion board / forum web application.
// The bebop command is a command-line tool that manages the bebop web server.
package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	cmds := map[string]func(){
		"start":        startServer,
		"init":         initConfig,
		"gen-key":      genKey,
		"admins":       printAdmins,
		"add-admin":    addAdmin,
		"remove-admin": removeAdmin,
		"help":         printUsage,
	}

	if cmdFunc, ok := cmds[os.Args[1]]; ok {
		cmdFunc()
	} else {
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	log.Println(`usage:
	bebop start                      - start the server
	bebop init                       - create an initial configuration file
	bebop gen-key                    - generate a random 32-byte hex-encoded key
	bebop admins                     - show the admin list
	bebop add-admin <username>       - add a user to the admin list
	bebop remove-admin <username>    - remove a user from the admin list
	bebop help                       - show this message`)
}
