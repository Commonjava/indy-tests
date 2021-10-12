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

package dataset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/commonjava/indy-tests/pkg/common"
)

type Info struct {
	PncBaseUrl     string `json:"pncBaseUrl"`
	BuildId        string `json:"buildId"`
	BuildType      string `json:"buildType"`
	TemporaryBuild bool   `json:"temporaryBuild"`
}

const (
	DATASET_DIR   = "dataset"
	DA_JSON       = "da.json"
	TRACKING_JSON = "tracking.json"
	INFO_JSON     = "info.json"
)

/**
 * For each group build or normal build, generate folder structure as below.
 *
 * dataset
 * # for group build
 * |--2836 => group build id
 *     |-- info.json => info about this test dataset, e.g, pnc base url, which is useful to orchestrator
 *     |-- group-build.json => Get by "/pnc-rest/v2/group-builds/2836"
 *     |-- dependency-graph.json => Get by "/pnc-rest/v2/group-builds/2836/dependency-graph"
 *     |-- builds
 *           |-- ...
 *           |-- AMJMVSDA5EAAE
 *                  |-- da.json => metadata list, parsed from alignment log "/pnc-rest/v2/builds/AMJMVSDA5EAAE/logs/align"
 *                  |-- tracking.json => indy tracking record, get by "http://<indy>/api/folo/admin/build-AMJMVSDA5EAAE/report"
 *
 * # for normal build, we download similar files but ignore the dependencies
 * |-- AMJMVSDA5EAAA => build id
 *     |-- info.json => same as above
 *     |-- build.json => Get by "/pnc-rest/v2/builds/AMJMVSDA5EAAA"
 *     |-- da.json => same as above
 *     |-- tracking.json => same as above
 */
func Run(pncBaseUrl, indyBaseUrl, buildId string) {
	//Create folder, e.g, 'dataset/2836'
	dirLoc := path.Join(DATASET_DIR, buildId)
	err := os.MkdirAll(dirLoc, 0755)
	common.RePanic(err)

	//Check if this is a group build
	isGroupBuild := false
	buildURL := pncBaseUrl + "/pnc-rest/v2/group-builds/" + buildId
	if common.HttpExists(buildURL) {
		isGroupBuild = true
	}

	//Download group-build.json or build.json
	var buildJsonFileLoc string
	if isGroupBuild {
		buildJsonFileLoc = path.Join(dirLoc, "group-build.json")
	} else {
		buildURL = pncBaseUrl + "/pnc-rest/v2/builds/" + buildId
		buildJsonFileLoc = path.Join(dirLoc, "build.json")
	}

	if downloadFileIfNotExist(buildURL, buildJsonFileLoc) {
		formatJsonFile(buildJsonFileLoc)
	}

	//Create info.json
	infoFileLoc := path.Join(dirLoc, INFO_JSON)
	if !common.FileOrDirExists(infoFileLoc) {
		success := createInfoFile(pncBaseUrl, buildId, buildJsonFileLoc, infoFileLoc)
		if !success {
			fmt.Println("Create info.json failed.")
			return
		}
	}

	if isGroupBuild {
		//Download dependency-graph.json
		dependencyGraphURL := buildURL + "/dependency-graph"
		dependencyGraphFileLoc := path.Join(dirLoc, "dependency-graph.json")
		if !common.FileOrDirExists(dependencyGraphFileLoc) {
			success := common.DownloadFile(dependencyGraphURL, dependencyGraphFileLoc)
			if !success {
				fmt.Println("Download dependency-graph.json failed.")
				return
			}
			formatJsonFile(dependencyGraphFileLoc)
		}

		//Create 'builds' dir if not exist
		buildsDir := path.Join(dirLoc, "builds")
		os.MkdirAll(buildsDir, 0755)

		//Parse dependency-graph.json to generate data for each bc
		parseDependency(pncBaseUrl, indyBaseUrl, buildsDir, dependencyGraphFileLoc)
	} else {
		generateFile(pncBaseUrl, indyBaseUrl, dirLoc, buildId)
	}
}

func downloadFileIfNotExist(url, fileLoc string) bool {
	if !common.FileOrDirExists(fileLoc) {
		success := common.DownloadFile(url, fileLoc)
		if !success {
			fmt.Printf("Download file %s failed\n", fileLoc)
			return false
		}
	}
	return true
}

