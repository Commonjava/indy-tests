/*
 *  Copyright (C) 2021-2023 Red Hat, Inc.
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

 package migrate

 import (
	 "fmt"
	 "os"
 
	 "github.com/commonjava/indy-tests/pkg/integrationtest"
	 "github.com/spf13/cobra"
 )
 
 func NewMigrateCmd() *cobra.Command {
 
	 exec := &cobra.Command{
		 Use:     "migrate $indyBaseUrl $datasetRepoUrl $buildId $migrateTargetIndy --dryRun(optional)",
		 Short:   "To migrate artifact",
		 Example: "integrationtest http://indy.xyz.com https://gitlab.xyz.com/nos/nos-integrationtest-dataset 2836 test-builds",
		 Run: func(cmd *cobra.Command, args []string) {
			 if !validate(args) {
				 cmd.Help()
				 os.Exit(1)
			 }
			 clearCache, _ := cmd.Flags().GetBool("clearCache")
			 dryRun, _ := cmd.Flags().GetBool("dryRun")
			 keepPod, _ := cmd.Flags().GetBool("keepPod")
			 promoteTargetStore := "pnc-builds"
			 integrationtest.Run(args[0], args[1], args[2], promoteTargetStore, "", clearCache, dryRun, keepPod, false, "", args[3])
		 },
	 }
 
	 exec.Flags().BoolP("clearCache", "c", false, "Clear cached built artifact files. This will force download from origin again.")
	 exec.Flags().BoolP("dryRun", "d", false, "Print msg for repo creation, down/upload, promote, and clean up, without really doing it.")
	 exec.Flags().BoolP("keepPod", "k", false, "Keep the pod after test to debug.")
	 return exec
 }
 
 func validate(args []string) bool {
	 if len(args) < 4 {
		 fmt.Printf("There are 4 mandatory arguments: indyBaseUrl, datasetRepoUrl, buildId, migrateTargetIndy!\n")
		 return false
	 }
	 return true
 }
 