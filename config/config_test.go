package config_test

import (
	"testing"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/config"
	"github.com/go-playground/validator/v10"
	"go.uber.org/thriftrw/ptr"
)

func TestMountConfigValidates(t *testing.T) {
	var (
		validate *validator.Validate
	)

	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"SuccessfullyWithFullConfig",
			func(t *testing.T) {
				// Given
				mountConfig := config.NewMountConfig(*validate)
				mountConfig.Project = "test-project"
				mountConfig.Env = "dev"
				mountConfig.Path = "/path/to/secrets"
				mountConfig.AuthSecretName = "test"
				mountConfig.AuthSecretNamespace = "test-namepace"
				mountConfig.RawObjects = ptr.String("- objectName: test")

				// When
				err := mountConfig.Validate()

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			},
		},
		{
			"FailedWithoutRequiredFields",
			func(t *testing.T) {
				// Given
				mountConfig := config.NewMountConfig(*validate)

				// When
				err := mountConfig.Validate()

				// Then
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
		{
			"FailedWithInvalidRawObjects",
			func(t *testing.T) {
				// Given
				mountConfig := config.NewMountConfig(*validate)
				mountConfig.Project = "test-project"
				mountConfig.Env = "dev"
				mountConfig.AuthSecretName = "test"
				mountConfig.AuthSecretNamespace = "test-namepace"
				mountConfig.RawObjects = ptr.String("objectName: test")

				// When
				err := mountConfig.Validate()

				// Then
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
	} {
		validate = validator.New(validator.WithRequiredStructEnabled())

		t.Run(testcase.name, testcase.f)
	}
}

func TestMountConfigGetObjects(t *testing.T) {
	var (
		validate *validator.Validate
	)

	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"SuccessfullyWithCorrectRawObjects",
			func(t *testing.T) {
				// Given
				mountConfig := config.NewMountConfig(*validate)
				mountConfig.RawObjects = ptr.String("- objectName: test")

				// When
				objects, err := mountConfig.Objects()

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if len(objects) != 1 {
					t.Errorf("unexpected objects: %v", objects)
					t.FailNow()
				}
				if objects[0].Name != "test" {
					t.Errorf("unexpected objects: %v", objects)
				}
			},
		},
		{
			"ReturnsEmptyWithEmptyRawObjects",
			func(t *testing.T) {
				// Given
				mountConfig := config.NewMountConfig(*validate)
				mountConfig.RawObjects = ptr.String("")

				// When
				objects, err := mountConfig.Objects()

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if len(objects) != 0 {
					t.Errorf("unexpected objects: %v", objects)
				}
			},
		},
		{
			"ReturnsEmptyWithoutRawObjects",
			func(t *testing.T) {
				// Given
				mountConfig := config.NewMountConfig(*validate)

				// When
				objects, err := mountConfig.Objects()

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if len(objects) != 0 {
					t.Errorf("unexpected objects: %v", objects)
				}
			},
		},
		{
			"FailedWithInvalidRawObjects",
			func(t *testing.T) {
				// Given
				mountConfig := config.NewMountConfig(*validate)
				mountConfig.RawObjects = ptr.String("objectName: test")

				// When
				_, err := mountConfig.Objects()

				// Then
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
	} {
		validate = validator.New(validator.WithRequiredStructEnabled())

		t.Run(testcase.name, testcase.f)
	}
}
