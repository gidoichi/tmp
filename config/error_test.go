package config_test

import (
	"errors"
	"testing"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/config"
)

func TestConfigErrorCallingError(t *testing.T) {
	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"ReturnsStringWithColonSeparatorWhenWrappedErrorIsGeneralError",
			func(t *testing.T) {
				// Given
				childErr := errors.New("child")
				err := config.NewConfigError("path", childErr)

				// When
				errStr := err.Error()

				// Then
				if errStr != "path: child" {
					t.Errorf("unexpected error string: %s", errStr)
				}
			},
		},
		{
			"ReturnsPathJoinedStringWhenWrappedErrorIsConfigError",
			func(t *testing.T) {
				// Given
				childErr := config.NewConfigError("childPath", errors.New("child"))
				err := config.NewConfigError("path", childErr)

				// When
				errStr := err.Error()

				// Then
				if errStr != "path.childPath: child" {
					t.Errorf("unexpected error string: %s", errStr)
				}
			},
		},
	} {
		t.Run(testcase.name, testcase.f)
	}
}

func TestConfigErrorUnwraps(t *testing.T) {
	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"WrappedError",
			func(t *testing.T) {
				// Given
				childErr := errors.New("child")
				err := config.NewConfigError("path", childErr)

				// When
				unrapped := err.Unwrap()

				// Then
				if unrapped != childErr {
					t.Errorf("unexpected unwrapped error: %v", unrapped)
				}
			},
		},
	} {
		t.Run(testcase.name, testcase.f)
	}
}
