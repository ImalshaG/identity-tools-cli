/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
*
* WSO2 LLC. licenses this file to you under the Apache License,
* Version 2.0 (the "License"); you may not use this file except
* in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/cmd"
	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

var serverConfigTemplate = map[string]string{

	utils.SERVER_URL_CONFIG:    "",
	utils.CLIENT_ID_CONFIG:     "",
	utils.CLIENT_SECRET_CONFIG: "",
	utils.TENANT_DOMAIN_CONFIG: "",
	utils.USERNAME_CONFIG:      "",
	utils.PASSWORD_CONFIG:      "",
}

var setupCmd = &cobra.Command{
	Use:   "setupCLI",
	Short: "Setup the CLI tool",
	Long:  `You can setup the config folder structure for the CLI tool`,
	Run: func(cmd *cobra.Command, args []string) {
		baseDirPath, _ := cmd.Flags().GetString("baseDir")

		createConfigFolder(baseDirPath)
	},
}

func init() {

	cmd.RootCmd.AddCommand(setupCmd)
	setupCmd.Flags().StringP("baseDir", "d", "", "Path to the base directory")
}

func createConfigFolder(baseDirPath string) {

	templateEnvName := "env"

	// If baseDirPath is not provided, create the config folder in the current working directory
	var err error
	if baseDirPath == "" {
		baseDirPath, err = os.Getwd()
		if err != nil {
			baseDirPath = "."
		}
	}

	// Create environment specific config folder with the name "env"
	envConfigDir := baseDirPath + "/configs/" + templateEnvName + "/"
	os.MkdirAll(envConfigDir, 0700)

	// Create server config file
	serverConfigs, err := json.Marshal(serverConfigTemplate)
	if err != nil {
		fmt.Println("Error in creating the server config template", err)
	}
	os.WriteFile(envConfigDir+utils.SERVER_CONFIG_FILE, serverConfigs, 0644)

	// Create tool config directory
	file, err := os.OpenFile(envConfigDir+utils.TOOL_CONFIG_FILE, os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error in creating the tool config file", err)
	}
	defer file.Close()
}
