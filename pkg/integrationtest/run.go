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
	"os"
	"path"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/commonjava/indy-tests/pkg/buildtest"
	"github.com/commonjava/indy-tests/pkg/common"
	"github.com/commonjava/indy-tests/pkg/dataset"
	"github.com/commonjava/indy-tests/pkg/datest"
)

const (
	TARGET_DIR       = "target"
	DEFAULT_ROUTINES = 4
)

/*
 * When we start the integration test, it will (in order):
 *
 * a. Load the corresponding dataset
 * b. Retrieve the metadata files in da.json to make sure they can be downloaded successfully
 * c. Create a temp group which is same to PNC temp build group (with a hosted repo, namely A)
 * d. Download the files in tracking "downloads" section from this temp group
 * e. Download the files in tracking "uploads" section, and rename all the files (jar, pom, and so on)
 *    with a new version suffix and upload them again to the hosted repo A
 * f. Retrieve the metadata files that will be updated by below promotion (we need to foresee those metadata files)
 * g. Promote the files in hosted repo A to hosted repo pnc-builds.
 * h. Retrieve the metadata files from step #f again, check if the new version is available
 * i. Clean the test files by rollback the promotion.
 * j. Retrieve the metadata files from step #f again, check if the new version is gone
 * k. Delete the temp group and the hosted repo A.
 */
func Run(indyBaseUrl, datasetRepoUrl, buildId string) {
	//Create target folder (to store downloaded files), e.g, 'target'
	err := os.MkdirAll(TARGET_DIR, 0755)
	common.RePanic(err)

	//a. Clone dataset repo
	datasetRepoDir := cloneRepo(datasetRepoUrl)
	fmt.Printf("Clone SUCCESS, dir: %s\n", datasetRepoDir)

	//Load the info.json
	var info dataset.Info
	infoFileLoc := path.Join(datasetRepoDir, buildId, dataset.INFO_JSON)
	json.Unmarshal(common.ReadByteFromFile(infoFileLoc), &info)

	start := time.Now()

	//b. Retrieve the metadata files in da.json
	retrieveMetadata(indyBaseUrl, datasetRepoDir, buildId, info)
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("Retrieve metadata SUCCESS, elapsed(s): %f\n", elapsed.Seconds())

	//c/d/e. Create a mock build group, download files, rename to-be-uploaded files
	foloFileLoc := path.Join(datasetRepoDir, buildId, dataset.TRACKING_JSON)
	foloTrackContent := common.GetFoloRecordFromFile(foloFileLoc)
	originalIndy := getOriginalIndyBaseUrl(foloTrackContent.Uploads[0].LocalUrl)
	processNum := 1
	prev := t
	buildName := buildtest.DoRun(originalIndy, "", indyBaseUrl, info.BuildType, foloTrackContent, processNum)
	t = time.Now()
	elapsed = t.Sub(prev)
	fmt.Printf("Create mock group(%s) and download/upload SUCCESS, elapsed(s): %f\n", buildName, elapsed.Seconds())

	funcF()

	funcG()

	funcH()

	funcI()

	funcJ()

	//k. Delete the temp group and the hosted repo
	cleanUp(indyBaseUrl, getPackageType(info), buildName)
}

func cloneRepo(datasetRepoUrl string) string {
	return common.DownloadRepo(datasetRepoUrl)
}

func retrieveMetadata(indyBaseUrl, datasetRepoDir, buildId string, info dataset.Info) {
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

func funcF() {

}

func funcG() {

}

func funcH() {

}

func funcI() {

}

func funcJ() {

}

func cleanUp(indyBaseUrl, packageType, buildName string) {
	buildtest.DeleteIndyTestRepos(indyBaseUrl, packageType, buildName)
}