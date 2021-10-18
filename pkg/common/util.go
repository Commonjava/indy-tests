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

package common

import "regexp"

const (
	ENVAR_TEST_MOUNT_PATH = "TEST_MOUNT_PATH"
)

var (
	versionRegexp = regexp.MustCompile(`redhat-([0-9]+)`)
)

func RePanic(e error) {
	if e != nil {
		panic(e)
	}
}

func AlterUploadPath(rawPath, buildNumber string) string {
	return versionRegexp.ReplaceAllString(rawPath, "redhat-"+buildNumber) // replace with same build number
}
