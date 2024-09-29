package webhook_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/admission-webhook/pkg/webhook"
	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/config"
	"github.com/sirupsen/logrus"
	kwhlogrus "github.com/slok/kubewebhook/v2/pkg/log/logrus"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	corev1 "k8s.io/api/core/v1"
	secretstorecsidriverv1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

func TestSecretProviderClassWebhookValidates(t *testing.T) {
	var (
		ctx               context.Context
		validatingWebhook webhook.SecretProviderClassWebhook
		ar                *kwhmodel.AdmissionReview
	)

	for _, testcase := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"SuccessfullyWithNotSupportedKubernetesObject",
			func(t *testing.T) {
				// Given
				object := &corev1.Pod{}

				// When
				result, err := validatingWebhook.Validate(ctx, ar, object)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if !result.Valid {
					t.Errorf("expected valid, got invalid")
				}
			},
		},
		{
			"SuccessfullyWithNotSupportedSecretProvider",
			func(t *testing.T) {
				// Given
				spc := &secretstorecsidriverv1.SecretProviderClass{
					Spec: secretstorecsidriverv1.SecretProviderClassSpec{
						Provider: "not-supported-provider",
					},
				}

				// When
				result, err := validatingWebhook.Validate(ctx, ar, spc)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if !result.Valid {
					t.Errorf("expected valid, got invalid")
				}
			},
		},
		{
			"SuccessfullyWithCorrectSecretProviderClass",
			func(t *testing.T) {
				// Given
				spc := &secretstorecsidriverv1.SecretProviderClass{
					Spec: secretstorecsidriverv1.SecretProviderClassSpec{
						Provider: "infisical",
						Parameters: map[string]string{
							"projectSlug":         "project",
							"envSlug":             "env",
							"authSecretName":      "auth-secret",
							"authSecretNamespace": "default",
						},
					},
				}

				// When
				result, err := validatingWebhook.Validate(ctx, ar, spc)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if !result.Valid {
					t.Errorf("expected valid, got invalid")
				}
			},
		},
		{
			"FailedWithUnkonwnFields",
			func(t *testing.T) {
				// Given
				spc := &secretstorecsidriverv1.SecretProviderClass{
					Spec: secretstorecsidriverv1.SecretProviderClassSpec{
						Provider: "infisical",
						Parameters: map[string]string{
							"unknown": "unknown",
						},
					},
				}

				// When
				result, err := validatingWebhook.Validate(ctx, ar, spc)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if result.Valid {
					t.Errorf("expected invalid, got valid")
				}
				if !strings.HasPrefix(result.Message, "spec.parameters: ") {
					t.Errorf("unexpected error: %s", result.Message)
				}
			},
		},
		{
			"FailedWithoutRequiredFields",
			func(t *testing.T) {
				// Given
				spc := &secretstorecsidriverv1.SecretProviderClass{
					Spec: secretstorecsidriverv1.SecretProviderClassSpec{
						Provider:   "infisical",
						Parameters: map[string]string{},
					},
				}

				// When
				result, err := validatingWebhook.Validate(ctx, ar, spc)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if result.Valid {
					t.Errorf("expected invalid, got valid")
				}
				if !strings.HasPrefix(result.Message, "spec.parameters: ") {
					t.Errorf("unexpected error: %s", result.Message)
				}
			},
		},
		{
			"FailedWithInvalidObjectsField",
			func(t *testing.T) {
				// Given
				spc := &secretstorecsidriverv1.SecretProviderClass{
					Spec: secretstorecsidriverv1.SecretProviderClassSpec{
						Provider: "infisical",
						Parameters: map[string]string{
							"projectSlug":         "project",
							"envSlug":             "env",
							"authSecretName":      "auth-secret",
							"authSecretNamespace": "default",
							"objects":             "invalid",
						},
					},
				}

				// When
				result, err := validatingWebhook.Validate(ctx, ar, spc)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if result.Valid {
					t.Errorf("expected invalid, got valid")
				}
				if !strings.HasPrefix(result.Message, "spec.parameters.objects: ") {
					t.Errorf("unexpected error: %s", result.Message)
				}
			},
		},
		{
			"FailedWithInvalidSecretObjects",
			func(t *testing.T) {
				// Given
				spc := &secretstorecsidriverv1.SecretProviderClass{
					Spec: secretstorecsidriverv1.SecretProviderClassSpec{
						Provider: "infisical",
						Parameters: map[string]string{
							"projectSlug":         "project",
							"envSlug":             "env",
							"authSecretName":      "auth-secret",
							"authSecretNamespace": "default",
							"objects":             "",
						},
						SecretObjects: []*secretstorecsidriverv1.SecretObject{
							{
								Data: []*secretstorecsidriverv1.SecretObjectData{
									{
										ObjectName: "object1",
									},
								},
							},
						},
					},
				}

				// When
				result, err := validatingWebhook.Validate(ctx, ar, spc)

				// Then
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if result.Valid {
					t.Errorf("expected invalid, got valid")
				}
			},
		},
	} {
		ctx = context.Background()
		validatingWebhook = webhook.SecretProviderClassWebhook{}
		validatingWebhook.SetLogger(kwhlogrus.NewLogrus(logrus.NewEntry(logrus.New())))
		validatingWebhook.SetValidator(config.NewValidator())
		ar = &kwhmodel.AdmissionReview{}

		t.Run(testcase.name, testcase.f)
	}
}
