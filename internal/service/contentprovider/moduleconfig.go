package contentprovider

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
)

var ErrDuplicateMapEntries = errors.New("map contains duplicate entries")

type ModuleConfigProvider struct {
	yamlConverter ObjectToYAMLConverter
}

func NewModuleConfigProvider(yamlConverter ObjectToYAMLConverter) (*ModuleConfigProvider, error) {
	if yamlConverter == nil {
		return nil, fmt.Errorf("yamlConverter must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &ModuleConfigProvider{
		yamlConverter: yamlConverter,
	}, nil
}

func (s *ModuleConfigProvider) GetDefaultContent(args types.KeyValueArgs) (string, error) {
	if err := s.validateArgs(args); err != nil {
		return "", err
	}

	moduleConfig, err := s.getModuleConfig(args)
	if err != nil {
		return "", fmt.Errorf("failed to get module config: %w", err)
	}

	return s.yamlConverter.ConvertToYaml(*moduleConfig), nil
}

func (s *ModuleConfigProvider) getModuleConfig(args types.KeyValueArgs) (*ModuleConfig, error) {
	var manifest UrlOrLocalFile
	var defaultCR UrlOrLocalFile

	if err := manifest.FromString(args[ArgManifestFile]); err != nil {
		return nil, fmt.Errorf("invalid manifest file: %w", err)
	}
	if err := defaultCR.FromString(args[ArgDefaultCRFile]); err != nil {
		return nil, fmt.Errorf("invalid default CR file: %w", err)
	}

	return &ModuleConfig{
		Name:      args[ArgModuleName],
		Version:   args[ArgModuleVersion],
		Manifest:  manifest,
		Security:  args[ArgSecurityConfigFile],
		DefaultCR: defaultCR,
	}, nil
}

func (s *ModuleConfigProvider) validateArgs(args types.KeyValueArgs) error {
	if args == nil {
		return fmt.Errorf("args must not be nil: %w", ErrInvalidArg)
	}

	if value, ok := args[ArgModuleName]; !ok {
		return fmt.Errorf("%s: %w", ArgModuleName, ErrMissingArg)
	} else if value == "" {
		return fmt.Errorf("%s must not be empty: %w", ArgModuleName, ErrInvalidArg)
	}

	if value, ok := args[ArgModuleVersion]; !ok {
		return fmt.Errorf("%s: %w", ArgModuleVersion, ErrMissingArg)
	} else if value == "" {
		return fmt.Errorf("%s must not be empty: %w", ArgModuleVersion, ErrInvalidArg)
	}

	return nil
}

type ModuleConfig struct {
	Name                string                     `comment:"required, the name of the module"                                                                                                  yaml:"name"`
	Version             string                     `comment:"required, the version of the module"                                                                                               yaml:"version"`
	Manifest            UrlOrLocalFile             `comment:"required, reference to the manifest, must be a URL or a local file path"                                                           yaml:"manifest"`
	Repository          string                     `comment:"required, reference to the repository, must be a URL"                                                                              yaml:"repository"`
	Documentation       string                     `comment:"required, reference to the documentation, must be a URL"                                                                           yaml:"documentation"`
	Icons               Icons                      `comment:"required, icons used for UI"                                                                                                       yaml:"icons,omitempty"`
	DefaultCR           UrlOrLocalFile             `comment:"optional, reference to a YAML file containing the default CR for the module, must be a URL or a local file path"                   yaml:"defaultCR"` //nolint:tagliatelle // prefer defaultCR over defaultCr
	Security            string                     `comment:"optional, reference to a YAML file containing the security scanners config, must be a local file path"                             yaml:"security"`
	Labels              map[string]string          `comment:"optional, additional labels for the generated ModuleTemplate CR"                                                                   yaml:"labels"`
	Annotations         map[string]string          `comment:"optional, additional annotations for the generated ModuleTemplate CR"                                                              yaml:"annotations"`
	Manager             *Manager                   `comment:"optional, module resource that indicates the installation readiness of the module, typically the manager deployment of the module" yaml:"manager"`
	AssociatedResources []*metav1.GroupVersionKind `comment:"optional, optional, resources that should be cleaned up with the module deletion"                                                  yaml:"associatedResources"`
	Resources           Resources                  `comment:"optional, additional resources of the module that may be fetched"                                                                  yaml:"resources,omitempty"`
	RequiresDowntime    bool                       `comment:"optional, default=false, indicates whether the module requires downtime to support maintenance windows during module upgrades"     yaml:"requiresDowntime"`
	Internal            bool                       `comment:"optional, default=false, indicates whether the module is internal"                                                                 yaml:"internal"`
	Beta                bool                       `comment:"optional, default=false, indicates whether the module is beta"                                                                     yaml:"beta"`
}

type Manager struct {
	metav1.GroupVersionKind `comment:"required, the GVK of the manager" yaml:",inline"`

	Name      string `comment:"required, the name of the manager" yaml:"name"`
	Namespace string `comment:"optional, the path to the manager" yaml:"namespace"`
}

// Icons represents a map of icon names to links.
type Icons map[string]string

// UnmarshalYAML unmarshals Icons from YAML format.
func (i *Icons) UnmarshalYAML(unmarshal func(interface{}) error) error {
	dataMap, err := unmarshalToMap(unmarshal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Icons: %w", err)
	}
	*i = dataMap
	return nil
}

// MarshalYAML marshals Icons to YAML format.
func (i *Icons) MarshalYAML() (interface{}, error) {
	return marshalFromMap(*i)
}

// Resources represents a map of resource names to links.
type Resources map[string]string

// UnmarshalYAML unmarshals Resources from YAML format.
func (rm *Resources) UnmarshalYAML(unmarshal func(interface{}) error) error {
	dataMap, err := unmarshalToMap(unmarshal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Resources: %w", err)
	}
	*rm = dataMap
	return nil
}

// MarshalYAML marshals Resources to YAML format.
func (rm *Resources) MarshalYAML() (interface{}, error) {
	return marshalFromMap(*rm)
}

func unmarshalToMap(unmarshal func(interface{}) error) (map[string]string, error) {
	var items []nameLinkItem
	if err := unmarshal(&items); err == nil {
		resultMap := make(map[string]string)
		for _, item := range items {
			if _, exists := resultMap[item.Name]; exists {
				return nil, ErrDuplicateMapEntries
			}
			resultMap[item.Name] = item.Link
		}
		return resultMap, nil
	}

	resultMap := make(map[string]string)
	if err := unmarshal(&resultMap); err != nil {
		return nil, err
	}

	return resultMap, nil
}

func marshalFromMap(dataMap map[string]string) (interface{}, error) {
	items := make([]nameLinkItem, 0, len(dataMap))
	for name, link := range dataMap {
		items = append(items, nameLinkItem{Name: name, Link: link})
	}
	return items, nil
}

type nameLinkItem struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
}
