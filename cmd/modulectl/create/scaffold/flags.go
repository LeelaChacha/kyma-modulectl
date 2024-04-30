package scaffold

import (
	"github.com/kyma-project/modulectl/internal/scaffold"
	"github.com/spf13/pflag"
)

const (
	DirectoryFlagName    = "directory"
	directoryFlagShort   = "d"
	DirectoryFlagDefault = "./"
	directoryFlagUsage   = "Specifies the target directory where the scaffolding shall be generated"

	ModuleConfigFileFlagName    = "module-config"
	ModuleConfigFileFlagDefault = "scaffold-module-config.yaml"
	moduleConfigFileFlagUsage   = "Specifies the name of the generated module configuration file"

	ModuleConfigFileOverwriteFlagName    = "overwrite"
	moduleConfigFileOverwriteFlagShort   = "o"
	ModuleConfigFileOverwriteFlagDefault = false
	moduleConfigFileOverwriteFlagUsage   = "Specifies if the command overwrites an existing module configuration file"

	ManifestFileFlagName    = "gen-manifest"
	ManifestFileFlagDefault = "manifest.yaml"
	manifestFileFlagUsage   = "Specifies the manifest in the generated module config. A blank manifest file is generated if it doesn't exist"

	DefaultCRFlagName    = "gen-default-cr"
	DefaultCRFlagDefault = "default-cr.yaml"
	defaultCRFlagUsage   = "Specifies the default CR in the generated module config. A blank default CR file is generated if it doesn't exist"

	SecurityConfigFileFlagName    = "gen-security-config"
	SecurityConfigFileFlagDefault = "sec-scanners-config.yaml"
	securityConfigFileFlagUsage   = "Specifies the security file in the generated module config. A scaffold security config file is generated if it doesn't exist"

	ModuleNameFlagName    = "module-name"
	ModuleNameFlagDefault = "kyma-project.io/module/mymodule"
	moduleNameFlagUsage   = "Specifies the module name in the generated config file"

	ModuleVersionFlagName    = "module-version"
	ModuleVersionFlagDefault = "0.0.1"
	moduleVersionFlagUsage   = "Specifies the module version in the generated module config file"

	ModuleChannelFlagName    = "module-channel"
	ModuleChannelFlagDefault = "regular"
	moduleChannelFlagUsage   = "Specifies the module channel in the generated module config file"
)

func parseFlags(flags *pflag.FlagSet, opts *scaffold.Options) error {
	flags.StringVarP(&opts.Directory, DirectoryFlagName, directoryFlagShort, DirectoryFlagDefault, directoryFlagUsage)
	flags.StringVar(&opts.ModuleConfigFileName, ModuleConfigFileFlagName, ModuleConfigFileFlagDefault, moduleConfigFileFlagUsage)
	flags.BoolVarP(&opts.ModuleConfigFileOverwrite, ModuleConfigFileOverwriteFlagName, moduleConfigFileOverwriteFlagShort, ModuleConfigFileOverwriteFlagDefault, moduleConfigFileOverwriteFlagUsage)
	flags.StringVar(&opts.ManifestFileName, ManifestFileFlagName, ManifestFileFlagDefault, manifestFileFlagUsage)
	flags.StringVar(&opts.DefaultCRFileName, DefaultCRFlagName, DefaultCRFlagDefault, defaultCRFlagUsage)
	flags.StringVar(&opts.SecurityConfigFileName, SecurityConfigFileFlagName, SecurityConfigFileFlagDefault, securityConfigFileFlagUsage)
	flags.StringVar(&opts.ModuleName, ModuleNameFlagName, ModuleNameFlagDefault, moduleNameFlagUsage)
	flags.StringVar(&opts.ModuleVersion, ModuleVersionFlagName, ModuleVersionFlagDefault, moduleVersionFlagUsage)
	flags.StringVar(&opts.ModuleChannel, ModuleChannelFlagName, ModuleChannelFlagDefault, moduleChannelFlagUsage)

	return nil
}
