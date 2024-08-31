package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/auth"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/config"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/provider"
	"github.com/go-playground/validator/v10"
	infisical "github.com/infisical/go-sdk"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

var (
	ErrorInvalidSecretProviderClass = "InvalidSecretProviderClass"
	ErrorUnauthorized               = "Unauthorized"
	ErrorBadRequest                 = "BadRequest"
)

type CSIProviderServer struct {
	version                string
	grpcServer             *grpc.Server
	listener               net.Listener
	socketPath             string
	auth                   auth.Auth
	infisicalClientFactory provider.InfisicalClientFactory
	validator              *validator.Validate
}

var _ v1alpha1.CSIDriverProviderServer = &CSIProviderServer{}

// NewCSIProviderServer returns a mock csi-provider grpc server
func NewCSIProviderServer(version, socketPath string, auth auth.Auth, infisicalClientFactory provider.InfisicalClientFactory) *CSIProviderServer {
	server := grpc.NewServer()
	s := &CSIProviderServer{
		version:                version,
		grpcServer:             server,
		socketPath:             socketPath,
		auth:                   auth,
		infisicalClientFactory: infisicalClientFactory,
		validator:              validator.New(validator.WithRequiredStructEnabled()),
	}
	v1alpha1.RegisterCSIDriverProviderServer(server, s)
	return s
}

func (m *CSIProviderServer) Start() error {
	var err error
	m.listener, err = net.Listen("unix", m.socketPath)
	if err != nil {
		return err
	}
	go func() {
		if err = m.grpcServer.Serve(m.listener); err != nil {
			return
		}
	}()
	return nil
}

func (m *CSIProviderServer) Stop() {
	m.grpcServer.GracefulStop()
}

// Mount implements provider csi-provider method
func (s *CSIProviderServer) Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	mountResponse := &v1alpha1.MountResponse{
		Error: &v1alpha1.Error{},
	}

	slog.Info("mount", "request", req)

	// parse request
	mountConfig := config.NewMountConfig(*s.validator)
	var secret map[string]string
	var filePermission os.FileMode
	attributesDecoder := json.NewDecoder(strings.NewReader(req.GetAttributes()))
	attributesDecoder.DisallowUnknownFields()
	if err := attributesDecoder.Decode(&mountConfig); err != nil {
		mountResponse.Error.Code = ErrorInvalidSecretProviderClass
		return mountResponse, fmt.Errorf("failed to unmarshal parameters, error: %w", err)
	}
	if err := mountConfig.Validate(); err != nil {
		mountResponse.Error.Code = ErrorInvalidSecretProviderClass
		return mountResponse, fmt.Errorf("failed to validate parameters, error: %w", err)
	}
	if err := json.Unmarshal([]byte(req.GetSecrets()), &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets, error: %w", err)
	}
	if err := json.Unmarshal([]byte(req.GetPermission()), &filePermission); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file permission, error: %w", err)
	}
	objects, err := mountConfig.Objects()
	if err != nil {
		mountResponse.Error.Code = ErrorInvalidSecretProviderClass
		return mountResponse, fmt.Errorf("failed to get objects, error: %w", err)
	}
	if mountConfig.RawObjects != nil && len(objects) == 0 {
		mountResponse.ObjectVersion = []*v1alpha1.ObjectVersion{
			{
				Id:      "NO_SECRETS",
				Version: "0",
			},
		}
		return mountResponse, nil
	}

	// get credentials
	kubeSecret := types.NamespacedName{
		Namespace: mountConfig.AuthSecretNamespace,
		Name:      mountConfig.AuthSecretName,
	}
	credentials, err := s.auth.TokenFromKubeSecret(ctx, kubeSecret)
	if err != nil {
		mountResponse.Error.Code = ErrorBadRequest
		return mountResponse, fmt.Errorf("failed to get credentials, error: %w", err)
	}

	// get secrets
	infisicalClient := s.infisicalClientFactory.NewClient(infisical.Config{})
	if _, err := infisicalClient.UniversalAuthLogin(credentials.ID, credentials.Secret); err != nil {
		mountResponse.Error.Code = ErrorUnauthorized
		return mountResponse, fmt.Errorf("failed to login infisical, error: %w", err)
	}
	secrets, err := infisicalClient.ListSecrets(infisical.ListSecretsOptions{
		ProjectSlug:            mountConfig.Project,
		Environment:            mountConfig.Env,
		SecretPath:             mountConfig.Path,
		ExpandSecretReferences: true,
	})
	if err != nil {
		mountResponse.Error.Code = ErrorBadRequest
		return mountResponse, fmt.Errorf("failed to list secrets, error: %w", err)
	}

	// store secrets
	var objectVersions []*v1alpha1.ObjectVersion
	var files []*v1alpha1.File
	if mountConfig.RawObjects == nil {
		// all secrets
		mode := int32(filePermission)
		for _, secret := range secrets {
			objectVersions = append(objectVersions, &v1alpha1.ObjectVersion{
				Id:      secret.SecretKey,
				Version: fmt.Sprint(secret.Version),
			})

			files = append(files, &v1alpha1.File{
				Path:     secret.SecretKey,
				Mode:     mode,
				Contents: []byte(secret.SecretValue),
			})
		}
	} else {
		// specified secrets
		secretsMap := map[string]infisical.Secret{}
		for _, secret := range secrets {
			secretsMap[secret.SecretKey] = secret
		}
		for _, object := range objects {
			secret, ok := secretsMap[object.Name]
			if !ok {
				mountResponse.Error.Code = ErrorBadRequest
				return mountResponse, fmt.Errorf("object %s not found in secrets", object.Name)
			}

			objectVersions = append(objectVersions, &v1alpha1.ObjectVersion{
				Id:      object.Name,
				Version: fmt.Sprint(secret.Version),
			})

			files = append(files, &v1alpha1.File{
				Path: func() string {
					if object.Alias != "" {
						return object.Alias
					} else {
						return object.Name
					}
				}(),
				Mode:     int32(filePermission),
				Contents: []byte(secret.SecretValue),
			})
		}
	}
	mountResponse.ObjectVersion = objectVersions
	mountResponse.Files = files

	return mountResponse, nil
}

// Version implements provider csi-provider method
func (m *CSIProviderServer) Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	return &v1alpha1.VersionResponse{
		Version:        "v1alpha1",
		RuntimeName:    "secrets-store-csi-driver-provider-infisical",
		RuntimeVersion: m.version,
	}, nil
}
