package main

import (
	"github.com/go-chef/chef"
)

// ChefClient represents a ready-to-use chef client
type ChefClient struct {
	client *chef.Client
}

// CreateChefClient creates a Chef client
func CreateChefClient(key string) (ChefClient, error) {
	apiClient, err := chef.NewClient(&chef.Config{
		Name:    cfg.ChefClientName,
		Key:     key,
		BaseURL: cfg.ChefServer,
		SkipSSL: !cfg.ChefVerifyTLS})
	chef := ChefClient{apiClient}

	return chef, err
}

// GetClient fetches a client record from the Chef server
func (c *ChefClient) GetClient(name string) (client chef.ApiClient, err error) {
	record, err := c.client.Clients.Get(name)
	return record, err
}
