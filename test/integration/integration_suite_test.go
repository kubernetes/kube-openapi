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
	"path"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	headerFilePath             = "../../boilerplate/boilerplate.go.txt"
	testdataDir                = "./testdata"
	testPkgRoot                = "k8s.io/kube-openapi/test/integration/testdata"
	outputPkg                  = testPkgRoot + "/pkg/generated"
	generatedCodeFileName      = "openapi_generated.go"
	goldenCodeFilePath         = "pkg/generated/" + generatedCodeFileName
	generatedSwaggerFileName   = "generated.v2.json"
	goldenSwaggerFileName      = "golden.v2.json"
	generatedReportFileName    = "generated.v2.report"
	goldenReportFileName       = "golden.v2.report"
	generatedOpenAPIv3FileName = "generated.v3.json"
	goldenOpenAPIv3Filename    = "golden.v3.json"

	timeoutSeconds = 20.0
)

var (
	workingDirectory string
	tempDir          string
	terr             error
	openAPIGenPath   string

	inputDirs = []string{
		// `path` vs. `filepath` because packages use '/'
		path.Join(testPkgRoot, "listtype"),
		path.Join(testPkgRoot, "maptype"),
		path.Join(testPkgRoot, "structtype"),
		path.Join(testPkgRoot, "dummytype"),
		path.Join(testPkgRoot, "uniontype"),
		path.Join(testPkgRoot, "enumtype"),
		path.Join(testPkgRoot, "custom"),
		path.Join(testPkgRoot, "valuevalidation"),
		path.Join(testPkgRoot, "defaults"),
	}
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

	// Run the OpenAPI code generator.
	By("running openapi-gen")
	args := append([]string{
		"--output-dir", tempDir,
		"--output-pkg", outputPkg,
		"--output-file", generatedCodeFileName,
		"--report-filename", generatedFile(generatedReportFileName),
		"--go-header-file", headerFilePath,
	}, inputDirs...)
	command := exec.Command(openAPIGenPath, args...)
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
	Describe("Validating generated code", func() {
		It("Generated code should match golden files", func() {
			// Diff the generated code against the golden code. Exit code should be zero.
			command := exec.Command(
				"diff", "-u",
				goldenCodeFilePath,
				generatedFile(generatedCodeFileName),
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
				"diff", "-u",
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
				"diff", "-u",
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
				"diff", "-u",
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
