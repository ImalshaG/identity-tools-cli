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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all userstores to the UserStores folder.
	log.Println("Exporting user stores...")
	exportFilePath = filepath.Join(exportFilePath, utils.USERSTORES)

	if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
		os.MkdirAll(exportFilePath, 0700)
	} else {
		if utils.TOOL_CONFIGS.AllowDelete {
			utils.RemoveDeletedLocalResources(exportFilePath, getDeployedUserstoreNames())
		}
	}

	userstores, err := getUserStoreList()
	if err != nil {
		log.Println("Error: when exporting userstores.", err)
	} else {
		if !utils.AreSecretsExcluded(utils.TOOL_CONFIGS.ApplicationConfigs) {
			log.Println("Warn: Secrets exclusion cannot be disabled for userstores. All secrets will be masked.")
		}
		for _, userstore := range userstores {
			if !utils.IsResourceExcluded(userstore.Name, utils.TOOL_CONFIGS.UserStoreConfigs) {
				log.Println("Exporting user store: ", userstore.Name)

				err := exportUserStore(userstore.Id, exportFilePath, format)
				if err != nil {
					log.Printf("Error while exporting user store: %s. %s", userstore.Name, err)
				} else {
					log.Println("User store exported successfully: ", userstore.Name)
				}
			}
		}
	}
}

func exportUserStore(userStoreId string, outputDirPath string, format string) error {

	var fileType string
	// TODO: Extend support for json and xml formats.
	switch format {
	case "json":
		fileType = utils.MEDIA_TYPE_JSON
	case "xml":
		fileType = utils.MEDIA_TYPE_XML
	default:
		fileType = utils.MEDIA_TYPE_YAML
	}

	var reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/userstores/" + userStoreId + "/export"

	var err error
	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("error while creating the request to export user store: %s", err)
	}
	req.Header.Set("Content-Type", utils.MEDIA_TYPE_FORM)
	req.Header.Set("accept", fileType)
	req.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)

	defer req.Body.Close()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error while exporting the user store: %s", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		var attachmentDetail = resp.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(attachmentDetail)
		if err != nil {
			return fmt.Errorf("error while parsing the content disposition header: %s", err)
		}

		fileName := params["filename"]
		exportedFileName := filepath.Join(outputDirPath, fileName)
		fileInfo := utils.GetFileInfo(exportedFileName)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error while reading the response body when exporting userstore: %s. %s", fileName, err)
		}

		// Use the common mask for senstive data.
		modifiedBody := []byte(strings.ReplaceAll(string(body), USERSTORE_SECRET_MASK, utils.SENSITIVE_FIELD_MASK))

		userStoreKeywordMapping := getUserStoreKeywordMapping(fileInfo.ResourceName)
		modifiedFile, err := utils.ProcessExportedContent(exportedFileName, modifiedBody, userStoreKeywordMapping)
		if err != nil {
			return fmt.Errorf("error while processing the exported content: %s", err)
		}

		err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
		if err != nil {
			return fmt.Errorf("error when writing the exported content to file: %w", err)
		}
		return nil
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		return fmt.Errorf("error while exporting the user store: %s", error)
	}
	return fmt.Errorf("unexpected error while exporting the user store with status code: %s", strconv.FormatInt(int64(statusCode), 10))
}
