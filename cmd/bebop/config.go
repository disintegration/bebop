package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/disintegration/bebop/config"
	"github.com/disintegration/bebop/filestorage"
	"github.com/disintegration/bebop/store"
	"github.com/disintegration/bebop/store/mysql"
	"github.com/disintegration/bebop/store/postgresql"
)

const configFile = "bebop.conf"

// initConfig creates an initial configuration file inside the current directory.
func initConfig() {
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		log.Fatalf("configuration file already exists: %s", configFile)
	}

	log.Printf("creating initial configuration: %s", configFile)

	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("failed to generate initial configuration: %s", err)
	}

	err = ioutil.WriteFile(configFile, []byte(cfg), 0666)
	if err != nil {
		log.Fatalf("failed to write configuration file: %s", err)
	}
}

// genKey generates a random 32-byte hex-encoded key.
func genKey() {
	log.Printf("key: %s", config.GenKeyHex(32))
}

func getFileStorage(cfg *config.Config) (filestorage.FileStorage, error) {
	switch cfg.FileStorage.Type {
	case "local":
		return filestorage.NewLocal(
			cfg.FileStorage.Local.Dir,
			cfg.BaseURL+"/static",
		)
	case "google_cloud_storage":
		return filestorage.NewGoogleCloudStorage(
			cfg.FileStorage.GoogleCloudStorage.ServiceAccountFile,
			cfg.FileStorage.GoogleCloudStorage.Bucket,
		)
	case "amazon_s3":
		return filestorage.NewAmazonS3(
			cfg.FileStorage.AmazonS3.AccessKey,
			cfg.FileStorage.AmazonS3.SecretKey,
			cfg.FileStorage.AmazonS3.Region,
			cfg.FileStorage.AmazonS3.Bucket,
		)
	}
	return nil, fmt.Errorf("unknown file storage type: %s", cfg.FileStorage.Type)
}

func getStore(cfg *config.Config) (store.Store, error) {
	switch cfg.Store.Type {
	case "mysql":
		return mysql.Connect(
			cfg.Store.MySQL.Address,
			cfg.Store.MySQL.Username,
			cfg.Store.MySQL.Password,
			cfg.Store.MySQL.Database,
		)
	case "postgresql":
		return postgresql.Connect(
			cfg.Store.PostgreSQL.Address,
			cfg.Store.PostgreSQL.Username,
			cfg.Store.PostgreSQL.Password,
			cfg.Store.PostgreSQL.Database,
			cfg.Store.PostgreSQL.SSLMode,
			cfg.Store.PostgreSQL.SSLRootCert,
		)
	}
	return nil, fmt.Errorf("unknown store type: %s", cfg.Store.Type)
}
