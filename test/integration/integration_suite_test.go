/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	headerFilePath = "../../boilerplate/boilerplate.go.txt"
	testdataDir    = "./testdata"
	testPkgDir     = "k8s.io/kube-openapi/test/integration/testdata"
	inputDir       = testPkgDir + "/listtype" +
		"," + testPkgDir + "/maptype" +
		"," + testPkgDir + "/structtype" +
		"," + testPkgDir + "/dummytype" +
		"," + testPkgDir + "/uniontype" +
		"," + testPkgDir + "/enumtype" +
		"," + testPkgDir + "/custom" +
		"," + testPkgDir + "/valuevalidation" +
		"," + testPkgDir + "/defaults"
	outputBase                 = "pkg"
	outputPackage              = "generated"
	outputBaseFileName         = "openapi_generated"
	generatedSwaggerFileName   = "generated.v2.json"
	generatedReportFileName    = "generated.v2.report"
	goldenSwaggerFileName      = "golden.v2.json"
	goldenReportFileName       = "golden.v2.report"
	generatedOpenAPIv3FileName = "generated.v3.json"
	goldenOpenAPIv3Filename    = "golden.v3.json"

	timeoutSeconds = 10.0
)

var (
	workingDirectory string
	tempDir          string
	terr             error
	openAPIGenPath   string
)

func generatedFile(filename string) string { return filepath.Join(tempDir, filename) }
func testdataFile(filename string) string  { return filepath.Join(testdataDir, filename) }

func TestGenerators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var _ = BeforeSuite(func() {
	// Explicitly manage working directory
	abs, err := filepath.Abs("")
	Expect(err).ShouldNot(HaveOccurred())
	workingDirectory = abs

	// Create a temporary directory for generated swagger files.
	tempDir, terr = os.MkdirTemp("./", "openapi")
	Expect(terr).ShouldNot(HaveOccurred())

	// Build the OpenAPI code generator.
	By("building openapi-gen")
	binaryPath, berr := gexec.Build("../../cmd/openapi-gen/openapi-gen.go")
	Expect(berr).ShouldNot(HaveOccurred())
	openAPIGenPath = binaryPath

	// Run the OpenAPI code generator, creating OpenAPIDefinition code
	// to be compiled into builder.
	By("processing go idl with openapi-gen")
	gr := generatedFile(generatedReportFileName)
	command := exec.Command(openAPIGenPath,
		"-i", inputDir,
		"-o", outputBase,
		"-p", outputPackage,
		"-O", outputBaseFileName,
		"-r", gr,
		"-h", headerFilePath,
	)
	command.Dir = workingDirectory
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, timeoutSeconds).Should(gexec.Exit(0))

	By("writing swagger v2.0")
	// Create the OpenAPI swagger builder.
	binaryPath, berr = gexec.Build("./builder/main.go")
	Expect(berr).ShouldNot(HaveOccurred())

	// Execute the builder, generating an OpenAPI swagger file with definitions.
	gs := generatedFile(generatedSwaggerFileName)
	By("writing swagger to " + gs)
	command = exec.Command(binaryPath, gs)
	command.Dir = workingDirectory
	session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, timeoutSeconds).Should(gexec.Exit(0))

	By("writing OpenAPI v3.0")
	// Create the OpenAPI swagger builder.
	binaryPath, berr = gexec.Build("./builder3/main.go")
	Expect(berr).ShouldNot(HaveOccurred())

	// Execute the builder, generating an OpenAPI swagger file with definitions.
	gov3 := generatedFile(generatedOpenAPIv3FileName)
	By("writing swagger to " + gov3)
	command = exec.Command(binaryPath, gov3)
	command.Dir = workingDirectory
	session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	os.RemoveAll(tempDir)
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("Open API Definitions Generation", func() {
	Describe("openapi-gen --verify", func() {
		It("Verifies that the existing files are correct", func() {
			command := exec.Command(openAPIGenPath,
				"-i", inputDir,
				"-o", outputBase,
				"-p", outputPackage,
				"-O", outputBaseFileName,
				"-r", testdataFile(goldenReportFileName),
				"-h", headerFilePath,
				"--verify-only",
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})

	Describe("Validating OpenAPI V2 Definition Generation", func() {
		It("Generated OpenAPI swagger definitions should match golden files", func() {
			// Diff the generated swagger against the golden swagger. Exit code should be zero.
			command := exec.Command(
				"diff",
				testdataFile(goldenSwaggerFileName),
				generatedFile(generatedSwaggerFileName),
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})

	Describe("Validating OpenAPI V3 Definition Generation", func() {
		It("Generated OpenAPI swagger definitions should match golden files", func() {
			// Diff the generated swagger against the golden swagger. Exit code should be zero.
			command := exec.Command(
				"diff",
				testdataFile(goldenOpenAPIv3Filename),
				generatedFile(generatedOpenAPIv3FileName),
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})

	Describe("Validating API Rule Violation Reporting", func() {
		It("Generated API rule violations should match golden report files", func() {
			// Diff the generated report against the golden report. Exit code should be zero.
			command := exec.Command(
				"diff",
				testdataFile(goldenReportFileName),
				generatedFile(generatedReportFileName),
			)
			command.Dir = workingDirectory
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, timeoutSeconds).Should(gexec.Exit(0))
		})
	})
})
