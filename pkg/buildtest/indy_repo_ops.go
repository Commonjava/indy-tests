package buildtest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	common "github.com/commonjava/indy-tests/pkg/common"
)

const (
	DEFAULT_SHARED_GROUP   = "builds-untested+shared-imports+public"
	DEFAULT_MVN_CENTRAL    = "central"
	DEFAULT_NPM_CENTRAL    = "npmjs"
	TYPE_MVN               = "maven"
	TYPE_NPM               = "npm"
	MISSING_CONTENT_PATH   = "org/missing/1.0/missing-1.0.pom"
	ACCESSIBLE_REMOTE_PATH = "org/apache/apache"
	ABC                    = "xyz"
)

type BuildMetadata struct {
	buildType     string
	centralName   string
	sharedGrpName string
}

func decideMeta(buildType string) *BuildMetadata {
	if buildType == TYPE_MVN {
		return &BuildMetadata{
			buildType:     buildType,
			centralName:   DEFAULT_MVN_CENTRAL,
			sharedGrpName: DEFAULT_SHARED_GROUP,
		}
	} else if buildType == TYPE_NPM {
		return &BuildMetadata{
			buildType:     buildType,
			centralName:   DEFAULT_NPM_CENTRAL,
			sharedGrpName: DEFAULT_SHARED_GROUP,
		}
	}
	return nil
}

func prepareIndyRepos(indyURL, buildName string, buildMeta BuildMetadata, additionalRepos []string, dryRun bool) bool {
	if dryRun {
		fmt.Printf("Dry run prepareIndyRepos\n")
		return true
	}

	prepared := prepareIndyHosted(indyURL, buildMeta.buildType, buildName) &&
		prepareIndyRemote(indyURL, buildMeta.buildType, buildName) &&
		prepareIndyGroup(indyURL, buildName, buildMeta, additionalRepos)

	return prepared
}

func prepareIndyRemote(indyURL, buildType, buildName string) bool {
	remoteVars := IndyRemoteVars{
		Name: buildName,
		Type: buildType,
	}

	URL := fmt.Sprintf("%s/api/admin/stores/%s/remote/%s", indyURL, buildType, buildName)

	remote := IndyRemoteTemplate(&remoteVars)
	fmt.Printf("Start creating remote repo %s\n", buildName)
	result := putRequest(URL, strings.NewReader(remote))
	if result {
		fmt.Printf("Remote repo %s created successfully, check %s for details\n", buildName, URL)
	} else {
		fmt.Printf("Remote repo %s creation fail, no following operations\n", buildName)
	}
	verifyRepoContentCleanup(indyURL, buildType, "remote", buildName)
	return result
}

func prepareIndyHosted(indyURL, buildType, buildName string) bool {
	hostedVars := IndyHostedVars{
		Name: buildName,
		Type: buildType,
	}

	URL := fmt.Sprintf("%s/api/admin/stores/%s/hosted/%s", indyURL, buildType, buildName)

	hosted := IndyHostedTemplate(&hostedVars)
	fmt.Printf("Start creating hosted repo %s\n", buildName)
	result := putRequest(URL, strings.NewReader(hosted))
	if result {
		fmt.Printf("Hosted repo %s created successfully, check %s for details\n", buildName, URL)
	} else {
		fmt.Printf("Hosted repo %s creation fail, no following operations\n", buildName)
	}
	verifyRepoContentCleanup(indyURL, buildType, "hosted", buildName)
	return result
}

func prepareIndyGroup(indyURL, buildName string, buildMeta BuildMetadata, additionalRepos []string) bool {
	var constituents []string
	buildType, sharedGrpName := buildMeta.buildType, buildMeta.sharedGrpName
	hostedChild := fmt.Sprintf("%s:hosted:%s", buildType, buildName)
	remoteChild := fmt.Sprintf("%s:remote:%s", buildType, buildName)
	sharedGrpChild := fmt.Sprintf("%s:group:%s", buildType, sharedGrpName)
	constituents = append(constituents, hostedChild, sharedGrpChild)

	URL := fmt.Sprintf("%s/api/admin/stores/%s/group/%s", indyURL, buildType, buildName)

	created := createOrUpdateGroupRepo(URL, buildName, buildMeta, additionalRepos, constituents)
	if created {
		fmt.Printf("Group repo %s created successfully, check %s for details\n", buildName, URL)
	} else {
		fmt.Printf("Group repo %s created failed, no following operations\n", buildName)
	}

	// remote is not added into group, group NFC & merged path are not accessible
	verifyNFC(indyURL, buildType, buildName)
	verifyMergedPaths(indyURL, buildType, buildName)

	// Update group to add remote constituent
	constituents = append(constituents, remoteChild)
	updated := createOrUpdateGroupRepo(URL, buildName, buildMeta, additionalRepos, constituents)
	if updated {
		fmt.Printf("Group repo %s updated successfully, check %s for details\n", buildName, URL)
	} else {
		fmt.Printf("Group repo %s updated failed, no following operations\n", buildName)
	}

	// remote is added into group, group NFC & merged path are accessible
	verifyNFC(indyURL, buildType, buildName)
	verifyMergedPaths(indyURL, buildType, buildName)
	return updated
}

