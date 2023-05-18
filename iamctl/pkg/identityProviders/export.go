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
	"mime"
	"os"
	"path/filepath"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all identity providers to the IdentityProviders folder.
	log.Println("Exporting identity providers...")
	exportFilePath = filepath.Join(exportFilePath, utils.IDENTITY_PROVIDERS)
	os.MkdirAll(exportFilePath, 0700)

	excludeSecerts := utils.AreSecretsExcluded(utils.TOOL_CONFIGS.IdpConfigs)
	idps, err := getIdpList()
	if err != nil {
		log.Println("Error: when exporting identity providers.", err)
	} else {
		for _, idp := range idps {
			if !utils.IsResourceExcluded(idp.Name, utils.TOOL_CONFIGS.IdpConfigs) {
				log.Println("Exporting identity provider: ", idp.Name)

				err := exportIdp(idp.Id, exportFilePath, format, excludeSecerts)
				if err != nil {
					log.Printf("Error while exporting identity providers: %s. %s", idp.Name, err)
				} else {
					log.Println("Identity provider exported successfully: ", idp.Name)
				}
			}
		}
	}
	if !utils.IsResourceExcluded(utils.RESIDENT_IDP_NAME, utils.TOOL_CONFIGS.IdpConfigs) {
		log.Println("Exporting Resident identity provider")
		err := exportIdp(utils.RESIDENT_IDP_NAME, exportFilePath, format, excludeSecerts)
		if err != nil {
			log.Printf("Error while exporting resident identity provider: %s", err)
		} else {
			log.Println("Resident identity provider exported successfully")
		}
	}
}

func exportIdp(idpId string, outputDirPath string, format string, excludeSecrets bool) error {

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

	resp, err := utils.SendExportRequest(idpId, fileType, utils.IDENTITY_PROVIDERS, excludeSecrets)
	if err != nil {
		return fmt.Errorf("error while exporting the identity provider: %s", err)
	}
	defer resp.Body.Close()

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
		return fmt.Errorf("error while reading the response body when exporting IDP: %s. %s", fileName, err)
	}

	idpKeywordMapping := getIdpKeywordMapping(fileInfo.ResourceName)
	modifiedFile, err := utils.ProcessExportedContent(exportedFileName, body, idpKeywordMapping)
	if err != nil {
		return fmt.Errorf("error while processing the exported content: %s", err)
	}

	err = ioutil.WriteFile(exportedFileName, modifiedFile, 0644)
	if err != nil {
		return fmt.Errorf("error when writing the exported content to file: %w", err)
	}
	return nil
}
