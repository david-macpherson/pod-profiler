//go:build never
// +build never

package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	filesystem "github.com/tensorworks/go-build-helpers/pkg/filesystem"
	module "github.com/tensorworks/go-build-helpers/pkg/module"
	process "github.com/tensorworks/go-build-helpers/pkg/process"
	validation "github.com/tensorworks/go-build-helpers/pkg/validation"
)

// Alias validation.ExitIfError() as check()
var check = validation.ExitIfError

// new type for taking in an array of strings as a flag for build tags
type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {

	// Parse our command-line flags
	var buildTags arrayFlags
	flag.Var(&buildTags, "tags", "a list of build tags to consider satisfied during the build. Each build tags requires a new flag prefix. e.g. --tags tag1 --tags tag2")
	doClean := flag.Bool("clean", false, "cleans build outputs")
	doGenerate := flag.Bool("generate", false, "performs code generation")
	doRelease := flag.Bool("release", false, "builds executables for all target platforms")
	doBuildImages := flag.Bool("images", false, "builds container images for executables using ko")
	flag.Parse()

	// Disable CGO
	os.Setenv("CGO_ENABLED", "0")

	// Create a build helper for the Go module
	mod, err := module.ModuleInCwd()
	check(err)

	// Determine if we're cleaning the build outputs
	if *doClean == true {
		check(mod.CleanAll())
		os.Exit(0)
	}

	// Determine if we're performing code generation
	if *doGenerate == true || filesystem.Exists(mod.CodegenToolsDir()) == false {

		// Install the Go tools that we require for code generation
		check(mod.InstallGoTools([]string{}))

		// Perform code generation
		check(mod.Generate())
	}

	// Determine if we're building our executables for just the host platform or for the full matrix of release platforms
	if *doRelease == false {
		buildOptions := module.BuildOptions{Scheme: module.Undecorated, BuildTags: buildTags}
		check(mod.BuildBinariesForHost(module.DefaultBinDir, buildOptions))
	} else {
		buildOptions := module.BuildOptions{Scheme: module.PrefixedDirs, BuildTags: buildTags, AdditionalFlags: []string{}}
		check(mod.BuildBinariesForMatrix(
			module.DefaultBinDir,
			buildOptions,
			module.BuildMatrix{
				Platforms:     []string{"linux"},
				Architectures: []string{"amd64"},
				Ignore:        []string{},
			},
		))
	}

	// Determine if we're building container images for our executables
	if *doBuildImages == true {

		// Verify that we are running under a platform that ko supports
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			log.Fatal("Error: cannot build container images since ko only supports Linux and macOS!")
		}

		// Ensure ko is installed
		check(mod.InstallGoTools([]string{
			"github.com/google/ko@v0.8.3",
		}))

		// Retrieve the list of executables for which we will build container images
		output, err := exec.Command("go", "list", "./cmd/...").CombinedOutput()
		executables := strings.Split(strings.Trim(string(output), "\n"), "\n")
		check(err)

		// Use ko to build container images for each of our executables and store them in the local Docker image cache
		for _, executable := range executables {
			check(process.Run([]string{
				filepath.Join(mod.CodegenToolsDir(), "ko"),
				"publish", "--local",
				executable,
			}, nil, nil))
		}
	}
}
