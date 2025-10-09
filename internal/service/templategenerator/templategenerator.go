package templategenerator

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"strings"
	"text/template"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/compdesc"
	"sigs.k8s.io/yaml"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

var ErrEmptyModuleConfig = errors.New("can not generate module template from empty module config")

type FileSystem interface {
	WriteFile(path, content string) error
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("fileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem: fileSystem,
	}, nil
}

const (
	modTemplate = `apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  name: {{.ResourceName}}
  namespace: ""
{{- with .Labels}}
  labels:
    {{- range $key, $value := . }}
    {{ printf "%q" $key }}: {{ printf "%q" $value }}
    {{- end}}
{{- end}} 
{{- with .Annotations}}
  annotations:
    {{- range $key, $value := . }}
    {{ printf "%q" $key }}: {{ printf "%q" $value }}
    {{- end}}
{{- end}} 
spec:
  moduleName: {{.ModuleName}}
  version: {{.ModuleVersion}}
  requiresDowntime: {{.RequiresDowntime}}
  info:
    repository: {{.Repository}}
    documentation: {{.Documentation}}
    {{- with .Icons}}
    icons:
      {{- range $key, $value := . }}
    - name: {{ $key }}
      link: {{ $value }}
      {{- end}}
    {{- end}}
{{- with .AssociatedResources}}
  associatedResources:
  {{- range .}}
  - group: {{.Group}}
    version: {{.Version}}
    kind: {{.Kind}}
  {{- end}}
{{- end}}
{{- with .Data}}
  data:
{{. | indent 4}}
{{- end}}
{{- with .Manager}}
  manager:
    name: {{.Name}}
    {{- if .Namespace}}      
    namespace: {{.Namespace}}
    {{- end}}
    group: {{.GroupVersionKind.Group}}
    version: {{.GroupVersionKind.Version}}
    kind: {{.GroupVersionKind.Kind}}
{{- end}}
{{- if .Descriptor}}
  descriptor:
{{yaml .Descriptor | printf "%s" | indent 4}}
{{- else}}
  descriptor: {}
{{- end}}
{{- with .Resources}}
  resources:
    {{- range $key, $value := . }}
  - name: {{ $key }}
    link: {{ $value }}
    {{- end}}
{{- end}}
`
)

type moduleTemplateData struct {
	ModuleName          string
	ResourceName        string
	Namespace           string
	ModuleVersion       string
	Descriptor          *compdesc.ComponentDescriptorVersion
	Repository          string
	Documentation       string
	Icons               contentprovider.Icons
	Labels              map[string]string
	Annotations         map[string]string
	Data                string
	AssociatedResources []*metav1.GroupVersionKind
	Resources           contentprovider.Resources
	Manager             *contentprovider.Manager
	RequiresDowntime    bool
}

