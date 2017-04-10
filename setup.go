package main

import (
	"log"
	"os"
	"strconv"
)

// Config holds the user-specified configuration values
// Standard environment variables for Vault will be respected
type Config struct {
	ChefServer        string
	ChefClientName    string
	ChefVerifyTLS     bool
	VaultAddr         string
	VaultToken        string
	VaultSkipVerify   bool
	SecretCertPath    string
	SecretChefPemPath string
	RenewToken        bool
	ListenAddr        string
}

// Clients is a wrapper for our clients
type Clients struct {
	Vault VaultClient
	Chef  ChefClient
}

// DefaultConfig returns a Config struct with some default values
func DefaultConfig() *Config {
	config := &Config{
		VaultAddr:       "https://127.0.0.1:8200",
		ChefVerifyTLS:   true,
		VaultSkipVerify: false,
		RenewToken:      true,
		ListenAddr:      ":8443",
	}
	return config
}

// ParseConfig creates a Config struct object to be used within this application
func ParseConfig() Config {
	c := DefaultConfig()

	if c.ChefServer = os.Getenv("CHEF_SERVER"); c.ChefServer == "" {
		log.Fatalf("[FATAL] CHEF_SERVER was not defined")
	}
	if c.ChefClientName = os.Getenv("CHEF_CLIENT_NAME"); c.ChefClientName == "" {
		log.Fatalf("[FATAL] CHEF_CLIENT_NAME was not defined")
	}
	if os.Getenv("CHEF_VERIFY_TLS") == "" {
		log.Printf("[WARN] CHEF_VERIFY_TLS set to false")
	} else {
		chefVerifyTLS, err := strconv.ParseBool(os.Getenv("CHEF_VERIFY_TLS"))
		if err != nil {
			log.Fatalf("[FATAL] Invalid value for CHEF_VERIFY_TLS: %s", err)
		}
		c.ChefVerifyTLS = chefVerifyTLS
	}
	if c.VaultAddr = os.Getenv("VAULT_ADDR"); c.VaultAddr == "" {
		log.Fatalf("[FATAL] VAULT_ADDR was not defined")
	}
	if c.VaultToken = os.Getenv("VAULT_TOKEN"); c.VaultToken == "" {
		log.Fatalf("[FATAL] VAULT_TOKEN was not defined")
	}
	if os.Getenv("VAULT_SKIP_VERIFY") == "" {
		log.Printf("[WARN] VAULT_SKIP_VERIFY set to false")
	} else {
		vaultSkipVerify, err := strconv.ParseBool(os.Getenv("VAULT_SKIP_VERIFY"))
		if err != nil {
			log.Fatalf("[FATAL] Invalid value for VAULT_SKIP_VERIFY: %s", err)
		}
		c.VaultSkipVerify = vaultSkipVerify
	}
	if c.SecretCertPath = os.Getenv("SECRET_CERT_PATH"); c.SecretCertPath == "" {
		log.Fatalf("[FATAL] SECRET_CERT_PATH was not defined")
	}
	if c.SecretChefPemPath = os.Getenv("SECRET_CHEF_PEM_PATH"); c.SecretChefPemPath == "" {
		log.Fatalf("[FATAL] SECRET_CHEF_PEM_PATH was not defined")
	}
	if os.Getenv("RENEW_TOKEN") == "" {
		log.Printf("[INFO] Skipping Vault token renewal")
	} else {
		renewToken, err := strconv.ParseBool(os.Getenv("RENEW_TOKEN"))
		if err != nil {
			log.Fatalf("[FATAL] Invalid value for RENEW_TOKEN: %s", err)
		}
		c.RenewToken = renewToken
	}
	if c.ListenAddr = os.Getenv("LISTEN_ADDR"); c.ListenAddr == "" {
		log.Fatalf("[FATAL] LISTEN_ADDR was not defined")
	}

	return *c
}

// SetupClients creates clients to both Vault & Chef
func SetupClients() Clients {
	vaultClient, err := CreateVaultClient()
	if err != nil {
		log.Fatalf("[FATAL] Failed to create Vault client: %s", err)
	}
	// Either unwraps it if possible, or just uses the value as is
	token, _ := vaultClient.UnwrapToken(cfg.VaultToken)
	vaultClient.client.SetToken(token)

	chefPEM, err := vaultClient.ReadSecretPath(cfg.SecretChefPemPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to get Chef PEM from Vault: %s", err)
	}
	chefClient, err := CreateChefClient(chefPEM.Data["key"].(string))
	if err != nil {
		log.Fatalf("[FATAL] Failed to create Chef client: %s", err)
	}
	return Clients{
		Vault: *vaultClient,
		Chef:  chefClient,
	}
}
