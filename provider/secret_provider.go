//go:generate mockgen -destination=mock_$GOPACKAGE/mock_$GOFILE -source=$GOFILE
package provider

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	infisical "github.com/infisical/go-sdk"
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
	UniversalAuthLogin(string, string) (infisical.MachineIdentityCredential, error)
	ListSecrets(infisical.ListSecretsOptions) ([]infisical.Secret, error)
}

type infisicalClient struct {
	client infisical.InfisicalClientInterface
	auth   *infisical.MachineIdentityCredential
}

func NewInfisicalClient(config infisical.Config) InfisicalClient {
	return &infisicalClient{
		client: infisical.NewInfisicalClient(config),
	}
}

func (c *infisicalClient) UniversalAuthLogin(clientID, clientSecret string) (infisical.MachineIdentityCredential, error) {
	return c.client.Auth().UniversalAuthLogin(clientID, clientSecret)
}

func (c *infisicalClient) ListSecrets(options infisical.ListSecretsOptions) ([]infisical.Secret, error) {
	secrets, err := c.client.Secrets().List(options)
	if err != nil {
		return nil, err
	}
	if !options.ExpandSecretReferences {
		return secrets, nil
	}

	return c.expandSecrets(secrets, options)
}

func (c *infisicalClient) GetAllEnvironmentVariables(options infisical.ListSecretsOptions) ([]infisical.Secret, error) {
	options.ExpandSecretReferences = false
	return c.ListSecrets(options)
}

// c.f. https://github.com/Infisical/infisical/blob/a6f4a95821d2dd597a801af7ec873a98d46b5ff8/cli/packages/util/secrets.go#L333
var secRefRegex = regexp.MustCompile(`\${([^\}]*)}`)

// c.f. https://github.com/Infisical/infisical/blob/a6f4a95821d2dd597a801af7ec873a98d46b5ff8/cli/packages/util/secrets.go#L335
func (c *infisicalClient) recursivelyExpandSecret(expandedSecs map[string]string, interpolatedSecs map[string]string, crossSecRefFetch func(env string, path []string, key string) (string, error), key string) (string, error) {
	if v, ok := expandedSecs[key]; ok {
		return v, nil
	}

	interpolatedVal, ok := interpolatedSecs[key]
	if !ok {
		return "", fmt.Errorf("could not find refered secret -  %s", key)
	}

	refs := secRefRegex.FindAllStringSubmatch(interpolatedVal, -1)
	for _, val := range refs {
		// key: "${something}" val: [${something},something]
		interpolatedExp, interpolationKey := val[0], val[1]
		ref := strings.Split(interpolationKey, ".")

		// ${KEY1} => [key1]
		if len(ref) == 1 {
			val, err := c.recursivelyExpandSecret(expandedSecs, interpolatedSecs, crossSecRefFetch, interpolationKey)
			if err != nil {
				return "", err
			}
			interpolatedVal = strings.ReplaceAll(interpolatedVal, interpolatedExp, val)
			continue
		}

		// cross board reference ${env.folder.key1} => [env folder key1]
		if len(ref) > 1 {
			var err error
			secEnv, tmpSecPath, secKey := ref[0], ref[1:len(ref)-1], ref[len(ref)-1]
			interpolatedSecs[interpolationKey], err = crossSecRefFetch(secEnv, tmpSecPath, secKey) // get the reference value
			if err != nil {
				return "", err
			}
			val, err := c.recursivelyExpandSecret(expandedSecs, interpolatedSecs, crossSecRefFetch, interpolationKey)
			if err != nil {
				return "", err
			}
			interpolatedVal = strings.ReplaceAll(interpolatedVal, interpolatedExp, val)
		}

	}
	expandedSecs[key] = interpolatedVal
	return interpolatedVal, nil
}

// c.f. https://github.com/Infisical/infisical/blob/a6f4a95821d2dd597a801af7ec873a98d46b5ff8/cli/packages/util/secrets.go#L371
func getSecretsByKeys(secrets []infisical.Secret) map[string]infisical.Secret {
	secretMapByName := make(map[string]infisical.Secret, len(secrets))

	for _, secret := range secrets {
		secretMapByName[secret.SecretKey] = secret
	}

	return secretMapByName
}

// c.f. https://github.com/Infisical/infisical/blob/a6f4a95821d2dd597a801af7ec873a98d46b5ff8/cli/packages/util/secrets.go#L381
func (c *infisicalClient) expandSecrets(secrets []infisical.Secret, options infisical.ListSecretsOptions) ([]infisical.Secret, error) {
	expandedSecs := make(map[string]string)
	interpolatedSecs := make(map[string]string)
	// map[env.secret-path][keyname]Secret
	crossEnvRefSecs := make(map[string]map[string]infisical.Secret) // a cache to hold all cross board reference secrets

	for _, sec := range secrets {
		// get all references in a secret
		refs := secRefRegex.FindAllStringSubmatch(sec.SecretValue, -1)
		// nil means its a secret without reference
		if refs == nil {
			expandedSecs[sec.SecretKey] = sec.SecretValue // atomic secrets without any interpolation
		} else {
			interpolatedSecs[sec.SecretKey] = sec.SecretValue
		}
	}

	for i, sec := range secrets {
		// already present pick that up
		if expandedVal, ok := expandedSecs[sec.SecretKey]; ok {
			secrets[i].SecretValue = expandedVal
			continue
		}

		expandedVal, err := c.recursivelyExpandSecret(expandedSecs, interpolatedSecs, func(env string, secPaths []string, secKey string) (string, error) {
			secPaths = append([]string{"/"}, secPaths...)
			secPath := path.Join(secPaths...)

			secPathDot := strings.Join(secPaths, ".")
			uniqKey := fmt.Sprintf("%s.%s", env, secPathDot)

			if crossRefSec, ok := crossEnvRefSecs[uniqKey]; !ok {
				// if not in cross reference cache, fetch it from server
				options := options
				options.Environment = env
				options.SecretPath = secPath
				refSecs, err := c.GetAllEnvironmentVariables(options)
				if err != nil {
					return "", fmt.Errorf("Could not fetch secrets in environment: %s secret-path: %s: %w", env, secPath, err)
				}
				refSecsByKey := getSecretsByKeys(refSecs)
				// save it to avoid calling api again for same environment and folder path
				crossEnvRefSecs[uniqKey] = refSecsByKey
				return refSecsByKey[secKey].SecretValue, nil

			} else {
				return crossRefSec[secKey].SecretValue, nil
			}
		}, sec.SecretKey)
		if err != nil {
			return nil, err
		}

		secrets[i].SecretValue = expandedVal
	}
	return secrets, nil
}
