package moduleconfigreader_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	moduleconfigreader "github.com/kyma-project/modulectl/internal/service/moduleconfig/reader"
)

const (
	moduleConfigFile     = "config.yaml"
	exampleRepository    = "https://example.com/path/to/repository"
	exampleDocumentation = "https://example.com/path/to/documentation"
	exampleIcon          = "https://example.com/path/to/some-icon"
	exampleRawManifest   = "https://github.com/kyma-project/template-operator/" +
		"releases/download/1.0.1/template-operator.yaml"
)

func Test_ParseModuleConfig_ReturnsError_WhenFileReaderReturnsError(t *testing.T) {
	result, err := moduleconfigreader.ParseModuleConfig(moduleConfigFile, &fileDoesNotExistStub{})

	require.ErrorIs(t, err, errReadingFile)
	require.Nil(t, result)
}

func Test_ParseModuleConfig_Returns_CorrectModuleConfig(t *testing.T) {
	result, err := moduleconfigreader.ParseModuleConfig(moduleConfigFile, &fileExistsStub{})

	require.NoError(t, err)
	require.Equal(t, "github.com/module-name", result.Name)
	require.Equal(t, "0.0.1", result.Version)
	require.Equal(t, "https://example.com/path/to/manifests", result.Manifest.String())
	require.Equal(t, "https://example.com/path/to/defaultCR", result.DefaultCR.String())
	require.Equal(t, exampleRepository, result.Repository)
	require.Equal(t, exampleDocumentation, result.Documentation)
	require.Equal(t, "path/to/securityConfig", result.Security)
	require.Equal(t, contentprovider.Icons{
		"module-icon": exampleIcon,
	}, result.Icons)
	require.False(t, result.RequiresDowntime)
	require.Equal(t, "path/to/securityConfig", result.Security)
	require.Equal(t, map[string]string{"label1": "value1"}, result.Labels)
	require.Equal(t, map[string]string{"annotation1": "value1"}, result.Annotations)
	require.Equal(t, "networking.istio.io", result.AssociatedResources[0].Group)
	require.Equal(t, "v1alpha3", result.AssociatedResources[0].Version)
	require.Equal(t, "Gateway", result.AssociatedResources[0].Kind)
	require.Equal(t, contentprovider.Resources{
		"rawManifest": exampleRawManifest,
	}, result.Resources)
	require.Equal(t, "manager-name", result.Manager.Name)
	require.Equal(t, "manager-namespace", result.Manager.Namespace)
	require.Equal(t, "apps", result.Manager.Group)
	require.Equal(t, "v1", result.Manager.Version)
	require.Equal(t, "Deployment", result.Manager.Kind)
	require.False(t, result.Internal)
	require.False(t, result.Beta)
}

func TestNew_CalledWithNilDependencies_ReturnsErr(t *testing.T) {
	_, err := moduleconfigreader.NewService(nil)
	require.Error(t, err)
}

