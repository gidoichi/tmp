package webhook

import (
	"github.com/go-playground/validator/v10"
	kwhlog "github.com/slok/kubewebhook/v2/pkg/log"
)

type SecretProviderClassWebhook struct {
	secretProviderClassWebhook
}

func (w *SecretProviderClassWebhook) SetLogger(logger kwhlog.Logger) {
	w.logger = logger
}

func (w *SecretProviderClassWebhook) SetValidator(validator *validator.Validate) {
	w.validator = validator
}
