package main

import (
	"flag"
	"os"

	"github.com/disintegration/bebop/store"
)

// printAdmins prints all the administrator users.
func printAdmins() {
	cfg, err := getConfig()
	if err != nil {
		logger.Fatalf("failed to load configuration: %s", err)
	}

	s, err := getStore(cfg)
	if err != nil {
		logger.Fatalf("failed to get data store: %s", err)
	}

	admins, err := s.Users().GetAdmins()
	if err != nil {
		logger.Fatalf("failed to get the list of admins: %s", err)
	}

	for i, user := range admins {
		logger.Printf("%d: %s", i+1, user.Name)
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
	username := flag.Arg(1)
	if username == "" {
		help()
		os.Exit(2)
	}

	cfg, err := getConfig()
	if err != nil {
		logger.Fatalf("failed to load configuration: %s", err)
	}

	s, err := getStore(cfg)
	if err != nil {
		logger.Fatalf("failed to get data store: %s", err)
	}

	user, err := s.Users().GetByName(username)
	if err != nil {
		if err == store.ErrNotFound {
			logger.Fatalf("user not found: %s", username)
		} else {
			logger.Fatalf("user search by username failed: %s", err)
		}
	}

	if user.Admin == isAdmin {
		if isAdmin {
			logger.Fatalf("user %s is already an admin", username)
		} else {
			logger.Fatalf("user %s is not an admin", username)
		}
	}

	err = s.Users().SetAdmin(user.ID, isAdmin)
	if err != nil {
		logger.Fatalf("failed to change user admin rights: %s", err)
	}

	if isAdmin {
		logger.Printf("user %s is added to admin list", username)
	} else {
		logger.Printf("user %s is removed from admin list", username)
	}
}
