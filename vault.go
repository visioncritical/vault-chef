package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

// VaultClient represents a ready-to-use Vault client
type VaultClient struct {
	client *api.Client
	renew  bool
}

// CreateVaultClient creates a Vault client
func CreateVaultClient() (*VaultClient, error) {
	apiClient, err := api.NewClient(nil)
	if err != nil {
		return nil, err
	}
	vault := &VaultClient{apiClient, cfg.RenewToken}
	return vault, err
}

// UnwrapToken returns the token inside a wrapped token
// or uses the token as is if it's not a wrapped token
func (v *VaultClient) UnwrapToken(vaultToken string) (string, error) {
	secret, err := v.client.Logical().Unwrap(vaultToken)
	if err != nil {
		// Assume it's not a wrapped token
		return vaultToken, nil
	}

	return secret.Auth.ClientToken, err
}

// ReadSecretPath retrieves a secret from Vault
func (v *VaultClient) ReadSecretPath(path string) (*api.Secret, error) {
	secret, err := v.client.Logical().Read(path)
	return secret, err
}

// IssueTokenRole returns a Vault Role Token
func (v *VaultClient) IssueTokenRole(tcr *api.TokenCreateRequest, role string) (string, error) {
	secret, err := v.client.Auth().Token().CreateWithRole(tcr, role)
	if err != nil {
		return "", err
	}
	return secret.Auth.ClientToken, err
}

// RenewToken sets up a loop to renew the calling Vault token
// Should be used in a Goroutine
func (v *VaultClient) RenewToken() error {
	secret, err := v.client.Auth().Token().LookupSelf()
	if err != nil {
		log.Fatalf("[FATAL] Could not set up Vault token renewal. Failed to lookup-self. %e", err)
	}
	creationTTL, err := secret.Data["creation_ttl"].(json.Number).Int64()
	if err != nil {
		log.Fatalf("[FATAL] Could not determine creation_ttl time for Vault token. %e", err)
	}
	renewalDelay := creationTTL / 3
	for {
		_, err := v.client.Auth().Token().RenewSelf(int(creationTTL))
		if err != nil {
			log.Printf("[WARN] Failed to renew Vault token. %s", err)
		}
		log.Printf("[INFO] Successfully renewed Vault token. Sleeping for %d", int(renewalDelay))
		time.Sleep(time.Duration(renewalDelay) * time.Second)
	}
}
