package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/openshift/ci-tools/pkg/api"
	"sigs.k8s.io/yaml"
)

var (
	openshiftReleaseRepository = "./../../openshift/release"
	release                    = "4.8.0-fc.9"
	branch                     = "master"
)

var whitespaceRegexp = regexp.MustCompile(` +`)

func getGitRepositories() ([]string, error) {
	buf, err := exec.Command("oc", "adm", "release", "info", release, "--commits").Output()
	if err != nil {
		return nil, err
	}
	out := string(buf)

	var repos []string
	added := map[string]bool{}

	const sectionPrefix = "\nImages:\n"
	idx := strings.Index(out, sectionPrefix)
	if idx == -1 {
		return nil, fmt.Errorf("oc adm release output does not have images")
	}
	for i, line := range strings.Split(out[idx+len(sectionPrefix):], "\n") {
		if i == 0 {
			// Skip header
			continue
		}
		if line == "" {
			break
		}
		parts := whitespaceRegexp.Split(line, -1)
		if len(parts) != 4 {
			continue
		}
		repo := parts[2]
		if !added[repo] {
			repos = append(repos, repo)
			added[repo] = true
		}
	}
	return repos, nil
}

type testDefinition interface {
	Check(config *api.ReleaseBuildConfiguration) error
}

type namedTest struct {
	namePattern *regexp.Regexp
}

func newNamedTest(pattern string) testDefinition {
	return &namedTest{
		namePattern: regexp.MustCompile(pattern),
	}
}

func (n *namedTest) Check(config *api.ReleaseBuildConfiguration) error {
	var tests []api.TestStepConfiguration
	for _, t := range config.Tests {
		if n.namePattern.MatchString(t.As) {
			tests = append(tests, t)
		}
	}
	if len(tests) == 0 {
		return fmt.Errorf("no tests found for %s", n.namePattern.String())
	}
	if len(tests) > 1 {
		var names []string
		for _, t := range tests {
			names = append(names, t.As)
		}
		return fmt.Errorf("found %d tests for %s: %v", len(tests), n.namePattern.String(), names)
	}
	return fmt.Errorf("found %s", tests[0].As)
}

func configFile(repo, branch string) string {
	const httpsGitHub = "https://github.com/"
	if strings.HasPrefix(repo, httpsGitHub) {
		ownerRepo := repo[len(httpsGitHub):]
		parts := strings.Split(ownerRepo, "/")
		if len(parts) != 2 {
			return ""
		}

		return fmt.Sprintf(
			"%s/ci-operator/config/%s/%s/%s-%s-%s.yaml",
			openshiftReleaseRepository,
			parts[0], parts[1],
			parts[0], parts[1], branch,
		)
	}
	return ""
}

func main() {
	repos, err := getGitRepositories()
	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range repos {
		f, err := os.Open(configFile(repo, branch))
		if err != nil {
			log.Printf("%s: %s", repo, err)
			continue
		}

		buf, err := io.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}

		config := &api.ReleaseBuildConfiguration{}
		if err := yaml.UnmarshalStrict(buf, config); err != nil {
			log.Fatal(err)
		}

		checks := []testDefinition{
			newNamedTest(`^e2e-(?:aws|gcp|agnostic)$`),
			newNamedTest(`^e2e-(?:[a-z]+-)?serial$`),
			newNamedTest(`^e2e-(?:[a-z]+-)?upgrade$`),
		}
		for _, c := range checks {
			if err := c.Check(config); err != nil {
				log.Printf("%s: %s", repo, err)
			}
		}
	}
}
