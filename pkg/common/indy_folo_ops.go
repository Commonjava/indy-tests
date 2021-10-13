/*
 *  Copyright (C) 2011-2020 Red Hat, Inc.
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
package common

import (
	"encoding/json"
	"fmt"
	"os"
)

type TrackingKey struct {
	Id string `json:"id"`
}

type TrackedContent struct {
	Uploads     []TrackedContentEntry `json:"uploads"`
	Downloads   []TrackedContentEntry `json:"downloads"`
	TrackingKey TrackingKey           `json:"key"`
}

type TrackedContentEntry struct {
	AccessChannel string  `json:"accessChannel"`
	Path          string  `json:"path"`
	OriginUrl     string  `json:"originUrl"`
	LocalUrl      string  `json:"localUrl"`
	Effect        string  `json:"effect"`
	Md5           string  `json:"md5"`
	Sha256        string  `json:"sha256"`
	Sha1          string  `json:"sha1"`
	Size          int64   `json:"size"`
	Timestamps    []int64 `json:"timestamps"`
	StoreKey      string  `json:"storeKey"`
}

func GetFoloRecord(indyURL, foloRecordId string) TrackedContent {
	URL := fmt.Sprintf("%s/api/folo/admin/%s/record", indyURL, foloRecordId)
	fmt.Printf("Start to get folo tracking record through: %s\n", URL)
	trackContent := &TrackedContent{}
	err := GetRespAsJSONType(URL, trackContent)
	if err != nil {
		fmt.Printf("Error: cannot get folo record %s at indy instance %s, error is: %s\n", foloRecordId, indyURL, err.Error())
		os.Exit(1)
	}
	return *trackContent
}

func GetFoloRecordFromFile(fileLoc string) TrackedContent {
	trackContent := &TrackedContent{}
	b := ReadByteFromFile(fileLoc)
	json.Unmarshal(b, trackContent)
	return *trackContent
}

func SealFoloRecord(indyURL, foloRecordId string) bool {
	URL := fmt.Sprintf("%s/api/folo/admin/%s/record", indyURL, foloRecordId)
	fmt.Printf("Start to seal folo tracking record through: %s\n", URL)
	_, _, result := HTTPRequest(URL, MethodPost, nil, false, nil, nil, "", false)
	return result
}

func DeleteFoloRecord(indyURL, foloRecordId string) bool {
	URL := fmt.Sprintf("%s/api/folo/admin/%s/record", indyURL, foloRecordId)
	fmt.Printf("Start to delete folo tracking record through: %s\n", URL)
	_, _, result := HTTPRequest(URL, MethodDelete, nil, false, nil, nil, "", false)
	return result
}
