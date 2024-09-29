package config_test

import (
	"strings"
	"testing"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/config"
	"github.com/go-playground/validator/v10"
	"go.uber.org/thriftrw/ptr"
)

func TestMountConfigValidates(t *testing.T) {
	var (
		idealMountConfig *config.MountConfig
		validate         *validator.Validate
	)

	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"SuccessfullyWithFullConfig",
			func(t *testing.T) {
				// Given
				mountConfig := idealMountConfig

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
				mountConfig := idealMountConfig
				mountConfig.RawObjects = ptr.String("objectName: test")

				// When
				err := mountConfig.Validate()

				// Then
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if !strings.HasPrefix(err.Error(), "objects: ") {
					t.Errorf("unexpected error: %s", err)
				}
			},
		},
		{
			"FailedWithoutRequiredRawObjectFields",
			func(t *testing.T) {
				// Given
				mountConfig := idealMountConfig
				mountConfig.RawObjects = ptr.String("- objectAlias: secret")

				// When
				err := mountConfig.Validate()

				// Then
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if !strings.HasPrefix(err.Error(), "objects: [0]: ") {
					t.Errorf("unexpected error: %s", err)
				}
			},
		},
		{
			"FailedWithInvalidCaractersWithinRawObjectFields",
			func(t *testing.T) {
				// Given
				mountConfig := idealMountConfig
				mountConfig.RawObjects = ptr.String("- objectName: test\n  objectAlias: path/to/secret")

				// When
				err := mountConfig.Validate()

				// Then
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if !strings.HasPrefix(err.Error(), "objects: [0]: ") {
					t.Errorf("unexpected error: %s", err)
				}
			},
		},
	} {
		validate = config.NewValidator()
		idealMountConfig = config.NewMountConfig(*validate)
		idealMountConfig.Project = "test-project"
		idealMountConfig.Env = "dev"
		idealMountConfig.Path = "/path/to/secrets"
		idealMountConfig.AuthSecretName = "test"
		idealMountConfig.AuthSecretNamespace = "test-namepace"
		idealMountConfig.RawObjects = ptr.String("- objectName: test")

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
			"FailedWithInvalidFormat",
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
		validate = config.NewValidator()

		t.Run(testcase.name, testcase.f)
	}
}
