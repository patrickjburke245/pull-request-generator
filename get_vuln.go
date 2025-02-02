package main

import (
	"context"
	"github.com/google/go-github/v57/github"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
)

func Ptr[T any](v T) *T {
	return &v
}

func getPythonVuln(githubPersonalAccessToken string) (cve, packageName, version string) {
	cve, packageName, version = getVuln(githubPersonalAccessToken, "pip")
	return
}

func getVuln(githubPersonalAccessToken string, ecosystem string) (cve, packageName, version string) {
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(githubPersonalAccessToken)

	options := github.ListGlobalSecurityAdvisoriesOptions{
		Ecosystem: Ptr(ecosystem),
		Severity:  Ptr("critical"),
	}

	securityService := client.SecurityAdvisories
	vulns, _, err := securityService.ListGlobalSecurityAdvisories(ctx, &options)
	if err != nil {
		log.Fatal(err)
	}

	// Pick one of the first five reports
	j := rand.IntN(20)
	// Some reports don't have CVEID, need to skip if that's the case
	for {
		log.Println("Evaluating " + vulns[j].GetCVEID() + "  " + vulns[j].Vulnerabilities[0].GetVulnerableVersionRange())
		if vulns[j].GetCVEID() == "" {
			j = j + 1
			continue
		}

		// Skip versions with < only
		if r := vulns[j].Vulnerabilities[0].GetVulnerableVersionRange(); strings.Contains(r, "< ") {
			j = j + 1
			continue
		}

		// Want only fixable vulns
		if r := vulns[j].Vulnerabilities[0].GetFirstPatchedVersion(); r == "" {
			j = j + 1
			continue
		}
		break
	}
	cve = vulns[j].GetCVEID()
	packageName = vulns[j].Vulnerabilities[0].GetPackage().GetName()
	version = ""
	r := vulns[j].Vulnerabilities[0].GetVulnerableVersionRange()

	// Extract version from version range
	switch {
	case strings.Contains(r, "<="):
		_, version, _ = strings.Cut(r, "<= ")
	/* revisit
	case strings.Contains(r, "<"):
		_, version, _ = strings.Cut(r, "< ")
		s := strings.Split(version, ".")
		log.Println(s)
		i, _ := strconv.Atoi(s[2])
		s[2] = strconv.Itoa(i - 1)

		version = s[0] + "." + s[1] + "." + s[2]
	*/
	case strings.Contains(r, "="):
		_, version, _ = strings.Cut(r, "= ")
	default:
		log.Fatal("Unexpected version range string " + r)
	}

	return
}

func writePythonVuln(packageName string, version string) {
	requirementsFile := findRequirementsTxt("./terragoat/")
	if requirementsFile == "" {
		return
	}
	b, err := os.ReadFile(requirementsFile) // just pass the file name
	if err != nil {
		log.Fatal(err)
	}

	requirementsContents := string(b)

	if strings.HasPrefix(requirementsContents, packageName+" ==") {

		lines := strings.Split(string(requirementsContents), "\n")

		for i, line := range lines {
			if strings.Contains(line, packageName) {
				lines[i] = packageName + "==" + version
			}
		}
		requirementsContents = strings.Join(lines, "\n")

	} else {
		requirementsContents = requirementsContents + "\n" + packageName + "==" + version
	}

	err = os.WriteFile(requirementsFile, []byte(requirementsContents), 0644)
	if err != nil {
		log.Fatal(err)
	}

}

func findRequirementsTxt(repoRoot string) (requirementsFile string) {
	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if strings.Index(info.Name(), "requirements.txt") != -1 {
			requirementsFile = path
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return
}
