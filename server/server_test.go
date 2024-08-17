package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/auth"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/auth/mock_auth"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/provider/mock_provider"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/server"
	infisical "github.com/infisical/go-sdk"
	api "github.com/infisical/go-sdk/packages/api/auth"
	"github.com/infisical/go-sdk/packages/models"
	"go.uber.org/mock/gomock"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const (
	runtimeVersion = "0.0.1"
	socketPath     = "/tmp/test.sock"
)

var (
	ctx                        context.Context
	ctrl                       *gomock.Controller
	mockAuth                   *mock_auth.MockAuth
	mockInfisicalClientFactory *mock_provider.MockInfisicalClientFactory
	mockInfisicalClient        *mock_provider.MockInfisicalClient
)

func TestCSIProviderServerMounts(t *testing.T) {
	var (
		idealMountRequest      *v1alpha1.MountRequest
		idealKubeSecret        types.NamespacedName
		idealCredentials       *auth.Credentials
		expectedObjectVersions []*v1alpha1.ObjectVersion
		expectedFiles          []*v1alpha1.File
	)

	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"SuccessfullyWithSecrets",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(idealCredentials, nil)
				mockInfisicalClientFactory.EXPECT().NewClient(infisical.Config{}).Return(mockInfisicalClient)
				mockInfisicalClient.EXPECT().UniversalAuthLogin(idealCredentials.ID, idealCredentials.Secret)
				mockInfisicalClient.EXPECT().ListSecrets(infisical.ListSecretsOptions{
					ProjectSlug:            "test-project",
					Environment:            "dev",
					SecretPath:             "/",
					ExpandSecretReferences: true,
				}).Return([]models.Secret{
					{
						SecretKey:   "DB_USERNAME",
						Version:     1,
						SecretValue: "admin",
					},
					{
						SecretKey:   "DB_PASSWORD",
						Version:     1,
						SecretValue: "password",
					},
				}, nil)

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, idealMountRequest)

				// Then

				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if actual.Error != nil && actual.Error.Code != "" {
					t.Errorf("unexpected error: %v", actual.Error)
				}

				if len(actual.ObjectVersion) != 2 {
					t.Errorf("unexpected object versions: %v", actual.ObjectVersion)
					t.FailNow()
				}
				if actual.ObjectVersion[0].Id != expectedObjectVersions[0].Id ||
					actual.ObjectVersion[0].Version != expectedObjectVersions[0].Version ||
					actual.ObjectVersion[1].Id != expectedObjectVersions[1].Id ||
					actual.ObjectVersion[1].Version != expectedObjectVersions[1].Version {
					t.Errorf("unexpected object versions: %v", actual.ObjectVersion)
				}

				if len(actual.Files) != 2 {
					t.Errorf("unexpected files: %v", actual.Files)
					t.FailNow()
				}
				if actual.Files[0].Path != expectedFiles[0].Path ||
					actual.Files[0].Mode != expectedFiles[0].Mode ||
					string(actual.Files[0].Contents) != string(expectedFiles[0].Contents) ||
					actual.Files[1].Path != expectedFiles[1].Path ||
					actual.Files[1].Mode != expectedFiles[1].Mode ||
					string(actual.Files[1].Contents) != string(expectedFiles[1].Contents) {
					t.Errorf("unexpected files: %v", actual.Files)
				}
			},
		},
		{
			"SuccessfullyWithMinimumConfiguredMountRequest",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(idealCredentials, nil)
				mockInfisicalClientFactory.EXPECT().NewClient(infisical.Config{}).Return(mockInfisicalClient)
				mockInfisicalClient.EXPECT().UniversalAuthLogin(idealCredentials.ID, idealCredentials.Secret)
				mockInfisicalClient.EXPECT().ListSecrets(infisical.ListSecretsOptions{
					ProjectSlug:            "test-project",
					Environment:            "dev",
					SecretPath:             "/",
					ExpandSecretReferences: true,
				}).Return([]models.Secret{
					{
						SecretKey:   "DB_USERNAME",
						Version:     1,
						SecretValue: "admin",
					},
					{
						SecretKey:   "DB_PASSWORD",
						Version:     1,
						SecretValue: "password",
					},
				}, nil)
				mountRequest := &v1alpha1.MountRequest{
					Attributes: `{"projectSlug":"test-project","envSlug":"dev","authSecretName":"test-infisical-credentials","authSecretNamespace":"test-namepace"}`,
					Secrets:    "{}",
					Permission: "420",
				}

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, mountRequest)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if actual.Error != nil && actual.Error.Code != "" {
					t.Errorf("unexpected error: %v", actual.Error)
				}
			},
		},
		{
			"SuccessfullyWithNoSecretsWhenEmptyObjectsGiven",
			func(t *testing.T) {
				// Given
				mountRequest := &v1alpha1.MountRequest{
					Attributes: `{"projectSlug":"test-project","envSlug":"dev","authSecretName":"test-infisical-credentials","authSecretNamespace":"test-namepace","objects":""}`,
					Secrets:    "{}",
					Permission: "420",
				}

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, mountRequest)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if actual.Error != nil && actual.Error.Code != "" {
					t.Errorf("unexpected error: %v", actual.Error)
				}
				if len(actual.Files) != 0 {
					t.Errorf("unexpected files: %v", actual.Files)
				}
			},
		},
		{
			"SuccessfullyWithAllSecretsWhenNoObjectsGiven",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(idealCredentials, nil)
				mockInfisicalClientFactory.EXPECT().NewClient(infisical.Config{}).Return(mockInfisicalClient)
				mockInfisicalClient.EXPECT().UniversalAuthLogin(idealCredentials.ID, idealCredentials.Secret)
				mockInfisicalClient.EXPECT().ListSecrets(infisical.ListSecretsOptions{
					ProjectSlug:            "test-project",
					Environment:            "dev",
					SecretPath:             "/",
					ExpandSecretReferences: true,
				}).Return([]models.Secret{
					{
						SecretKey:   "DB_USERNAME",
						Version:     1,
						SecretValue: "admin",
					},
					{
						SecretKey:   "DB_PASSWORD",
						Version:     1,
						SecretValue: "password",
					},
				}, nil)
				mountRequest := &v1alpha1.MountRequest{
					Attributes: `{"projectSlug":"test-project","envSlug":"dev","authSecretName":"test-infisical-credentials","authSecretNamespace":"test-namepace"}`,
					Secrets:    "{}",
					Permission: "420",
				}

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, mountRequest)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if actual.Error != nil && actual.Error.Code != "" {
					t.Errorf("unexpected error: %v", actual.Error)
				}
				if len(actual.Files) != 2 {
					t.Errorf("unexpected files: %v", actual.Files)
				}
			},
		},
		{
			"SuccessfullyWithSpecifiedSecretsWhenSomeObjectsGiven",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(idealCredentials, nil)
				mockInfisicalClientFactory.EXPECT().NewClient(infisical.Config{}).Return(mockInfisicalClient)
				mockInfisicalClient.EXPECT().UniversalAuthLogin(idealCredentials.ID, idealCredentials.Secret)
				mockInfisicalClient.EXPECT().ListSecrets(infisical.ListSecretsOptions{
					ProjectSlug:            "test-project",
					Environment:            "dev",
					SecretPath:             "/",
					ExpandSecretReferences: true,
				}).Return([]models.Secret{
					{
						SecretKey:   "DB_USERNAME",
						Version:     1,
						SecretValue: "admin",
					},
					{
						SecretKey:   "DB_PASSWORD",
						Version:     1,
						SecretValue: "password",
					},
				}, nil)
				mountRequest := &v1alpha1.MountRequest{
					Attributes: `{"projectSlug":"test-project","envSlug":"dev","authSecretName":"test-infisical-credentials","authSecretNamespace":"test-namepace","objects":"- objectName: DB_USERNAME"}`,
					Secrets:    "{}",
					Permission: "420",
				}

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, mountRequest)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if actual.Error != nil && actual.Error.Code != "" {
					t.Errorf("unexpected error: %v", actual.Error)
				}
				if len(actual.Files) != 1 {
					t.Errorf("unexpected files: %v", actual.Files)
					t.FailNow()
				}
				if actual.Files[0].Path != "DB_USERNAME" ||
					actual.Files[0].Mode != 420 ||
					string(actual.Files[0].Contents) != "admin" {
					t.Errorf("unexpected files: %v", actual.Files)
				}
			},
		},
		{
			"FailedWithUnknownAttributes",
			func(t *testing.T) {
				// Given
				mountRequest := &v1alpha1.MountRequest{
					Attributes: `{"projectSlug":"test-project","envSlug":"dev","authSecretName":"test-infisical-credentials","authSecretNamespace":"test-namepace", "unknown": "unknown"}`,
					Secrets:    "{}",
					Permission: "420",
				}

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, mountRequest)

				// Then
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				if actual.Error == nil || actual.Error.Code != server.ErrorInvalidSecretProviderClass {
					t.Errorf("unexpected error: %v", actual.Error)
				}
			},
		},
		{
			"FailedWithUnknownObjects",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(idealCredentials, nil)
				mockInfisicalClientFactory.EXPECT().NewClient(infisical.Config{}).Return(mockInfisicalClient)
				mockInfisicalClient.EXPECT().UniversalAuthLogin(idealCredentials.ID, idealCredentials.Secret)
				mockInfisicalClient.EXPECT().ListSecrets(infisical.ListSecretsOptions{
					ProjectSlug:            "test-project",
					Environment:            "dev",
					SecretPath:             "/",
					ExpandSecretReferences: true,
				}).Return([]models.Secret{
					{
						SecretKey:   "DB_USERNAME",
						Version:     1,
						SecretValue: "admin",
					},
					{
						SecretKey:   "DB_PASSWORD",
						Version:     1,
						SecretValue: "password",
					},
				}, nil)
				mountRequest := &v1alpha1.MountRequest{
					Attributes: `{"projectSlug":"test-project","envSlug":"dev","authSecretName":"test-infisical-credentials","authSecretNamespace":"test-namepace","objects":"- objectName: NOT_REGISTERED"}`,
					Secrets:    "{}",
					Permission: "420",
				}

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, mountRequest)

				// Then
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				if actual.Error == nil || actual.Error.Code != server.ErrorBadRequest {
					t.Errorf("unexpected error: %v", actual.Error)
				}
			},
		},
		{
			"FailedWithoutNecessaryAttributes",
			func(t *testing.T) {
				// Given
				idealMountRequest.Attributes = "{}"

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(context.Background(), idealMountRequest)

				// Then
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				if actual.Error == nil || actual.Error.Code != server.ErrorInvalidSecretProviderClass {
					t.Errorf("unexpected error: %v", actual.Error)
				}
			},
		},
		{
			"FailedWithoutKubeSecret",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(nil, errors.New("kube secret not found"))

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, idealMountRequest)

				// Then
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				if actual.Error == nil || actual.Error.Code != server.ErrorBadRequest {
					t.Errorf("unexpected error: %v", actual.Error)
				}
			},
		},
		{
			"FailedWithUniversalAuthLoginFailure",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(idealCredentials, nil)
				mockInfisicalClientFactory.EXPECT().NewClient(infisical.Config{}).Return(mockInfisicalClient)
				mockInfisicalClient.EXPECT().UniversalAuthLogin(idealCredentials.ID, idealCredentials.Secret).Return(api.MachineIdentityAuthLoginResponse{}, errors.New("failed to login"))

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, idealMountRequest)

				// Then
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				if actual.Error == nil || actual.Error.Code != server.ErrorUnauthorized {
					t.Errorf("unexpected error: %v", actual.Error)
				}
			},
		},
		{
			"FailedWithListSecretFailure",
			func(t *testing.T) {
				// Given
				mockAuth.EXPECT().TokenFromKubeSecret(ctx, idealKubeSecret).Return(idealCredentials, nil)
				mockInfisicalClientFactory.EXPECT().NewClient(infisical.Config{}).Return(mockInfisicalClient)
				mockInfisicalClient.EXPECT().UniversalAuthLogin(idealCredentials.ID, idealCredentials.Secret).Return(api.MachineIdentityAuthLoginResponse{}, nil)
				mockInfisicalClient.EXPECT().ListSecrets(infisical.ListSecretsOptions{
					ProjectSlug:            "test-project",
					Environment:            "dev",
					SecretPath:             "/",
					ExpandSecretReferences: true,
				}).Return(nil, errors.New("failed to list secrets"))

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Mount(ctx, idealMountRequest)

				// Then
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				if actual.Error == nil || actual.Error.Code != server.ErrorBadRequest {
					t.Errorf("unexpected error: %v", actual.Error)
				}
			},
		},
	} {
		ctx = context.Background()
		ctrl = gomock.NewController(t)
		mockAuth = mock_auth.NewMockAuth(ctrl)
		mockInfisicalClientFactory = mock_provider.NewMockInfisicalClientFactory(ctrl)
		mockInfisicalClient = mock_provider.NewMockInfisicalClient(ctrl)
		idealMountRequest = &v1alpha1.MountRequest{
			Attributes: `{"projectSlug":"test-project","envSlug":"dev","secretsPath":"/","authSecretName":"test-infisical-credentials","authSecretNamespace":"test-namepace"}`,
			Secrets:    "{}",
			Permission: "420",
		}
		idealKubeSecret = types.NamespacedName{
			Namespace: "test-namepace",
			Name:      "test-infisical-credentials",
		}
		idealCredentials = &auth.Credentials{
			ID:     "test-client-id",
			Secret: "test-client-secret",
		}
		expectedObjectVersions = []*v1alpha1.ObjectVersion{
			{
				Id:      "DB_USERNAME",
				Version: "1",
			},
			{
				Id:      "DB_PASSWORD",
				Version: "1",
			},
		}
		expectedFiles = []*v1alpha1.File{
			{
				Path:     "DB_USERNAME",
				Mode:     420,
				Contents: []byte("admin"),
			},
			{
				Path:     "DB_PASSWORD",
				Mode:     420,
				Contents: []byte("password"),
			},
		}

		t.Run(testcase.name, testcase.f)
	}
}

func TestCSIProviderServerVersion(t *testing.T) {
	var (
		idealVersionRequest *v1alpha1.VersionRequest
	)

	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"Successfully",
			func(t *testing.T) {
				// Given

				// When
				providerServer := server.NewCSIProviderServer(runtimeVersion, socketPath, mockAuth, mockInfisicalClientFactory)
				actual, err := providerServer.Version(ctx, idealVersionRequest)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if actual.Version != "v1alpha1" ||
					actual.RuntimeName != "secrets-store-csi-driver-provider-infisical" ||
					!semver.IsValid("v"+actual.RuntimeVersion) {
					t.Errorf("unexpected version response: %v", actual)
				}
			},
		},
	} {
		ctx = context.Background()
		ctrl = gomock.NewController(t)
		mockAuth = mock_auth.NewMockAuth(ctrl)
		mockInfisicalClientFactory = mock_provider.NewMockInfisicalClientFactory(ctrl)
		mockInfisicalClient = mock_provider.NewMockInfisicalClient(ctrl)

		t.Run(testcase.name, testcase.f)
	}
}
