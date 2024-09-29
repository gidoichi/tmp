package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/gidoichi/secrets-store-csi-driver-provider-infisical/config"
	"github.com/go-playground/validator/v10"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhwebhook "github.com/slok/kubewebhook/v2/pkg/webhook"
	kwhvalidating "github.com/slok/kubewebhook/v2/pkg/webhook/validating"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	secretstorecsidriverv1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

const (
	InfisicalSecretProviderName = "infisical"
)

type secretProviderClassWebhook struct {
	logger    kwhlog.Logger
	validator *validator.Validate
}

var _ kwhvalidating.Validator = &secretProviderClassWebhook{}

// NewSecretProviderClassValidatingWebhook returns a new secretproviderclass validating webhook.
func NewSecretProviderClassValidatingWebhook(logger kwhlog.Logger) (kwhwebhook.Webhook, error) {
	// Create validators.
	validators := []kwhvalidating.Validator{
		&secretProviderClassWebhook{
			logger:    logger,
			validator: config.NewValidator(),
		},
	}

	return kwhvalidating.NewWebhook(kwhvalidating.WebhookConfig{
		ID:        "secretproviderclass-validator",
		Obj:       &secretstorecsidriverv1.SecretProviderClass{},
		Validator: kwhvalidating.NewChain(logger, validators...),
		Logger:    logger,
	})
}

func (w *secretProviderClassWebhook) Validate(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhvalidating.ValidatorResult, error) {
	spc, ok := obj.(*secretstorecsidriverv1.SecretProviderClass)
	if !ok {
		// If not a secretproviderclass just continue the validation chain(if there is one) and don't do nothing.
		return w.validateSkip()
	}
	if spc.Spec.Provider != InfisicalSecretProviderName {
		return w.validateSkip()
	}

	path := "spec.parameters"

	mountConfig := config.NewMountConfig(*w.validator)
	attributes, err := json.Marshal(spc.Spec.Parameters)
	if err != nil {
		return w.validateFailed(config.NewConfigError(path, err))
	}
	attributesDecoder := json.NewDecoder(bytes.NewReader(attributes))
	attributesDecoder.DisallowUnknownFields()
	if err := attributesDecoder.Decode(mountConfig); err != nil {
		return w.validateFailed(config.NewConfigError(path, err))
	}

	if err := mountConfig.Validate(); err != nil {
		return w.validateFailed(config.NewConfigError(path, err))
	}

	if _, found := spc.Spec.Parameters["objects"]; !found {
		return w.validateSucceeded()
	}

	path = "spec.parameters.objects"

	objects, err := mountConfig.Objects()
	if err != nil {
		return w.validateFailed(config.NewConfigError(path, err))
	}

	path = "spec.secretObjects"

	var objectNames []string
	for _, object := range objects {
		objectNames = append(objectNames, object.Name)
	}
	var errs error
	for sindex, secretObject := range spc.Spec.SecretObjects {
		for dindex, data := range secretObject.Data {
			if !slices.Contains(objectNames, data.ObjectName) {
				err := errors.New(data.ObjectName)
				err = config.NewConfigError(fmt.Sprintf(path+"[%d].data[%d].objectName", sindex, dindex), err)
				errs = errors.Join(errs, err)
			}
		}
	}
	if errs != nil {
		err := fmt.Errorf("not found in spec.parameters.objects: %w", errs)
		return w.validateFailed(err)
	}

	return w.validateSucceeded()
}

func (w *secretProviderClassWebhook) validateSkip() (*kwhvalidating.ValidatorResult, error) {
	return w.validateSucceeded()
}

func (w *secretProviderClassWebhook) validateSucceeded() (*kwhvalidating.ValidatorResult, error) {
	return &kwhvalidating.ValidatorResult{
		Valid: true,
	}, nil
}

func (w *secretProviderClassWebhook) validateFailed(err error) (*kwhvalidating.ValidatorResult, error) {
	return &kwhvalidating.ValidatorResult{
		Valid:   false,
		Message: err.Error(),
	}, nil
}
