//go:generate mockgen -destination=mock_$GOPACKAGE/mock_$GOFILE -source=$GOFILE
package provider

import (
	infisical "github.com/infisical/go-sdk"
	api "github.com/infisical/go-sdk/packages/api/auth"
	"github.com/infisical/go-sdk/packages/models"
)

type InfisicalClientFactory interface {
	NewClient(config infisical.Config) InfisicalClient
}

func NewInfisicalClientFactory() InfisicalClientFactory {
	return &infisicalClientFactory{}
}

type infisicalClientFactory struct{}

func (f *infisicalClientFactory) NewClient(config infisical.Config) InfisicalClient {
	return NewInfisicalClient(config)
}

type InfisicalClient interface {
	UniversalAuthLogin(string, string) (api.MachineIdentityAuthLoginResponse, error)
	ListSecrets(infisical.ListSecretsOptions) ([]models.Secret, error)
}

type infisicalClient struct {
	client infisical.InfisicalClientInterface
}

func NewInfisicalClient(config infisical.Config) InfisicalClient {
	return &infisicalClient{
		client: infisical.NewInfisicalClient(config),
	}
}

func (c *infisicalClient) UniversalAuthLogin(clientID, clientSecret string) (api.MachineIdentityAuthLoginResponse, error) {
	return c.client.Auth().UniversalAuthLogin(clientID, clientSecret)
}

func (c *infisicalClient) ListSecrets(options infisical.ListSecretsOptions) ([]models.Secret, error) {
	return c.client.Secrets().List(options)
}
