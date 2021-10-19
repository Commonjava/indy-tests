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
	"fmt"
	"os"

	"github.com/commonjava/indy-tests/pkg/integrationtest"
	"github.com/spf13/cobra"
)

func NewIntegrationTestCmd() *cobra.Command {

	exec := &cobra.Command{
		Use:     "integrationtest $indyBaseUrl $datasetRepoUrl $buildId $promoteTargetStore(optional) --dryRun(optional)",
		Short:   "To run integration test",
		Example: "integrationtest http://indy.xyz.com https://gitlab.xyz.com/nos/nos-integrationtest-dataset 2836",
		Run: func(cmd *cobra.Command, args []string) {
			if !validate(args) {
				cmd.Help()
				os.Exit(1)
			}
			dryRun, _ := cmd.Flags().GetBool("dryRun")
			promoteTargetStore := ""
			if len(args) >= 4 {
				promoteTargetStore = args[3]
			}
			integrationtest.Run(args[0], args[1], args[2], promoteTargetStore, dryRun)
		},
	}

	exec.Flags().BoolP("dryRun", "d", false, "Print msg for repo creation, down/upload, promote, and clean up, without really doing it.")

	return exec
}

func validate(args []string) bool {
	if len(args) < 3 {
		fmt.Printf("there are 3 mandatory arguments: indyBaseUrl, datasetRepoUrl, buildId!\n\n")
		return false
	}
	return true
}
