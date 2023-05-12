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

package userstores

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

func ImportAll(inputDirPath string) {

	log.Println("Importing user stores...")
	importFilePath := filepath.Join(inputDirPath, utils.USERSTORES)

	var files []os.FileInfo
	if _, err := os.Stat(importFilePath); os.IsNotExist(err) {
		log.Println("No user stores to import.")
	} else {
		files, err = ioutil.ReadDir(importFilePath)
		if err != nil {
			log.Println("Error importing user stores: ", err)
		}
		if utils.TOOL_CONFIGS.AllowDelete {
			deployedUserstoreList, err := getUserStoreList()
			if err != nil {
				log.Println("Error retrieving deployed userstores: ", err)
			} else {
				removeDeletedDeployedUserstores(files, deployedUserstoreList)
			}
		}
	}

	for _, file := range files {
		userStoreFilePath := filepath.Join(importFilePath, file.Name())
		userStoreName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		if !utils.IsResourceExcluded(userStoreName, utils.TOOL_CONFIGS.UserStoreConfigs) {
			userStoreId, err := getUserStoreId(userStoreFilePath)
			if err != nil {
				log.Printf("Invalid file configurations for user store: %s. %s", userStoreName, err)
			} else {
				err := importUserStore(userStoreId, userStoreFilePath)
				if err != nil {
					log.Println("Error importing user store: ", err)
				}
			}
		}
	}
}

func importUserStore(userStoreId string, importFilePath string) error {

	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		return fmt.Errorf("error when reading the file for user store: %s", err)
	}

	// Replace keyword placeholders in the local file according to the keyword mappings added in configs.
	fileInfo := utils.GetFileInfo(importFilePath)
	userStoreKeywordMapping := getUserStoreKeywordMapping(fileInfo.ResourceName)
	modifiedFileData := utils.ReplaceKeywords(string(fileBytes), userStoreKeywordMapping)

	err = sendImportRequest(userStoreId, importFilePath, modifiedFileData)
	if err != nil {
		return fmt.Errorf("error when importing user store: %s", err)
	}
	return nil
}

func sendImportRequest(userStoreId string, importFilePath string, fileData string) error {

	fileInfo := utils.GetFileInfo(importFilePath)

	var requestMethod, reqUrl string
	if userStoreId != "" {
		log.Println("Updating user store: " + fileInfo.ResourceName)
		reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/userstores/" + userStoreId + "/import"
		requestMethod = "PUT"
	} else {
		log.Println("Creating new user store: " + fileInfo.ResourceName)
		reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/userstores/import"
		requestMethod = "POST"
	}

	var buf bytes.Buffer
	var err error
	_, err = io.WriteString(&buf, fileData)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	mime.AddExtensionType(".yml", "application/yaml")
	mime.AddExtensionType(".xml", "application/xml")
	mime.AddExtensionType(".json", "application/json")

	mimeType := mime.TypeByExtension(fileInfo.FileExtension)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", fileInfo.FileName)},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	_, err = io.Copy(part, &buf)
	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}

	request, err := http.NewRequest(requestMethod, reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		return fmt.Errorf("error when creating the import request: %s", err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error when sending the import request: %s", err)
	}

	statusCode := resp.StatusCode
	if statusCode == 201 {
		log.Println("User store created successfully.")
		return nil
	} else if statusCode == 200 {
		log.Println("User store updated successfully.")
		return nil
	} else if statusCode == 409 {
		log.Println("A user store with the same name already exists.")
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error response for the import request: %s", error)
	}
	return fmt.Errorf("unexpected error when importing user store: %s", resp.Status)
}

func getUserStoreId(userStoreFilePath string) (string, error) {

	fileContent, err := ioutil.ReadFile(userStoreFilePath)
	if err != nil {
		return "", fmt.Errorf("error when reading the file: %s. %s", userStoreFilePath, err)
	}
	var userStoreConfig UserStoreConfigurations
	err = yaml.Unmarshal(fileContent, &userStoreConfig)
	if err != nil {
		return "", fmt.Errorf("invalid file content at: %s. %s", userStoreFilePath, err)
	}

	existingUserStoreList, err := getUserStoreList()
	if err != nil {
		return "", fmt.Errorf("error when retrieving the deployed userstore list: %s", err)
	}

	for _, userstore := range existingUserStoreList {
		if userstore.Id == userStoreConfig.ID {
			return userstore.Id, nil
		}
	}
	return "", nil
}

func removeDeletedDeployedUserstores(localFiles []os.FileInfo, deployedUserstores []userStore) {

	// Remove deployed user stores that do not exist locally.
deployedResourcess:
	for _, userstore := range deployedUserstores {
		for _, file := range localFiles {
			if userstore.Name == utils.GetFileInfo(file.Name()).ResourceName {
				continue deployedResourcess
			}
		}
		log.Println("User store not found locally. Deleting userstore: ", userstore.Name)
		err := utils.DeleteResource(userstore.Id, "userstores")
		if err != nil {
			log.Println("Error deleting user store: ", err)
		}
	}
}
