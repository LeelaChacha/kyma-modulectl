---
title: modulectl create
---

Creates a module bundled as an OCI artifact.

## Synopsis

Use this command to create a Kyma module, bundle it as an OCI artifact, and push it to the OCI registry optionally.

### Detailed description

This command allows you to create a Kyma module as an OCI artifact and optionally push it to the OCI registry of your choice.
For more information about Kyma modules see the [documentation](https://kyma-project.io/#/06-modules/README).

### Configuration

Provide the `--config-file` flag with a config file path.
The module config file is a YAML file used to configure the following attributes for the module:

```yaml
- name:                 a string, required, the name of the module
- version:              a string, required, the version of the module
- manifest:             a string, required, reference to the manifest, must be a URL or a local file reference: name or a relative path
- repository:           a string, required, reference to the repository, must be a URL
- documentation:        a string, required, reference to the documentation, must be a URL
- icons:                a map with string keys and values, required, icons used for UI
    - name:             a string, required, the name of the icon
      link:             a URL, required, the link to the icon
- defaultCR:            a string, optional, reference to a YAML file containing the default CR for the module, must be a URL or a local file reference: name or a relative path
- security:             a string, optional, reference to a YAML file containing the security scanners config, must be a local file path
- labels:               a map with string keys and values, optional, additional labels for the generated ModuleTemplate CR
- annotations:          a map with string keys and values, optional, additional annotations for the generated ModuleTemplate CR
- manager:              an object, optional, module resource that indicates the installation readiness of the module, typically the manager deployment of the module
    name:               a string, required, the name of the module resource
    namespace:          a string, optional, the namespace of the module resource
    group:              a string, required, the API group of the module resource
    version:            a string, required, the API version of the module resource
    kind:               a string, required, the API kind of the module resource
- associatedResources:  a list of Group-Version-Kind(GVK), optional, resources that should be cleaned up with the module deletion
- resources:            a map with string keys and values, optional, additional resources of the module that may be fetched
    - name:             a string, required, the name of the resource
      link:             a URL, required, the link to the resource
- requiresDowntime:     a boolean, optional, default=false, indicates whether the module requires downtime to support maintenance windows during module upgrades
- namespace:            a string, optional, default=kcp-system, the namespace where the ModuleTemplate will be deployed
- internal:             a boolean, optional, default=false, indicates whether the module is internal
- beta:                 a boolean, optional, default=false, indicates whether the module is beta
```

The file referenced by the **manifest** attribute contains all the module's resources in a single, multi-document YAML file. These resources will be created in the Kyma cluster when the module is activated. If the attribute is a file name or a relative path, modulectl resolves its location relative to the module config file location. If it is a URL, it must be accessible from the machine where the command is executed.
The file referenced by the **defaultCR** attribute contains a default custom resource for the module that is installed along with the module. It is additionally schema-validated against the Custom Resource Definition. If the attribute is a file name or a relative path, modulectl resolves its location relative to the module config file location. If it is a URL, it must be accessible from the machine where the command is executed.
The CRD used for the validation must exist in the set of the module's resources.
The **resources** are copied to the ModuleTemplate **spec.resources**. If it does not have an entry named 'raw-manifest', the ModuleTemplate **spec.resources** populates this entry from the **manifest** field specified in the module config file.

### Modules as OCI artifacts
Modules are built and distributed as OCI artifacts. 
This command creates a component descriptor in the configured descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI artifact.
The internal structure of the artifact conforms to the [Open Component Model](https://ocm.software/) scheme version 3.

If you configured the "--registry" flag, the created module is validated and pushed to the configured registry.


```bash
modulectl create [--config-file MODULE_CONFIG_FILE] [--registry MODULE_REGISTRY] [flags]
```

## Examples

```bash
Build a simple module and push it to a remote registry
		modulectl create --config-file=/path/to/module-config-file --registry http://localhost:5001/unsigned --insecure
```

## Flags

```bash
-c, --config-file string                    Specifies the path to the module configuration file.
    --disable-ocm-registry-push             Disables the push of the component version to the OCM registry.
    --dry-run                               Skips the push of the module descriptor to the registry. Checks if the component version already exists in the registry and fails the command if it does and --overwrite is not set to true.
-h, --help                                  Provides help for the create command.
    --insecure                              Allows to use a less secure (non-tls) connection for registry access, e.g. localhost when testing. Should only be used in dev scenarios.
    --module-sources-git-directory string   Path to the directory containing the module sources. If not set, the current directory is used. The directory must contain a valid Git repository.
-o, --output string                         Path to write the ModuleTemplate file to, if the module is uploaded to a registry (default "template.yaml").
    --output-constructor-file string        Path to write the component constructor file to (default "component-constructor.yaml").
    --overwrite                             Overwrites the pushed component version if it already exists in the OCI registry. Use the flag ONLY for testing purposes.
-r, --registry string                       Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor.
    --registry-credentials string           Basic authentication credentials for the given repository in the <user:password> format.
    --skip-version-validation               Skipping image and ocm version validation
```

## See also

* [modulectl](modulectl.md)	 - Command line tool for creating Kyma modules.

