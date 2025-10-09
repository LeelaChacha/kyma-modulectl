package moduleconfiggenerator_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/moduleconfig"
	moduleconfiggenerator "github.com/kyma-project/modulectl/internal/service/moduleconfig/generator"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

const (
	directory        = "./bin/dir"
	moduleConfigFile = "config.yaml"
)

func Test_NewService_ReturnsError_WhenFileSystemIsNil(t *testing.T) {
	_, err := moduleconfiggenerator.NewService(
		nil,
		&fileGeneratorErrorStub{},
	)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "fileSystem must not be nil")
}

func Test_NewService_ReturnsError_WhenFileGeneratorIsNil(t *testing.T) {
	_, err := moduleconfiggenerator.NewService(
		&errorStub{},
		nil,
	)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "fileGenerator must not be nil")
}

func Test_ForceExplicitOverwrite_ReturnsError_WhenFilesystemReturnsError(t *testing.T) {
	svc, _ := moduleconfiggenerator.NewService(
		&errorStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, true)

	require.ErrorIs(t, result, errSomeOSError)
}

func Test_ForceExplicitOverwrite_ReturnsError_WhenFileExistsAndNoOverwrite(t *testing.T) {
	svc, _ := moduleconfiggenerator.NewService(
		&fileExistsStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, false)

	require.ErrorIs(t, result, moduleconfig.ErrFileExists)
}

func Test_ForceExplicitOverwrite_ReturnsNil_WhenFileExistsAndOverwrite(t *testing.T) {
	svc, _ := moduleconfiggenerator.NewService(
		&fileExistsStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, true)

	require.NoError(t, result)
}

func Test_ForceExplicitOverwrite_ReturnsNil_WhenFileDoesNotExist(t *testing.T) {
	svc, _ := moduleconfiggenerator.NewService(
		&fileDoesNotExistStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.ForceExplicitOverwrite(directory, moduleConfigFile, true)

	require.NoError(t, result)
}

func Test_GenerateFile_ReturnsError_WhenFileGeneratorReturnsError(t *testing.T) {
	svc, _ := moduleconfiggenerator.NewService(
		&errorStub{},
		&fileGeneratorErrorStub{},
	)

	result := svc.GenerateFile(nil, moduleConfigFile, types.KeyValueArgs{})

	require.ErrorIs(t, result, errSomeFileGeneratorError)
}

func Test_GenerateFile_Succeeds(t *testing.T) {
	svc, _ := moduleconfiggenerator.NewService(
		&errorStub{},
		&fileGeneratorStub{},
	)

	result := svc.GenerateFile(nil, moduleConfigFile, types.KeyValueArgs{})

	require.NoError(t, result)
}

// Test Stubs

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

func (*fileExistsStub) ReadFile(_ string) ([]byte, error) {
	manifest := contentprovider.MustUrlOrLocalFile("path/to/manifests")
	defaultCR := contentprovider.MustUrlOrLocalFile("path/to/defaultCR")

	moduleConfig := contentprovider.ModuleConfig{
		Name:             "module-name",
		Version:          "0.0.1",
		Manifest:         manifest,
		RequiresDowntime: false,
		DefaultCR:        defaultCR,
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
		Resources: contentprovider.Resources{},
	}

	return yaml.Marshal(moduleConfig)
}

type fileDoesNotExistStub struct{}

func (*fileDoesNotExistStub) FileExists(_ string) (bool, error) {
	return false, nil
}

var errReadingFile = errors.New("some error reading file")

func (*fileDoesNotExistStub) ReadFile(_ string) ([]byte, error) {
	return nil, errReadingFile
}

var errSomeOSError = errors.New("some OS error")

type errorStub struct{}

func (*errorStub) FileExists(_ string) (bool, error) {
	return false, errSomeOSError
}

func (*errorStub) ReadFile(_ string) ([]byte, error) {
	return nil, nil
}

type fileGeneratorStub struct{}

func (*fileGeneratorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return nil
}

type fileGeneratorErrorStub struct{}

var errSomeFileGeneratorError = errors.New("some file generator error")

func (*fileGeneratorErrorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return errSomeFileGeneratorError
}
