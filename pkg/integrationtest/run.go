/*
 *  Copyright (C) 2011-2021 Red Hat, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/commonjava/indy-tests/pkg/buildtest"
	"github.com/commonjava/indy-tests/pkg/common"
	"github.com/commonjava/indy-tests/pkg/dataset"
	"github.com/commonjava/indy-tests/pkg/datest"
	"github.com/commonjava/indy-tests/pkg/promotetest"
)

const (
	DEFAULT_ROUTINES     = 4
	TMP_METADATA_DIR     = "/tmp/metadata"
	PROMOTE_TARGET_STORE = "pnc-builds"
)

/*
 * Run integration test. If dryRun is true, prints the repo creation, down/upload, promote, and clean up info without really doing them.
 * If not dryRun, it will (in order):
 *
 * a. Load the corresponding dataset
 * b. Retrieve the metadata files in da.json to make sure they can be downloaded successfully
 * c. Create a temp group which is same to PNC temp build group (with a hosted repo, namely A)
 * d. Download the files in tracking "downloads" section from this temp group
 * e. Download the files in tracking "uploads" section, and rename all the files (jar, pom, and so on)
 *    with a new version suffix and upload them again to the hosted repo A
 * f. Retrieve the metadata files that will be affected by promotion (we need to foresee those metadata files)
 * g. Promote the files in hosted repo A to hosted repo pnc-builds.
 * h. Retrieve the metadata files from step #f again, check if the new version is available
 * i. Clean the test files by rollback the promotion.
 * j. Retrieve the metadata files from step #f again, check if the new version is gone
 * k. Delete the temp group and the hosted repo A. Delete folo record.
 */
func Run(indyBaseUrl, datasetRepoUrl, buildId, promoteTargetStore, metaCheckRepo string, clearCache, dryRun, keepPod bool) {
	//a. Clone dataset repo
	datasetRepoDir := cloneRepo(datasetRepoUrl)
	fmt.Printf("Clone SUCCESS, dir: %s\n", datasetRepoDir)

	//Load the info.json
	var info dataset.Info
	infoFileLoc := path.Join(datasetRepoDir, buildId, dataset.INFO_JSON)
	json.Unmarshal(common.ReadByteFromFile(infoFileLoc), &info)

	start := time.Now()

	//b. Retrieve the metadata files in da.json
	retrieveAlignmentMetadata(indyBaseUrl, datasetRepoDir, buildId, info)
	t := time.Now()
	fmt.Printf("Retrieve metadata SUCCESS, elapsed(s): %f\n", t.Sub(start).Seconds())

	//c/d/e. Create a mock build group, download files, rename to-be-uploaded files
	packageType := getPackageType(info)
	foloFileLoc := path.Join(datasetRepoDir, buildId, dataset.TRACKING_JSON)
	foloTrackContent := common.GetFoloRecordFromFile(foloFileLoc)
	originalIndy := getOriginalIndyBaseUrl(foloTrackContent.Uploads[0].LocalUrl)

	prev := t
	buildName := buildtest.DoRun(originalIndy, "", indyBaseUrl, packageType, foloTrackContent, 1, clearCache, dryRun)
	t = time.Now()
	fmt.Printf("Create mock group(%s) and download/upload SUCCESS, elapsed(s): %f\n", buildName, t.Sub(prev).Seconds())

	//k. Delete the temp group and the hosted repo, and folo record
	defer cleanUp(indyBaseUrl, packageType, buildName, dryRun)

	//f. Retrieve the metadata files which will be affected by promotion
	metaFiles := calculateMetadataFiles(foloTrackContent)
	metaFilesLoc := path.Join(TMP_METADATA_DIR, "before-promote")
	newVersionNum := buildName[len(common.BUILD_TEST_):]
	exists := true
	passed, e := retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo, metaFiles, metaFilesLoc, newVersionNum, !exists)
	if !passed {
		logger.Infof("Metadata check failed (before). Errors: %s", e.Error())
		return
	}
	fmt.Printf("Metadata validate (before) SUCCESS\n")

	//g. Promote the files in hosted repo A to hosted repo pnc-builds
	foloTrackId := buildName
	sourceStore, targetStore := getPromotionSrcTargetStores(packageType, buildName, promoteTargetStore, foloTrackContent)
	resp, _, success := promotetest.DoRun(indyBaseUrl, foloTrackId, sourceStore, targetStore, newVersionNum, foloTrackContent, dryRun)
	if !success {
		fmt.Printf("Promote failed, %s\n", resp)
		return
	}

	//h. Retrieve the metadata files again, check the new version
	fmt.Printf("Waiting 30s...\n")
	time.Sleep(30 * time.Second) // wait for Indy event handled

	metaFilesLoc = path.Join(TMP_METADATA_DIR, "after-promote")
	passed, e = retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo, metaFiles, metaFilesLoc, newVersionNum, exists)
	if !passed {
		logger.Infof("Metadata check failed (after promotion). Errors: %s", e.Error())
		return
	}
	fmt.Printf("Metadata validate (after promotion) SUCCESS\n")

	//i. Rollback the promotion
	promotetest.Rollback(indyBaseUrl, resp, dryRun)

	//h. Retrieve the metadata files again, check the new version is GONE
	fmt.Printf("Waiting 30s...\n")
	time.Sleep(30 * time.Second)

	metaFilesLoc = path.Join(TMP_METADATA_DIR, "rollback")
	passed, e = retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo, metaFiles, metaFilesLoc, newVersionNum, !exists)
	if !passed {
		logger.Infof("Metadata check failed (rollback). Errors: %s", e.Error())
		return
	}
	fmt.Printf("Metadata validate (rollback) SUCCESS\n")

	// Pause and keep pod for debugging
	if keepPod {
		time.Sleep(30 * time.Minute)
	}
}

