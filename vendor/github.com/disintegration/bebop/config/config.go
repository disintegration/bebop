// Package config provides the bebop configuration file structure,
// initialization and reading.
package config

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl"
)

// Config is a bebop configuration struct.
type Config struct {
	Address string `hcl:"address"`
	BaseURL string `hcl:"base_url"`
	Title   string `hcl:"title"`

	JWT struct {
		Secret string `hcl:"secret"`
	} `hcl:"jwt"`

	FileStorage struct {
		Type string `hcl:"type"`

		Local struct {
			Dir string `hcl:"dir"`
		} `hcl:"local"`

		GoogleCloudStorage struct {
			ServiceAccountFile string `hcl:"service_account_file"`
			Bucket             string `hcl:"bucket"`
		} `hcl:"google_cloud_storage"`

		AmazonS3 struct {
			AccessKey string `hcl:"access_key"`
			SecretKey string `hcl:"secret_key"`
			Region    string `hcl:"region"`
			Bucket    string `hcl:"bucket"`
		} `hcl:"amazon_s3"`
	} `hcl:"file_storage"`

	Store struct {
		Type string `hcl:"type"`

		MySQL struct {
			Address  string `hcl:"address"`
			Username string `hcl:"username"`
			Password string `hcl:"password"`
			Database string `hcl:"database"`
		} `hcl:"mysql"`

		PostgreSQL struct {
			Address  string `hcl:"address"`
			Username string `hcl:"username"`
			Password string `hcl:"password"`
			Database string `hcl:"database"`
		} `hcl:"postgresql"`
	} `hcl:"store"`

	OAuth map[string]struct {
		ClientID string `hcl:"client_id"`
		Secret   string `hcl:"secret"`
	} `hcl:"oauth"`
}

// GenKeyHex generates a crypto-random key with byte length byteLen
// and hex-encodes it to a string.
func GenKeyHex(byteLen int) string {
	bytes := make([]byte, byteLen)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

// ReadFile reads a bebop config from file.
func ReadFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer f.Close()
	return Read(f)
}

// Read reads a bebop config from r.
func Read(r io.Reader) (*Config, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	cfg := &Config{}
	err = hcl.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal hcl: %v", err)
	}

	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")

	return cfg, nil
}

// Init generates an initial config string.
func Init() (string, error) {
	buf := new(bytes.Buffer)
	err := tpl.Execute(buf, map[string]interface{}{
		"jwt_secret": GenKeyHex(32),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

var tpl = template.Must(template.New("initial-config").Parse(strings.TrimSpace(`
address  = "127.0.0.1:8080"
base_url = "https://example.com/forum"
title    = "bebop"

jwt {
  secret = "{{.jwt_secret}}"
}

file_storage {
  type = "local"

  local {
    dir = "./bebop_data/public/"
  }

  google_cloud_storage {
    service_account_file = ""
    bucket               = ""
  }

  amazon_s3 {
    access_key = ""
    secret_key = ""
    region     = ""
    bucket     = ""
  }
}

store {
  type = "postgresql"

  postgresql {
    address  = "127.0.0.1:5432"
    username = ""
    password = ""
    database = ""
  }

  mysql {
    address  = "127.0.0.1:3306"
    username = ""
    password = ""
    database = ""
  }
}

oauth {
  google {
    client_id = ""
    secret    = ""
  }

  facebook {
    client_id = ""
    secret    = ""
  }

  github {
    client_id = ""
    secret    = ""
  }
}
`)))