func createOrUpdateGroupRepo(URL, buildName string, buildMeta BuildMetadata, additionalRepos, childRepos []string) bool {
	var constituents []string
	if childRepos != nil {
		constituents = append(constituents, childRepos...)
	}
	if additionalRepos != nil {
		constituents = append(constituents, additionalRepos...)
	}

	groupVars := IndyGroupVars{
		Name:         buildName,
		Type:         buildMeta.buildType,
		Constituents: constituents,
	}
	group := IndyGroupTemplate(&groupVars)

	fmt.Printf("Start creating/updating group repo %s\n", buildName)
	return putRequest(URL, strings.NewReader(group))
}

//Delete group and hosted, remote repo (with content)
func DeleteIndyTestRepos(indyURL, packageType, buildName string) {
	if !delAllowed(buildName) {
		return
	}
	deleteIndyHosted(indyURL, packageType, buildName)
	deleteIndyRemote(indyURL, packageType, buildName)
	deleteIndyGroup(indyURL, packageType, buildName)
}

func delAllowed(buildName string) bool {
	if strings.HasPrefix(buildName, common.BUILD_TEST_) {
		return true
	}
	fmt.Printf("Can not delete repo(s) %s (not test repo)", buildName)
	return false
}

//Delete hosted repo and content
func deleteIndyHosted(indyURL, buildType, repoName string) {
	URL := fmt.Sprintf("%s/api/admin/stores/%s/hosted/%s?deleteContent=true", indyURL, buildType, repoName)
	fmt.Printf("Start deleting hosted repo %s\n", repoName)
	result := delRequest(URL)
	if result {
		fmt.Printf("Hosted repo %s deleted successfully\n", repoName)
	}
	// verify hosted paths and contents cleanup
	verifyRepoContentCleanup(indyURL, buildType, "hosted", repoName)
	// verify paths in affected groups cleanup and group consistency
	verifyRepoAffectedGroupCleanup(indyURL, buildType, "hosted", repoName)
}

func deleteIndyRemote(indyURL, buildType, repoName string) {
	verifyNFC(indyURL, buildType, repoName)

	URL := fmt.Sprintf("%s/api/admin/stores/%s/remote/%s?deleteContent=true", indyURL, buildType, repoName)
	fmt.Printf("Start deleting remote repo %s\n", repoName)
	result := delRequest(URL)
	if result {
		fmt.Printf("Remote repo %s deleted successfully\n", repoName)
	}

	isGrpCached := isNFCCached(indyURL, buildType, "group", repoName)
	if !isGrpCached {
		fmt.Printf("Group repo %s NFC is removed\n", repoName)
	}
	// verify remote paths and contents cleanup
	verifyRepoContentCleanup(indyURL, buildType, "remote", repoName)
	// verify paths in affected groups cleanup and group consistency
	verifyRepoAffectedGroupCleanup(indyURL, buildType, "remote", repoName)
	// verify whether group merged path is accessible after remote is removed
	verifyMergedPaths(indyURL, buildType, repoName)
}

func deleteIndyGroup(indyURL, buildType, repoName string) {
	URL := fmt.Sprintf("%s/api/admin/stores/%s/group/%s", indyURL, buildType, repoName)
	fmt.Printf("Start deleting group repo %s\n", repoName)
	result := delRequest(URL)
	if result {
		fmt.Printf("Group repo %s deleted successfully\n", repoName)
	}
	// verify group paths and contents cleanup
	verifyRepoContentCleanup(indyURL, buildType, "group", repoName)
}

func verifyRepoContentCleanup(indyURL, buildType, storeType, repoName string) {
	URL := fmt.Sprintf("%s/api/admin/stores/%s/%s/%s", indyURL, buildType, storeType, repoName)
	_, code, _ := getRequest(URL)
	switch code {
	case 404:
		fmt.Printf("%s repo %s is removed\n", storeType, repoName)
	case 200:
		fmt.Printf("%s repo %s is created\n", storeType, repoName)
	}

	contentURL := fmt.Sprintf("%s/api/content/%s/%s/%s", indyURL, buildType, storeType, repoName)
	_, contentCode, _ := getRequest(contentURL)
	switch contentCode {
	case 404:
		fmt.Printf("%s repo %s contents are removed\n", storeType, repoName)
	case 200:
		fmt.Printf("%s repo %s contents are created\n", storeType, repoName)
	}
}

