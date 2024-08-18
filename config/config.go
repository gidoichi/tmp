package config

import (
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

type MountConfig struct {
	Project                  string  `json:"projectSlug" validate:"required"`
	Env                      string  `json:"envSlug" validate:"required"`
	Path                     string  `json:"secretsPath" validate:"required"`
	AuthSecretName           string  `json:"authSecretName" validate:"required"`
	AuthSecretNamespace      string  `json:"authSecretNamespace" validate:"required"`
	RawObjects               *string `json:"objects"`
	CSIPodName               string  `json:"csi.storage.k8s.io/pod.name"`
	CSIPodNamespace          string  `json:"csi.storage.k8s.io/pod.namespace"`
	CSIPodUID                string  `json:"csi.storage.k8s.io/pod.uid"`
	CSIPodServiceAccountName string  `json:"csi.storage.k8s.io/serviceAccount.name"`
	CSIEphemeral             string  `json:"csi.storage.k8s.io/ephemeral"`
	SecretProviderClass      string  `json:"secretProviderClass"`
	parsedObjects            []object
	validator                validator.Validate
}

type object struct {
	Name string `yaml:"objectName"`
}

func NewMountConfig(validator validator.Validate) *MountConfig {
	return &MountConfig{
		Path:      "/",
		validator: validator,
	}
}

func (a *MountConfig) Objects() ([]object, error) {
	if a.parsedObjects != nil {
		return a.parsedObjects, nil
	}

	if a.RawObjects == nil {
		return nil, nil
	}

	var objects []object
	if err := yaml.Unmarshal([]byte(*a.RawObjects), &objects); err != nil {
		return nil, err
	}

	a.parsedObjects = objects
	return objects, nil
}

func (a *MountConfig) Validate() error {
	if err := a.validator.Struct(a); err != nil {
		return err
	}

	if _, err := a.Objects(); err != nil {
		return err
	}

	return nil
}