//Read a json file, format and override it
func formatJsonFile(fileLoc string) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, common.ReadByteFromFile(fileLoc), "", "  ")
	common.RePanic(err)
	err = ioutil.WriteFile(fileLoc, prettyJSON.Bytes(), 0644)
	common.RePanic(err)
}

func createInfoFile(pncBaseUrl, buildId string, buildJsonLoc, fileLoc string) bool {
	tempB, bType := parseBuildJson(buildJsonLoc)
	info := &Info{PncBaseUrl: pncBaseUrl, BuildId: buildId, TemporaryBuild: tempB, BuildType: bType}
	fmt.Printf("Get %s, %s\n", info.PncBaseUrl, info.BuildId)
	b, _ := json.MarshalIndent(info, "", " ")
	err := ioutil.WriteFile(fileLoc, b, 0644)
	if err != nil {
		fmt.Printf("Warning: cannot create file due to io error! %s\n", err.Error())
		return false
	}
	return true
}

func parseBuildJson(fileLoc string) (bool, string) {
	// Read jsonFile
	byteValue := common.ReadByteFromFile(fileLoc)

	// Parse it
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	// Find temporaryBuild
	var temporaryBuild bool
	v := reflect.ValueOf(result["temporaryBuild"])
	if v.Kind() == reflect.Bool {
		temporaryBuild = v.Bool()
	}

	// Find buildType
	var buildType string
	v = reflect.ValueOf(result["buildConfigRevision"])
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if key.String() == "buildType" {
				buildTypeValue := v.MapIndex(key)
				buildType = buildTypeValue.Elem().String()
				break
			}
		}
	}
	return temporaryBuild, buildType
}

func parseDependency(pncBaseUrl, indyBaseUrl, buildsDir, fileLoc string) {
	// Read jsonFile
	byteValue := common.ReadByteFromFile(fileLoc)

	// Parse it
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	// Iterate through builds and generate files
	vertices := result["vertices"]
	v := reflect.ValueOf(vertices)
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			//val := v.MapIndex(key)
			buildId := key.String()
			buildDir := path.Join(buildsDir, buildId)
			generateFile(pncBaseUrl, indyBaseUrl, buildDir, buildId)
		}
	}
}

func generateFile(pncBaseUrl, indyBaseUrl, buildDir, buildId string) {
	alignLogFile := path.Join(buildDir, "align.log")
	daFile := path.Join(buildDir, DA_JSON)
	trackingFile := path.Join(buildDir, TRACKING_JSON)

	os.MkdirAll(buildDir, 0755)
	if !common.FileOrDirExists(alignLogFile) {
		alignLog := common.GetAlignLog(pncBaseUrl, buildId)
		err := ioutil.WriteFile(alignLogFile, []byte(alignLog), 0644)
		common.RePanic(err)
		paths := getMetadataPaths(alignLog)
		pathsJson, _ := json.MarshalIndent(paths, "", " ")
		err = ioutil.WriteFile(daFile, pathsJson, 0644)
		common.RePanic(err)
	}

	if !common.FileOrDirExists(trackingFile) {
		url := indyBaseUrl + "/api/folo/admin/build-" + buildId + "/report"
		common.DownloadFile(url, trackingFile)
	}
}

func getMetadataPaths(alignLog string) []string {
	// extract the gav list from alignment log
	// (?s) means single-line (hence the s) or DOTALL mode - it takes the whole alignlog as one string
	var re = regexp.MustCompile(`(?s)REST Client returned.*?\}`)
	var paths []string
	for _, match := range re.FindAllString(alignLog, -1) {
		i := strings.Index(match, "{")
		gavs := match[i+1 : len(match)-1]
		gavArray := strings.Split(gavs, ",")
		for _, gav := range gavArray {
			s := strings.Split(gav, ":")
			groupId := strings.Trim(s[0], " ")
			artifactId := s[1]
			//fmt.Println("GroupID: ", groupId, " ArtifactId: ", artifactId)
			groupIdPath := strings.ReplaceAll(groupId, ".", "/")
			p := fmt.Sprintf("%s/%s/maven-metadata.xml", groupIdPath, artifactId)
			paths = append(paths, p)
		}
		fmt.Println("Get metadata paths: ", len(gavArray))
	}
	fmt.Println("Get metadata paths (Total): ", len(paths))
	return paths
}