func getPromotionSrcTargetStores(packageType, buildName, targetStoreName string, foloTrackContent common.TrackedContent) (string, string) {
	toks := strings.Split(foloTrackContent.Uploads[0].StoreKey, ":")
	sourceStore := fmt.Sprintf("%s:%s:%s", toks[0], toks[1], buildName)
	if targetStoreName == "" {
		targetStoreName = PROMOTE_TARGET_STORE
	}
	targetStore := packageType + ":hosted:" + targetStoreName
	fmt.Printf("Get promotion sourceStore: %s, targetStore: %s\n", sourceStore, targetStore)
	return sourceStore, targetStore
}

func calculateMetadataFiles(foloTrackContent common.TrackedContent) []string {
	paths := []string{}
	for _, up := range foloTrackContent.Uploads {
		if strings.HasSuffix(up.Path, ".pom") {
			versionsDir := path.Dir(up.Path)
			artifactDir := path.Dir(versionsDir)
			metadataPath := path.Join(artifactDir, common.MAVEN_METADATA_XML)
			paths = append(paths, metadataPath)
		}
	}
	return paths
}

func retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo string, metaFiles []string, filesLoc, versionNumber string, exist bool) (bool, error) {
	if metaCheckRepo == "" {
		fmt.Printf("Skip metadata check, no metaCheckRepo specified.\n")
		return true, nil
	}

	repoType := "group"
	repoName := metaCheckRepo

	// Also support the full repo name, e.g, maven:group:test-builds
	index := strings.Index(metaCheckRepo, ":")
	if index > 0 {
		toks := strings.Split(metaCheckRepo, ":")
		repoType = toks[1]
		repoName = toks[2]
	}

	// Download meta files
	for _, p := range metaFiles {
		url := common.GetIndyContentUrl(indyBaseUrl, packageType, repoType, repoName, p)
		common.DownloadFile(url, path.Join(filesLoc, p))
	}

	// Check version
	success := true
	var e common.MultiError
	for _, p := range metaFiles {
		file := path.Join(filesLoc, p)
		// read file and see if version exist
		content := string(common.ReadByteFromFile(file))
		fmt.Printf("Check metadata, file: %s, content:\n%s\n", file, content)
		index := strings.Index(content, common.REDHAT_+versionNumber)
		isExist := false
		if index >= 0 {
			isExist = true
		}
		if isExist != exist {
			success = false
			e.Append(p)
		}
	}
	return success, &e
}

func cloneRepo(datasetRepoUrl string) string {
	return common.DownloadRepo(datasetRepoUrl)
}

func retrieveAlignmentMetadata(indyBaseUrl, datasetRepoDir, buildId string, info dataset.Info) {
	fileLoc := path.Join(datasetRepoDir, buildId, dataset.DA_JSON)

	// Read jsonFile
	byteValue := common.ReadByteFromFile(fileLoc)

	// Parse it
	var arr []string
	json.Unmarshal([]byte(byteValue), &arr)

	var urls []string
	packageType := getPackageType(info)
	groupName := "DA"
	if info.TemporaryBuild {
		groupName = "DA-temporary-builds"
	}

	for _, v := range arr {
		u := common.GetIndyContentUrl(indyBaseUrl, packageType, "group", groupName, v)
		urls = append(urls, u)
	}

	if logger.IsLevelEnabled(logger.DebugLevel) {
		for _, v := range urls {
			fmt.Println(v)
		}
	}

	datest.LookupMetadataByRoutines(urls, DEFAULT_ROUTINES)
}

func getPackageType(info dataset.Info) string {
	packageType := "maven"
	if strings.EqualFold(info.BuildType, "NPM") {
		packageType = "npm"
	}
	return packageType
}

func getOriginalIndyBaseUrl(localUrl string) string {
	u, err := url.Parse(localUrl)
	if err != nil {
		panic(err)
	}
	return u.Scheme + "://" + u.Host
}

func cleanUp(indyBaseUrl, packageType, buildName string, dryRun bool) {
	if dryRun {
		fmt.Printf("Dry run cleanUp\n")
		return
	}

	buildtest.DeleteIndyTestRepos(indyBaseUrl, packageType, buildName)

	if common.DeleteFoloRecord(indyBaseUrl, buildName) {
		fmt.Printf("Delete folo record %s SUCCESS\n", buildName)
	} else {
		fmt.Printf("Delete folo record %s FAILED\n", buildName)
	}
}