func Test_ValidateModuleConfig(t *testing.T) {
	exampleManifest := contentprovider.MustUrlOrLocalFile("https://example.com/path/to/manifest")
	emptyManifest := contentprovider.MustUrlOrLocalFile("")

	tests := []struct {
		name          string
		moduleConfig  *contentprovider.ModuleConfig
		expectedError error
	}{
		{
			name:          "valid module config",
			moduleConfig:  &expectedReturnedModuleConfig,
			expectedError: nil,
		},
		{
			name: "manifest file path",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      contentprovider.MustUrlOrLocalFile("./test"), // valid local file path
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: nil,
		},
		{
			name: "invalid manifest absolute file path",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      contentprovider.MustUrlOrLocalFile("/some/path/test.yaml"), // invalid absolute path
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf(
				"failed to validate manifest: must not be an absolute path: %w",
				commonerrors.ErrInvalidOption,
			),
		},
		{
			name: "default CR file name",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
				DefaultCR: contentprovider.MustUrlOrLocalFile("test"), // valid local file path
			},
			expectedError: nil,
		},
		{
			name: "invalid default CR absolute file path",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
				DefaultCR: contentprovider.MustUrlOrLocalFile("/some/path/test.yaml"), // invalid absolute path
			},
			expectedError: fmt.Errorf(
				"failed to validate default CR: must not be an absolute path: %w",
				commonerrors.ErrInvalidOption,
			),
		},
		{
			name: "invalid module name",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "invalid name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf(
				"opts.ModuleName must match the required pattern, e.g: 'github.com/path-to/your-repo': %w",
				commonerrors.ErrInvalidOption,
			),
		},
		{
			name: "invalid module version",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "invalid version",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf("opts.ModuleVersion failed to be parsed as semantic version: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "empty manifest path",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      emptyManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf("failed to validate manifest: must not be empty: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "empty repository",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    "",
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf("failed to validate repository: must not be empty: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "repository is not a URL",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    "some repository",
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf("failed to validate repository: 'some repository' is not using https scheme: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "empty documentation",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: "",
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf("failed to validate documentation: must not be empty: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "documentation is not a URL",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: "some documentation",
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf(
				"failed to validate documentation: 'some documentation' is not using https scheme: %w",
				commonerrors.ErrInvalidOption,
			),
		},
		{
			name: "empty icons",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons:         contentprovider.Icons{},
			},
			expectedError: fmt.Errorf("failed to validate module icons: must contain at least one icon: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid icon - empty name",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"": exampleIcon,
				},
			},
			expectedError: fmt.Errorf("failed to validate module icons: name must not be empty: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid icon - empty link",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": "",
				},
			},
			expectedError: fmt.Errorf("failed to validate module icons: link must not be empty: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid icon - not a URL",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": "this is not a URL",
				},
			},
			expectedError: fmt.Errorf(
				"failed to validate module icons: failed to validate link:"+
					" 'this is not a URL' is not using https scheme: %w",
				commonerrors.ErrInvalidOption,
			),
		},
		{
			name: "invalid module resources - not a URL",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
				Resources: contentprovider.Resources{
					"key": "%% not a URL",
				},
			},
			expectedError: fmt.Errorf(
				"failed to validate resources: failed to validate link: '%%%% not a URL' is not a valid URL: %w",
				commonerrors.ErrInvalidOption,
			),
		},
		{
			name: "invalid module resources - empty name",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
				Resources: contentprovider.Resources{
					"": exampleRawManifest,
				},
			},
			expectedError: fmt.Errorf("failed to validate resources: name must not be empty: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module resources - empty link",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
				Resources: contentprovider.Resources{
					"name": "",
				},
			},
			expectedError: fmt.Errorf("failed to validate resources: link must not be empty: %w",
				commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid module manifest - schema is not https",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      contentprovider.MustUrlOrLocalFile("file://path/to/manifest"),
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf(
				"failed to validate manifest: 'file://path/to/manifest' is not using https scheme: %w",
				commonerrors.ErrInvalidOption,
			),
		},
		{
			name: "invalid module defaultCR - schema is not https",
			moduleConfig: &contentprovider.ModuleConfig{
				Name:          "github.com/module-name",
				Version:       "0.0.1",
				Manifest:      exampleManifest,
				DefaultCR:     contentprovider.MustUrlOrLocalFile("file://path/to/defaultCR"),
				Repository:    exampleRepository,
				Documentation: exampleDocumentation,
				Icons: contentprovider.Icons{
					"module-icon": exampleIcon,
				},
			},
			expectedError: fmt.Errorf(
				"failed to validate default CR: 'file://path/to/defaultCR' is not using https scheme: %w",
				commonerrors.ErrInvalidOption,
			),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := moduleconfigreader.ValidateModuleConfig(test.moduleConfig)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_ValidateManager(t *testing.T) {
	tests := []struct {
		name          string
		manager       *contentprovider.Manager
		expectedError error
	}{
		{
			name:          "nil manager",
			manager:       nil,
			expectedError: nil,
		},
		{
			name: "valid manager",
			manager: &contentprovider.Manager{
				Name: "manager-name",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				Namespace: "manager-namespace",
			},
			expectedError: nil,
		},
		{
			name: "valid manager - empty namespace",
			manager: &contentprovider.Manager{
				Name: "manager-name",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			expectedError: nil,
		},
		{
			name: "invalid manager - empty name",
			manager: &contentprovider.Manager{
				Name:      "",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			expectedError: fmt.Errorf("name must not be empty: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid manager - empty kind",
			manager: &contentprovider.Manager{
				Name:      "manager-name",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
				},
			},
			expectedError: fmt.Errorf("kind must not be empty: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid manager - empty group",
			manager: &contentprovider.Manager{
				Name:      "manager-name",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			expectedError: fmt.Errorf("group must not be empty: %w", commonerrors.ErrInvalidOption),
		},
		{
			name: "invalid manager - empty version",
			manager: &contentprovider.Manager{
				Name:      "manager-name",
				Namespace: "manager-namespace",
				GroupVersionKind: metav1.GroupVersionKind{
					Kind:  "Deployment",
					Group: "apps",
				},
			},
			expectedError: fmt.Errorf("version must not be empty: %w", commonerrors.ErrInvalidOption),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := moduleconfigreader.ValidateManager(test.manager)
			if test.expectedError != nil {
				require.ErrorContains(t, err, test.expectedError.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_ValidateAssociatedResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []*metav1.GroupVersionKind
		wantErr   bool
	}{
		{
			name:      "pass on empty resources",
			resources: []*metav1.GroupVersionKind{},
			wantErr:   false,
		},
		{
			name: "pass when all resources are valid",
			resources: []*metav1.GroupVersionKind{
				{
					Group:   "networking.istio.io",
					Version: "v1alpha3",
					Kind:    "Gateway",
				},
				{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			wantErr: false,
		},
		{
			name: "fail when even one resources is invalid",
			resources: []*metav1.GroupVersionKind{
				{
					Group:   "networking.istio.io",
					Version: "v1alpha3",
					Kind:    "Gateway",
				},
				{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := moduleconfigreader.ValidateAssociatedResources(tt.resources); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAssociatedResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test Stubs

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

var expectedReturnedModuleConfig = contentprovider.ModuleConfig{
	Name:          "github.com/module-name",
	Version:       "0.0.1",
	Manifest:      contentprovider.MustUrlOrLocalFile("https://example.com/path/to/manifests"),
	Repository:    exampleRepository,
	Documentation: exampleDocumentation,
	Icons: contentprovider.Icons{
		"module-icon": exampleIcon,
	},
	RequiresDowntime: false,
	DefaultCR:        contentprovider.MustUrlOrLocalFile("https://example.com/path/to/defaultCR"),
	Security:         "path/to/securityConfig",
	Labels:           map[string]string{"label1": "value1"},
	Annotations:      map[string]string{"annotation1": "value1"},
	AssociatedResources: []*metav1.GroupVersionKind{
		{
			Group:   "networking.istio.io",
			Version: "v1alpha3",
			Kind:    "Gateway",
		},
	},
	Manager: &contentprovider.Manager{
		Name:      "manager-name",
		Namespace: "manager-namespace",
		GroupVersionKind: metav1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
	},
	Resources: contentprovider.Resources{
		"rawManifest": exampleRawManifest,
	},
	Internal: false,
	Beta:     false,
}

func (*fileExistsStub) ReadFile(_ string) ([]byte, error) {
	return yaml.Marshal(expectedReturnedModuleConfig)
}

type fileDoesNotExistStub struct{}

func (*fileDoesNotExistStub) FileExists(_ string) (bool, error) {
	return false, nil
}

var errReadingFile = errors.New("some error reading file")

func (*fileDoesNotExistStub) ReadFile(_ string) ([]byte, error) {
	return nil, errReadingFile
}
