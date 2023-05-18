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

package identityproviders

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing identity providers...")
	importFilePath := filepath.Join(inputDirPath, utils.IDENTITY_PROVIDERS)

	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No identity providers to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing identity providers: ", err)
		}
	}

	for _, file := range files {
		idpFilePath := filepath.Join(importFilePath, file.Name())
		idpName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(idpName, utils.TOOL_CONFIGS.IdpConfigs) {
			var idpId string
			var err error
			if idpName == utils.RESIDENT_IDP_NAME {
				idpId = utils.RESIDENT_IDP_NAME
			} else {
				idpId, err = getIdpId(idpFilePath, idpName)
			}

			if err != nil {
				log.Printf("Invalid file configurations for identity provider: %s. %s", idpName, err)
			} else {
				err := importIdp(idpId, idpFilePath)
				if err != nil {
					log.Println("Error importing identity provider: ", err)
				}
			}
		}
	}
}

func importIdp(idpId string, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for identity provider: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	idpKeywordMapping := getIdpKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), idpKeywordMapping)

	if idpId == "" {
		log.Println("Creating new identity provider: " + fileInfo.ResourceName)
		err = utils.SendImportRequest(importFilePath, modifiedFileData, utils.IDENTITY_PROVIDERS)
	} else {
		log.Println("Updating identity provider: " + fileInfo.ResourceName)
		err = utils.SendUpdateRequest(idpId, importFilePath, modifiedFileData, utils.IDENTITY_PROVIDERS)
	}
	if err != nil {
		return fmt.Errorf("error when importing identity provider: %s", err)
	}
	log.Println("Identity provider imported successfully.")
	return nil
}

func getIdpId(idpFilePath string, idpName string) (string, error) {

	fileContent, err := ioutil.ReadFile(idpFilePath)
	if err != nil {
		return "", fmt.Errorf("error when reading the file for idp: %s. %s", idpName, err)
	}
	var idpConfig idpConfig
	err = yaml.Unmarshal(fileContent, &idpConfig)
	if err != nil {
		return "", fmt.Errorf("invalid file content for idp: %s. %s", idpName, err)
	}
	existingIdpList, err := getIdpList()
	if err != nil {
		return "", fmt.Errorf("error when retrieving the deployed idp list: %s", err)
	}

	for _, idp := range existingIdpList {
		if idp.Name == idpConfig.IdentityProviderName {
			return idp.Id, nil
		}
	}
	return "", nil
}
