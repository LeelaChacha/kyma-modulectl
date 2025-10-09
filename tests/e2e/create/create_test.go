//go:build e2e

package create_test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"k8s.io/apimachinery/pkg/util/yaml"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"

	"github.com/kyma-project/modulectl/internal/common"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	moduleVersion = "1.0.3"
)

var _ = Describe("Test 'create' command", Ordered, func() {
	BeforeEach(func() {
		for _, file := range filesIn("/tmp/") {
			if file == "template.yaml" {
				err := os.Remove(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
			}
		}

		_, err := exec.Command("k3d", "registry", "delete", "--all").CombinedOutput()
		Expect(err).ToNot(HaveOccurred())
		_, err = exec.Command("k3d", "registry", "create", "oci.localhost", "--port",
			"5001").CombinedOutput()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked without config-file arg", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})

		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to read file module-config.yaml: open module-config.yaml: no such file or directory"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked without registry arg", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})

		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("opts.RegistryURL must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing name", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          missingNameConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("opts.ModuleName must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing version", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          missingVersionConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("opts.ModuleVersion must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing manifest", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          missingManifestConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate manifest: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing repository", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          missingRepositoryConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing documentation", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          missingDocumentationConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https repository", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          nonHttpsRepository,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: 'http://github.com/kyma-project/template-operator' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https documentation", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          nonHttpsDocumentation,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: 'http://github.com/kyma-project/template-operator/blob/main/README.md' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing icons", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          missingIconsConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: must contain at least one icon: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with duplicate entry in icons", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          duplicateIcons,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Icons: map contains duplicate entries"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid icon - link missing", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          iconsWithoutLink,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: link must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid icon - name missing", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          iconsWithoutName,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: name must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with duplicate entry in resources", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          duplicateResources,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Resources: map contains duplicate entries"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https resource", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          nonHttpsResource,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: failed to validate link: 'http://some.other/location/template-operator.yaml' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid resource - link missing", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          resourceWithoutLink,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: link must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid resource - name missing", func() {
			cmd = createCmd{
				registry:                  ociRegistry,
				moduleConfigFile:          resourceWithoutName,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: name must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with existing '--registry' and missing '--insecure' flag", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("could not push"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with a non git directory for module-sources-git-directory arg", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				moduleSourcesGitDirectory: "/tmp/not-a-git-dir",
			}
		})

		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("currently configured module-sources-git-directory \"/tmp/not-a-git-dir\" must point to a valid git repository: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config and dry-run flag", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				dryRun:                    true,
				skipVersionValidation:     false,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)

				validateMinimalModuleTemplate(template, descriptor)

				By("And descriptor.component.repositoryContexts should be empty")
				Expect(descriptor.RepositoryContexts).To(HaveLen(0))

				By("And descriptor.component.resources should have one from manifest")
				Expect(descriptor.Resources).To(HaveLen(1))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)

				validateMinimalModuleTemplate(template, descriptor)

				By("And descriptor.component.repositoryContexts should be correct")
				Expect(descriptor.RepositoryContexts).To(HaveLen(1))
				repo := descriptor.GetEffectiveRepositoryContext()
				Expect(repo.Object["baseUrl"]).To(Equal(ociRegistry))
				Expect(repo.Object["componentNameMapping"]).To(Equal(string(ocireg.OCIRegistryURLPathMapping)))
				Expect(repo.Object["type"]).To(Equal(ocireg.Type))

				By("And descriptor.component.resources should be correct")
				Expect(descriptor.Resources).To(HaveLen(2))
				resource := descriptor.Resources[0]
				Expect(resource.Name).To(Equal("template-operator"))
				Expect(resource.Relation).To(Equal(ocmv1.ExternalRelation))
				Expect(resource.Type).To(Equal("ociArtifact"))
				Expect(resource.Version).To(Equal(moduleVersion))
				resource = descriptor.Resources[1]
				Expect(resource.Name).To(Equal("raw-manifest"))
				Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
				Expect(resource.Type).To(Equal("directoryTree"))
				Expect(resource.Version).To(Equal(moduleVersion))

				By("And descriptor.component.resources[0].access should be correct")
				resourceAccessSpec0, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[0].Access)
				Expect(err).ToNot(HaveOccurred())
				ociartifactAccessSpec0, ok := resourceAccessSpec0.(*ociartifact.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(ociartifactAccessSpec0.GetType()).To(Equal(ociartifact.Type))

				By("And descriptor.component.resources[1].access should be correct")
				resourceAccessSpec2, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[1].Access)
				Expect(err).ToNot(HaveOccurred())
				localBlobAccessSpec2, ok := resourceAccessSpec2.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(localBlobAccessSpec2.GetType()).To(Equal(localblob.Type))
				Expect(localBlobAccessSpec2.LocalReference).To(ContainSubstring("sha256:"))
				Expect(localBlobAccessSpec2.MediaType).To(Equal("application/x-tar"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
			Expect(cmd.execute()).To(Succeed())
		})
		By("Then invoked with same version that already exists in the registry", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail with same version exists message", func() {
			err := cmd.execute()
			Expect(
				err.Error(),
			).Should(ContainSubstring(fmt.Sprintf("could not push component version: cannot push component version %s: component version already exists, cannot push the new version",
				moduleVersion)))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
			Expect(cmd.execute()).To(Succeed())
		})
		By("When invoked with same version that already exists in the registry and dry-run flag", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				dryRun:                    true,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail with same version exists message", func() {
			err := cmd.execute()
			Expect(
				err.Error(),
			).Should(ContainSubstring(fmt.Sprintf("component kyma-project.io/module/template-operator in version %s already exists: component version already exists",
				moduleVersion)))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with same version that already exists in the registry, and dry-run flag, and overwrite flag",
			func() {
				cmd = createCmd{
					moduleConfigFile:          minimalConfig,
					registry:                  ociRegistry,
					insecure:                  true,
					output:                    templateOutputPath,
					overwrite:                 true,
					dryRun:                    true,
					moduleSourcesGitDirectory: templateOperatorPath,
				}
			})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)

				validateMinimalModuleTemplate(template, descriptor)

				By("And descriptor.component.repositoryContexts should be empty")
				Expect(descriptor.RepositoryContexts).To(HaveLen(0))

				By("And descriptor.component.resources should have only from raw manifest entry")
				Expect(descriptor.Resources).To(HaveLen(1))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with same version that already exists in the registry and overwrite flag", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				overwrite:                 true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			err := cmd.execute()
			Expect(err).Should(Succeed())
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing annotations and different version", func() {
			cmd = createCmd{
				moduleConfigFile:          withAnnotationsConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And new annotation should be correctly added")
				annotations := template.Annotations
				Expect(annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))
				Expect(annotations["operator.kyma-project.io/doc-url"]).To(Equal("https://kyma-project.io"))

				By("And descriptor.component.resources should be correct")
				resource := descriptor.Resources[0]
				Expect(resource.Version).To(Equal(moduleVersion))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing default-cr and different version", func() {
			cmd = createCmd{
				moduleConfigFile:          withDefaultCrConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And descriptor.component.resources should be correct")
				Expect(descriptor.Resources).To(HaveLen(3))
				resource := descriptor.Resources[2]
				Expect(resource.Name).To(Equal("default-cr"))
				Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
				Expect(resource.Type).To(Equal("directoryTree"))
				Expect(resource.Version).To(Equal(moduleVersion))

				By("And descriptor.component.resources[2].access should be correct")
				defaultCRResourceAccessSpec, err := ocm.DefaultContext().
					AccessSpecForSpec(descriptor.Resources[2].Access)
				Expect(err).ToNot(HaveOccurred())
				defaultCRAccessSpec, ok := defaultCRResourceAccessSpec.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(defaultCRAccessSpec.GetType()).To(Equal(localblob.Type))
				Expect(defaultCRAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
				Expect(defaultCRAccessSpec.MediaType).To(Equal("application/x-tar"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing security-scanner-config and different version",
			func() {
				cmd = createCmd{
					moduleConfigFile:          withSecurityConfig,
					registry:                  ociRegistry,
					insecure:                  true,
					output:                    templateOutputPath,
					moduleSourcesGitDirectory: templateOperatorPath,
				}
			})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And descriptor.component.resources should be correct")
				Expect(descriptor.Resources).To(HaveLen(3))

				resource := findResourceByNameVersionType(descriptor.Resources, "template-operator", moduleVersion,
					"ociArtifact")
				Expect(resource).ToNot(BeNil())
				Expect(resource.Relation).To(Equal(ocmv1.ExternalRelation))

				resourceAccessSpec0, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
				Expect(err).ToNot(HaveOccurred())
				ociArtifactAccessSpec, ok := resourceAccessSpec0.(*ociartifact.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(ociArtifactAccessSpec.GetType()).To(Equal(ociartifact.Type))
				Expect(
					ociArtifactAccessSpec.ImageReference,
				).To(Equal(fmt.Sprintf("europe-docker.pkg.dev/kyma-project/prod/template-operator:%s",
					moduleVersion)))

				resource = findResourceByNameVersionType(descriptor.Resources, "template-operator", "2.0.0",
					"ociArtifact")
				Expect(resource).ToNot(BeNil())
				Expect(resource.Relation).To(Equal(ocmv1.ExternalRelation))

				resourceAccessSpec1, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
				Expect(err).ToNot(HaveOccurred())
				ociArtifactAccessSpec, ok = resourceAccessSpec1.(*ociartifact.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(ociArtifactAccessSpec.GetType()).To(Equal(ociartifact.Type))
				Expect(
					ociArtifactAccessSpec.ImageReference,
				).To(Equal("europe-docker.pkg.dev/kyma-project/prod/template-operator:2.0.0"))

				resource = findResourceByNameVersionType(descriptor.Resources, "raw-manifest", moduleVersion,
					"directoryTree")
				Expect(resource).ToNot(BeNil())
				resourceAccessSpec3, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
				Expect(err).ToNot(HaveOccurred())
				localBlobAccessSpec2, ok := resourceAccessSpec3.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(localBlobAccessSpec2.GetType()).To(Equal(localblob.Type))
				Expect(localBlobAccessSpec2.LocalReference).To(ContainSubstring("sha256:"))
				Expect(localBlobAccessSpec2.MediaType).To(Equal("application/x-tar"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid module-config containing not existing security-scanner-config",
			func() {
				cmd = createCmd{
					moduleConfigFile:          invalidSecurityConfig,
					registry:                  ociRegistry,
					insecure:                  true,
					output:                    templateOutputPath,
					moduleSourcesGitDirectory: templateOperatorPath,
				}
			})
		By("Then the command should succeed", func() {
			err := cmd.execute()

			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to configure security scanners: failed to get security config: failed to parse security config data: security config file does not exist"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing manager field and different version", func() {
			cmd = createCmd{
				moduleConfigFile:          withManagerConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And spec.manager should be correct")
				manager := template.Spec.Manager
				Expect(manager).ToNot(BeNil())
				Expect(manager.Name).To(Equal("template-operator-controller-manager"))
				Expect(manager.Namespace).To(Equal("template-operator-system"))
				Expect(manager.Version).To(Equal("v1"))
				Expect(manager.Group).To(Equal("apps"))
				Expect(manager.Kind).To(Equal("Deployment"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing manager field without namespace and different version",
			func() {
				cmd = createCmd{
					moduleConfigFile:          withNoNamespaceManagerConfig,
					registry:                  ociRegistry,
					insecure:                  true,
					output:                    templateOutputPath,
					moduleSourcesGitDirectory: templateOperatorPath,
				}
			})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And spec.manager should be correct")
				manager := template.Spec.Manager
				Expect(manager).ToNot(BeNil())
				Expect(manager.Name).To(Equal("template-operator-controller-manager"))
				Expect(manager.Namespace).To(BeEmpty())
				Expect(manager.Version).To(Equal("v1"))
				Expect(manager.Group).To(Equal("apps"))
				Expect(manager.Kind).To(Equal("Deployment"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing associatedResources list", func() {
			cmd = createCmd{
				moduleConfigFile:          withAssociatedResourcesConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And spec.associatedResources should be correct")
				resources := template.Spec.AssociatedResources
				Expect(resources).ToNot(BeEmpty())
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].Group).To(Equal("networking.istio.io"))
				Expect(resources[0].Version).To(Equal("v1alpha3"))
				Expect(resources[0].Kind).To(Equal("Gateway"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config containing resources", func() {
			cmd = createCmd{
				moduleConfigFile:          withResources,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain merged .spec.resources", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(template.Spec.Resources).To(HaveLen(2))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(
					template.Spec.Resources[0].Link,
				).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
					moduleVersion)))
				Expect(template.Spec.Resources[1].Name).To(Equal("someResource"))
				Expect(template.Spec.Resources[1].Link).To(Equal("https://some.other/location/template-operator.yaml"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config containing rawManfiest in resources", func() {
			cmd = createCmd{
				moduleConfigFile:          withResourcesOverwrite,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain rawManifest value from module-config", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(template.Spec.Resources).To(HaveLen(1))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(template.Spec.Resources[0].Link).To(Equal("https://some.other/location/template-operator.yaml"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with manifest being a local file reference", func() {
			cmd = createCmd{
				moduleConfigFile:          manifestFileref,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				validateTemplateWithFileReference(template, descriptor, moduleVersion)

				By("And template's spec.resources should NOT contain rawManifest")
				Expect(template.Spec.Resources).To(HaveLen(0))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with default CR being a fileref", func() {
			cmd = createCmd{
				moduleConfigFile:          defaultCRFileref,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				validateTemplateWithFileReference(template, descriptor, moduleVersion)

				By("And template's spec.resources should contain rawManifest")
				Expect(template.Spec.Resources).To(HaveLen(1))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(
					template.Spec.Resources[0].Link,
				).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
					moduleVersion)))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing requiresDowntime true and different version", func() {
			cmd = createCmd{
				moduleConfigFile:          withRequiresDowntimeConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And module template should have spec.requiresDowntime set to true")
				Expect(template.Spec.RequiresDowntime).To(BeTrue())
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing internal true and different version", func() {
			cmd = createCmd{
				moduleConfigFile:          withInternalConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And module template should have operator.kyma-project.io/internal label set to true")
				val, ok := template.Labels[shared.InternalLabel]
				Expect(val).To(Equal("true"))
				Expect(ok).To(BeTrue())
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing beta true and different version", func() {
			cmd = createCmd{
				moduleConfigFile:          withBetaConfig,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And module template should have operator.kyma-project.io/beta label set to true")
				val, ok := template.Labels[shared.BetaLabel]
				Expect(val).To(Equal("true"))
				Expect(ok).To(BeTrue())
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images from both manifest and security config", func() {
			cmd = createCmd{
				moduleConfigFile:          withManifestAndSecurity,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain merged and deduplicated images", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResourcesMap(descriptor)

				Expect(len(imageResources)).To(Equal(5), "Expected exactly 5 image resources")

				expectedImages := map[string]struct {
					reference string
					version   string
				}{
					"template-operator": {"europe-docker.pkg.dev/kyma-project/prod/template-operator:1.0.3", "1.0.3"},
					"webhook":           {"europe-docker.pkg.dev/kyma-project/prod/webhook:v1.2.0", "v1.2.0"},
					"nginx":             {"nginx:1.25.0", "1.25.0"},
					"postgres":          {"postgres:15.3", "0.0.0-15.3"},
					"static-c7742da0": {
						"gcr.io/distroless/static@sha256:c7742da01aa7ee169d59e58a91c35da9c13e67f555dcd8b2ada15887aa619e6c",
						"0.0.0+sha256.c7742da01aa7",
					},
				}

				for imageName, expected := range expectedImages {
					err := verifyImageResource(imageResources, imageName, expected.reference, expected.version)
					Expect(err).ToNot(HaveOccurred(), "Failed verification for image: %s", imageName)
				}

				templateOperatorCount := 0
				for _, resource := range imageResources {
					if imageRef, err := getImageReference(resource); err == nil {
						if imageRef == "europe-docker.pkg.dev/kyma-project/prod/template-operator:1.0.3" {
							templateOperatorCount++
						}
					}
				}
				Expect(templateOperatorCount).To(Equal(1), "template-operator:1.0.3 should appear exactly once")
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images in containers", func() {
			cmd = createCmd{
				moduleConfigFile:          withManifestContainers,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed and extract images from containers", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images from containers", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResourcesMap(descriptor)

				expectedImages := map[string]struct {
					reference string
					version   string
				}{
					"template-operator": {"europe-docker.pkg.dev/kyma-project/prod/template-operator:1.0.3", "1.0.3"},
					"webhook":           {"europe-docker.pkg.dev/kyma-project/prod/webhook:v1.2.0", "v1.2.0"},
					"nginx":             {"nginx:1.25.0", "1.25.0"},
				}

				Expect(len(imageResources)).To(Equal(len(expectedImages)), "Expected exactly %d image resources",
					len(expectedImages))

				for imageName, expected := range expectedImages {
					err := verifyImageResource(imageResources, imageName, expected.reference, expected.version)
					Expect(err).ToNot(HaveOccurred(), "Failed verification for image: %s", imageName)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images in initContainers", func() {
			cmd = createCmd{
				moduleConfigFile:          withManifestInitContainers,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed and extract images from initContainers", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images from initContainers", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResourcesMap(descriptor)

				expectedImages := map[string]struct {
					reference string
					version   string
				}{
					"busybox": {"busybox:1.35.0", "1.35.0"},
					"migrate": {"migrate/migrate:v4.16.0", "v4.16.0"},
					"alpine":  {"alpine:3.18.0", "3.18.0"},
				}

				Expect(len(imageResources)).To(Equal(len(expectedImages)),
					"Expected exactly %d instead of 2 images, as the alpine image is from containers, and it's not possible to have initContainers without containers for a valid k8s deployment/statefulset yaml file",
					len(expectedImages))

				for imageName, expected := range expectedImages {
					err := verifyImageResource(imageResources, imageName, expected.reference, expected.version)
					Expect(err).ToNot(HaveOccurred(), "Failed verification for image: %s", imageName)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images with SHA digest", func() {
			cmd = createCmd{
				moduleConfigFile:          withManifestShaDigest,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed and extract images with SHA digest", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images with SHA digest", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResourcesMap(descriptor)

				expectedImages := map[string]struct {
					reference string
					version   string
				}{
					"nginx-fff07cc3": {
						"nginx@sha256:fff07cc3a741c20b2b1e4bbc3bbd6d3c84859e5116fce7858d3d176542800c10",
						"0.0.0+sha256.fff07cc3a741",
					},
					"alpine": {"alpine:3.18.0", "3.18.0"},
				}

				Expect(len(imageResources)).To(Equal(len(expectedImages)), "Expected exactly %d image resources",
					len(expectedImages))

				for imageName, expected := range expectedImages {
					err := verifyImageResource(imageResources, imageName, expected.reference, expected.version)
					Expect(err).ToNot(HaveOccurred(), "Failed verification for image: %s", imageName)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images from env variables", func() {
			cmd = createCmd{
				moduleConfigFile:          withManifestEnvVariables,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed and extract images from env variables", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images from env variables", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResourcesMap(descriptor)
				Expect(len(imageResources)).To(Equal(4), "Expected exactly 4 image resources from env variables")

				err = verifyImageResource(imageResources, "webhook",
					"europe-docker.pkg.dev/kyma-project/prod/webhook:v1.2.0", "v1.2.0")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with manifest containing latest/main tags", func() {
			cmd = createCmd{
				moduleConfigFile:          withManifestLatestMainTags,
				registry:                  ociRegistry,
				insecure:                  true,
				output:                    templateOutputPath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail with specific error", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("image tag is disallowed"))

			Expect(err.Error()).Should(Or(
				ContainSubstring("latest"),
				ContainSubstring("main"),
			))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with --disable-ocm-registry-push flag", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				disableOCMRegistryPush:    true,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And component constructor file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("component-constructor.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())

				By("And template should have no descriptor")
				Expect(string(template.Spec.Descriptor.Raw)).To(MatchJSON(`{}`))

				By("And template should have basic info")
				Expect(template.Spec.ModuleName).To(Equal("template-operator"))
				Expect(template.Spec.Version).To(Equal(moduleVersion))
				Expect(template.Spec.Info.Repository).To(Equal("https://github.com/kyma-project/template-operator"))
				Expect(
					template.Spec.Info.Documentation,
				).To(Equal("https://github.com/kyma-project/template-operator/blob/main/README.md"))

				By("And template should have rawManifest resource")
				Expect(template.Spec.Resources).To(HaveLen(1))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(
					template.Spec.Resources[0].Link,
				).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
					moduleVersion)))
			})

			By("And the component constructor file should contain expected content", func() {
				constructorData, err := os.ReadFile(constructorFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(constructorData)).To(ContainSubstring("components:"))
				Expect(string(constructorData)).To(ContainSubstring("name: kyma-project.io/module/template-operator"))
				Expect(string(constructorData)).To(ContainSubstring("version: " + moduleVersion))
				Expect(string(constructorData)).To(ContainSubstring("resources:"))
				Expect(string(constructorData)).To(ContainSubstring(common.RawManifestResourceName))
				Expect(string(constructorData)).To(ContainSubstring(common.ModuleTemplateResourceName))
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})
})

func readModuleTemplate(filepath string) (*v1beta2.ModuleTemplate, error) {
	moduleTemplate := &v1beta2.ModuleTemplate{}
	moduleFile, err := os.ReadFile(filepath)
	if err != nil && len(moduleFile) > 0 {
		return nil, err
	}
	err = yaml.Unmarshal(moduleFile, moduleTemplate)
	if err != nil {
		return nil, err
	}
	return moduleTemplate, err
}

func getDescriptor(template *v1beta2.ModuleTemplate) *compdesc.ComponentDescriptor {
	ocmDesc, err := compdesc.Decode(
		template.Spec.Descriptor.Raw,
		[]compdesc.DecodeOption{compdesc.DisableValidation(true)}...)
	if err != nil {
		return nil
	}

	return ocmDesc
}

func flatten(labels ocmv1.Labels) map[string]string {
	labelsMap := make(map[string]string)
	for _, l := range labels {
		var value string
		_ = yaml.Unmarshal(l.Value, &value)
		labelsMap[l.Name] = value
	}
	return labelsMap
}

func filesIn(dir string) []string {
	fi, err := os.Stat(dir)
	Expect(err).ToNot(HaveOccurred())
	Expect(fi.IsDir()).To(BeTrueBecause("The provided path should be a directory: %s", dir))

	dirFs := os.DirFS(dir)
	entries, err := fs.ReadDir(dirFs, ".")
	Expect(err).ToNot(HaveOccurred())

	var res []string
	for _, ent := range entries {
		if ent.Type().IsRegular() {
			res = append(res, ent.Name())
		}
	}

	return res
}

func validateMinimalModuleTemplate(template *v1beta2.ModuleTemplate, descriptor *compdesc.ComponentDescriptor) {
	Expect(descriptor).ToNot(BeNil())
	Expect(descriptor.SchemaVersion()).To(Equal(v2.SchemaVersion))
	Expect(template.Name).To(Equal(fmt.Sprintf("template-operator-%s", moduleVersion)))

	By("And spec.info should be correct")
	Expect(template.Spec.ModuleName).To(Equal("template-operator"))
	Expect(template.Spec.Version).To(Equal(moduleVersion))
	Expect(template.Spec.Info.Repository).To(Equal("https://github.com/kyma-project/template-operator"))
	Expect(
		template.Spec.Info.Documentation,
	).To(Equal("https://github.com/kyma-project/template-operator/blob/main/README.md"))
	Expect(template.Spec.Info.Icons).To(HaveLen(1))
	Expect(template.Spec.Info.Icons[0].Name).To(Equal("module-icon"))
	Expect(
		template.Spec.Info.Icons[0].Link,
	).To(Equal("https://github.com/kyma-project/template-operator/blob/main/docs/assets/logo.png"))

	By("And annotations should be correct")
	annotations := template.Annotations
	Expect(annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))

	By("And descriptor.component.sources should contain repository entry")
	Expect(len(descriptor.Sources)).To(Equal(1))
	source := descriptor.Sources[0]
	sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
	Expect(err).ToNot(HaveOccurred())
	githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
	Expect(ok).To(BeTrue())
	Expect(github.Type).To(Equal(githubAccessSpec.Type))
	Expect(githubAccessSpec.RepoURL).To(Equal("https://github.com/kyma-project/template-operator"))

	By("And spec.associatedResources should be empty")
	Expect(template.Spec.AssociatedResources).To(BeEmpty())

	By("And spec.manager should be nil")
	Expect(template.Spec.Manager).To(BeNil())

	By("And spec.resources should contain rawManifest")
	Expect(template.Spec.Resources).To(HaveLen(1))
	Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
	Expect(
		template.Spec.Resources[0].Link,
	).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
		moduleVersion)))

	By("And spec.requiresDowntime should be set to false")
	Expect(template.Spec.RequiresDowntime).To(BeFalse())

	By("And module template should not have operator.kyma-project.io/internal label")
	val, ok = template.Labels[shared.InternalLabel]
	Expect(val).To(BeEmpty())
	Expect(ok).To(BeFalse())

	By("And module template should not have operator.kyma-project.io/beta label")
	val, ok = template.Labels[shared.BetaLabel]
	Expect(val).To(BeEmpty())
	Expect(ok).To(BeFalse())
}

func validateTemplateWithFileReference(template *v1beta2.ModuleTemplate, descriptor *compdesc.ComponentDescriptor,
	version string,
) {
	Expect(descriptor).ToNot(BeNil())
	Expect(descriptor.SchemaVersion()).To(Equal(v2.SchemaVersion))

	Expect(template).ToNot(BeNil())
	Expect(template.Name).To(Equal("template-operator-" + version))
	Expect(template.Spec.ModuleName).To(Equal("template-operator"))
	Expect(template.Spec.Version).To(Equal(version))

	By("And descriptor.component.resources should be correct")
	Expect(len(descriptor.Resources)).To(BeNumerically(">=", 2))

	resource := findResourceByName(descriptor.Resources, "raw-manifest")
	Expect(resource).ToNot(BeNil(), "raw-manifest resource not found")
	Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
	Expect(resource.Type).To(Equal("directoryTree"))
	Expect(resource.Version).To(Equal(version))

	manifestResourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
	Expect(err).ToNot(HaveOccurred())
	manifestAccessSpec, ok := manifestResourceAccessSpec.(*localblob.AccessSpec)
	Expect(ok).To(BeTrue())
	Expect(manifestAccessSpec.GetType()).To(Equal(localblob.Type))
	Expect(manifestAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
	Expect(manifestAccessSpec.MediaType).To(Equal("application/x-tar"))
	Expect(manifestAccessSpec.ReferenceName).To(Equal("raw-manifest"))

	resource = findResourceByName(descriptor.Resources, "default-cr")
	Expect(resource).ToNot(BeNil(), "default-cr resource not found")
	Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
	Expect(resource.Type).To(Equal("directoryTree"))
	Expect(resource.Version).To(Equal(version))

	defaultCRResourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
	Expect(err).ToNot(HaveOccurred())
	defaultCRAccessSpec, ok := defaultCRResourceAccessSpec.(*localblob.AccessSpec)
	Expect(ok).To(BeTrue())
	Expect(defaultCRAccessSpec.GetType()).To(Equal(localblob.Type))
	Expect(defaultCRAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
	Expect(defaultCRAccessSpec.MediaType).To(Equal("application/x-tar"))
	Expect(defaultCRAccessSpec.ReferenceName).To(Equal("default-cr"))
}

func getImageResources(descriptor *compdesc.ComponentDescriptor) []compdesc.Resource {
	var imageResources []compdesc.Resource
	for _, resource := range descriptor.Resources {
		if resource.Type == "ociArtifact" {
			imageResources = append(imageResources, resource)
		}
	}
	return imageResources
}

func getImageResourcesMap(descriptor *compdesc.ComponentDescriptor) map[string]compdesc.Resource {
	resourceMap := make(map[string]compdesc.Resource)
	for _, resource := range descriptor.Resources {
		if resource.Type == "ociArtifact" {
			resourceMap[resource.Name] = resource
		}
	}
	return resourceMap
}

func extractImageNamesFromResources(resources []compdesc.Resource) []string {
	var names []string
	for _, resource := range resources {
		names = append(names, resource.Name)
	}
	return names
}

func extractImageURLsFromResources(resources []compdesc.Resource) []string {
	var urls []string
	for _, resource := range resources {
		accessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
		if err != nil {
			continue
		}
		if ociSpec, ok := accessSpec.(*ociartifact.AccessSpec); ok {
			urls = append(urls, ociSpec.ImageReference)
		}
	}
	return urls
}

func getImageReference(resource compdesc.Resource) (string, error) {
	accessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
	if err != nil {
		return "", err
	}
	ociSpec, ok := accessSpec.(*ociartifact.AccessSpec)
	if !ok {
		return "", fmt.Errorf("resource is not an OCI artifact")
	}
	return ociSpec.ImageReference, nil
}

func verifyImageResource(resources map[string]compdesc.Resource,
	imageName, expectedReference, expectedVersion string,
) error {
	resource, exists := resources[imageName]
	if !exists {
		return fmt.Errorf("expected image '%s' not found in resources", imageName)
	}

	actualReference, err := getImageReference(resource)
	if err != nil {
		return fmt.Errorf("failed to get image reference for '%s': %w", imageName, err)
	}

	if actualReference != expectedReference {
		return fmt.Errorf("image '%s' has reference '%s', expected '%s'", imageName, actualReference, expectedReference)
	}

	if resource.Version != expectedVersion {
		return fmt.Errorf("image '%s' has version '%s', expected '%s'", imageName, resource.Version, expectedVersion)
	}

	return nil
}

func findResourceByName(resources []compdesc.Resource, name string) *compdesc.Resource {
	for i := range resources {
		if resources[i].Name == name {
			return &resources[i]
		}
	}
	return nil
}

func findResourceByNameVersionType(resources []compdesc.Resource, name, version, typ string) *compdesc.Resource {
	for i := range resources {
		r := &resources[i]
		if r.Name == name && r.Version == version && r.Type == typ {
			return r
		}
	}
	return nil
}
