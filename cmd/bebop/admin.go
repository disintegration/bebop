package main

import (
	"log"
	"os"

	"github.com/disintegration/bebop/config"
	"github.com/disintegration/bebop/store"
)

// printAdmins prints all the administrator users.
func printAdmins() {
	cfg, err := config.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to load the configuration file: %s", err)
	}

	s, err := getStore(cfg)
	if err != nil {
		log.Fatalf("failed to get data store: %s", err)
	}

	admins, err := s.Users().GetAdmins()
	if err != nil {
		log.Fatalf("failed to get the list of admins: %s", err)
	}

	for i, user := range admins {
		log.Printf("%d: %s", i+1, user.Name)
	}
}

// addAdmin adds a user to the list of administrators.
func addAdmin() {
	setAdmin(true)
}

// removeAdmin removes a user to the list of administrators.
func removeAdmin() {
	setAdmin(false)
}

func setAdmin(isAdmin bool) {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(2)
	}
	username := os.Args[2]

	cfg, err := config.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to load the configuration file: %s", err)
	}

	s, err := getStore(cfg)
	if err != nil {
		log.Fatalf("failed to get data store: %s", err)
	}

	user, err := s.Users().GetByName(username)
	if err != nil {
		if err == store.ErrNotFound {
			log.Fatalf("user not found: %s", username)
		} else {
			log.Fatalf("user search by username failed: %s", err)
		}
	}

	if user.Admin == isAdmin {
		if isAdmin {
			log.Fatalf("user %s is already an admin", username)
		} else {
			log.Fatalf("user %s is not an admin", username)
		}
	}

	err = s.Users().SetAdmin(user.ID, isAdmin)
	if err != nil {
		log.Fatalf("failed to change user admin rights: %s", err)
	}

	if isAdmin {
		log.Printf("user %s is added to admin list", username)
	} else {
		log.Printf("user %s is removed from admin list", username)
	}
}