func (s *Service) GenerateModuleTemplate(
	moduleConfig *contentprovider.ModuleConfig,
	descriptorToRender *compdesc.ComponentDescriptor,
	data []byte,
	isCrdClusterScoped bool,
	templateOutput string,
) error {
	if moduleConfig == nil {
		return ErrEmptyModuleConfig
	}

	labels := generateLabels(moduleConfig)
	annotations := generateAnnotations(moduleConfig, isCrdClusterScoped)

	ref, err := oci.ParseRef(moduleConfig.Name)
	if err != nil {
		return fmt.Errorf("failed to parse ref: %w", err)
	}
	shortName := trimShortNameFromRef(ref)
	labels[shared.ModuleName] = shortName
	moduleTemplateName := shortName + "-" + moduleConfig.Version

	moduleTemplate, err := template.New("moduleTemplate").Funcs(template.FuncMap{
		"yaml":   yaml.Marshal,
		"indent": indent,
	}).Parse(modTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse module template: %w", err)
	}

	covertedDescriptor, err := ConvertDescriptorIfNotNil(descriptorToRender)
	if err != nil {
		return err
	}

	mtData := moduleTemplateData{
		ModuleName:          shortName,
		ResourceName:        moduleTemplateName,
		ModuleVersion:       moduleConfig.Version,
		Descriptor:          covertedDescriptor,
		Repository:          moduleConfig.Repository,
		Documentation:       moduleConfig.Documentation,
		Icons:               moduleConfig.Icons,
		Labels:              labels,
		Annotations:         annotations,
		AssociatedResources: moduleConfig.AssociatedResources,
		Manager:             moduleConfig.Manager,
		RequiresDowntime:    moduleConfig.RequiresDowntime,
	}
	if moduleConfig.Manifest.IsURL() {
		mtData.Resources = contentprovider.Resources{
			// defaults rawManifest to Manifest; may be overwritten by explicitly provided entries
			"rawManifest": moduleConfig.Manifest.String(),
		}
	}

	if len(data) > 0 {
		crData, err := parseDefaultCRYaml(data)
		if err != nil {
			return fmt.Errorf("failed to parse cr data: %w", err)
		}
		mtData.Data = string(crData)
	}

	if len(moduleConfig.Resources) > 0 {
		mtData.Resources = copyEntries(mtData.Resources, moduleConfig.Resources)
	}

	w := &bytes.Buffer{}
	if err = moduleTemplate.Execute(w, mtData); err != nil {
		return fmt.Errorf("failed to execute template, %w", err)
	}

	if err = s.fileSystem.WriteFile(templateOutput, w.String()); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func ConvertDescriptorIfNotNil(
	descriptorToRender *compdesc.ComponentDescriptor,
) (*compdesc.ComponentDescriptorVersion, error) {
	var covertedDescriptor *compdesc.ComponentDescriptorVersion
	if descriptorToRender != nil {
		converted, err := compdesc.Convert(descriptorToRender)
		if err != nil {
			return nil, fmt.Errorf("failed to convert descriptor: %w", err)
		}
		covertedDescriptor = &converted
	}
	return covertedDescriptor, nil
}

func parseDefaultCRYaml(data []byte) ([]byte, error) {
	var crData map[string]interface{}
	if err := yaml.Unmarshal(data, &crData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cr data: %w", err)
	}

	cr, err := yaml.Marshal(crData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cr data yaml: %w", err)
	}

	return cr, nil
}

func generateLabels(config *contentprovider.ModuleConfig) map[string]string {
	labels := config.Labels

	if labels == nil {
		labels = make(map[string]string)
	}

	if config.Beta {
		labels[shared.BetaLabel] = shared.EnableLabelValue
	}

	if config.Internal {
		labels[shared.InternalLabel] = shared.EnableLabelValue
	}

	return labels
}

func generateAnnotations(config *contentprovider.ModuleConfig, isCrdClusterScoped bool) map[string]string {
	annotations := config.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if isCrdClusterScoped {
		annotations[shared.IsClusterScopedAnnotation] = shared.EnableLabelValue
	} else {
		annotations[shared.IsClusterScopedAnnotation] = shared.DisableLabelValue
	}
	return annotations
}

func indent(spaces int, input string) string {
	out := strings.Builder{}

	lines := strings.Split(input, "\n")

	// remove empty line at the end of the file if any
	if len(strings.TrimSpace(lines[len(lines)-1])) == 0 {
		lines = lines[:len(lines)-1]
	}

	for i, line := range lines {
		out.WriteString(strings.Repeat(" ", spaces))
		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func trimShortNameFromRef(ref oci.RefSpec) string {
	t := strings.Split(ref.Repository, "/")
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}

// copyEntries copies entries from src map to dst map, allocating dst if it is nil.
func copyEntries(dst map[string]string, src map[string]string) map[string]string {
	if len(src) > 0 {
		if dst == nil {
			dst = make(map[string]string)
		}
		maps.Copy(dst, src)
	}
	return dst
}
