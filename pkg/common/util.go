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

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const (
	BUILD_TEST_           = "build-test-"
	ENVAR_TEST_MOUNT_PATH = "TEST_MOUNT_PATH"
	MAVEN_METADATA_XML    = "maven-metadata.xml"
	REDHAT_               = "redhat-"
)

var (
	versionRegexp = regexp.MustCompile(`redhat-([0-9]+)`)
)

func RePanic(e error) {
	if e != nil {
		panic(e)
	}
}

func AlterUploadPath(rawPath, newReleaseNumber string) string {
	return versionRegexp.ReplaceAllString(rawPath, REDHAT_+newReleaseNumber) // replace with new rel number
}

// generate a random 5 digit number for a build repo like "build-test-9xxxx"
func GenerateRandomBuildName() string {
	rand.Seed(time.Now().UnixNano())
	min := 90000
	max := 99999
	return fmt.Sprintf(BUILD_TEST_+"%v", rand.Intn(max-min)+min)
}

type MultiError struct {
	errors []string
}

func (e *MultiError) Error() string {
	return strings.Join(e.errors, ", ")
}

func (e *MultiError) Append(err string) {
	e.errors = append(e.errors, err)
}

func IsMetadata(path string) bool {
	return strings.Index(path, MAVEN_METADATA_XML) > 0
}