func verifyRepoAffectedGroupCleanup(indyURL, buildType, storeType, repoName string) {
	groupURL := fmt.Sprintf("%s/api/admin/stores/%s/group/%s", indyURL, buildType, repoName)
	groupBody, groupCode, _ := getRequest(groupURL)
	if groupCode != 200 {
		return
	}
	fmt.Printf("Group repo %s is still existed\n", repoName)
	var group map[string]interface{}
	err := json.Unmarshal([]byte(groupBody), &group)
	if err != nil {
		fmt.Printf("Error: Group %s parse failed! Error message: %s\n", repoName, err.Error())
		os.Exit(1)
	}
	constituents := []string{fmt.Sprint(group["constituents"])}
	repoKey := strings.Join([]string{buildType, storeType, repoName}, ":")
	if !common.Contains(constituents, repoKey) {
		fmt.Printf("%s repo %s is removed from Group repo %s\n", storeType, repoName, repoName)
	}

	grpContentURL := fmt.Sprintf("%s/api/content/%s/group/%s", indyURL, buildType, repoName)
	_, grpContentCode, _ := getRequest(grpContentURL)
	if grpContentCode == 200 {
		fmt.Printf("Group repo %s contents are still existed\n", repoName)
	}
}

func verifyMergedPaths(indyURL, buildType, repoName string) {
	remoteAccessibleURL := fmt.Sprintf("%s/api/content/%s/remote/%s/%s", indyURL, buildType, repoName, ACCESSIBLE_REMOTE_PATH)
	_, remoteCode, _ := getRequest(remoteAccessibleURL)
	if remoteCode == 200 {
		fmt.Printf("Remote repo %s path '%s' is accessible\n", repoName, ACCESSIBLE_REMOTE_PATH)
	} else {
		fmt.Printf("Remote repo %s path '%s' is not accessible\n", repoName, ACCESSIBLE_REMOTE_PATH)
	}

	grpAccessibleURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, buildType, repoName, ACCESSIBLE_REMOTE_PATH)
	_, grpCode, _ := getRequest(grpAccessibleURL)
	if grpCode == 200 {
		fmt.Printf("Group repo %s merged path '%s' is accessible\n", repoName, ACCESSIBLE_REMOTE_PATH)
	} else {
		fmt.Printf("Group repo %s merged path '%s' is not accessible\n", repoName, ACCESSIBLE_REMOTE_PATH)
	}
}

func verifyNFC(indyURL, buildType, repoName string) {
	remoteMissingURL := fmt.Sprintf("%s/api/content/%s/remote/%s/%s", indyURL, buildType, repoName, MISSING_CONTENT_PATH)
	_, missingCode, _ := getRequest(remoteMissingURL)
	if missingCode != 404 {
		return
	}
	isCached := isNFCCached(indyURL, buildType, "remote", repoName)
	if isCached {
		fmt.Printf("Remote repo %s NFC caches successfully\n", repoName)
	}

	isGrpCached := isNFCCached(indyURL, buildType, "group", repoName)
	if isGrpCached {
		fmt.Printf("Group repo %s NFC caches successfully\n", repoName)
	} else {
		fmt.Printf("Group repo %s NFC doesn't cache\n", repoName)
	}
}

func isNFCCached(indyURL, buildType, storeType, repoName string) bool {
	nfcURL := fmt.Sprintf("%s/api/nfc/%s/%s/%s", indyURL, buildType, storeType, repoName)
	nfcContent, nfcCode, _ := getRequest(nfcURL)
	if nfcCode != 200 {
		return false
	}
	index := strings.Index(nfcContent, MISSING_CONTENT_PATH)
	if index >= 0 {
		return true
	}
	return false
}

func getRequest(url string) (string, int, bool) {
	content, code, succeeded := common.HTTPRequest(url, common.MethodGet, nil, true, nil, nil, "", false)
	return content, code, succeeded
}

func postRequest(url string, data io.Reader) (string, bool) {
	content, _, succeeded := common.HTTPRequest(url, common.MethodPost, nil, true, data, nil, "", false)
	return content, succeeded
}

func putRequest(url string, data io.Reader) bool {
	_, _, succeeded := common.HTTPRequest(url, common.MethodPut, nil, false, data, nil, "", false)
	return succeeded
}

func delRequest(url string) bool {
	_, _, succeeded := common.HTTPRequest(url, common.MethodDelete, nil, false, nil, nil, "", false)
	return succeeded
}
