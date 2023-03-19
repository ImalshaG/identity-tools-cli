/**
* Copyright (c) 2022, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
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

package applications

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(exportFilePath string, format string) {

	// Export all applications to the Applications folder
	exportFilePath = exportFilePath + "/Applications/"
	os.MkdirAll(exportFilePath, 0700)

	apps := getAppList()
	for _, app := range apps {
		if !isAppExcluded(app.Name) {
			exportApp(app.Id, exportFilePath, format)
		}
	}
}

func exportApp(appId string, outputDirPath string, format string) {

	var fileType = "application/yaml"
	if format == "json" {
		fileType = "application/json"
	} else if format == "xml" {
		fileType = "application/xml"
	}

	var APPURL = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications"
	var err error
	var reqUrl = APPURL + "/" + appId + "/exportFile"

	req, err := http.NewRequest("GET", reqUrl, strings.NewReader(""))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	switch statusCode {
	case 401:
		log.Println("Unauthorized access.\nPlease check your Username and password.")
	case 400:
		log.Println("Provided parameters are not in correct format.")
	case 403:
		log.Println("Forbidden request.")
	case 404:
		log.Println("Service Provider not found for the given ID.")
	case 500:
		log.Println("Internal server error.")
	case 200:
		var attachmentDetail = resp.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(attachmentDetail)
		if err != nil {
			log.Println("Error while parsing the content disposition header")
			panic(err)
		}

		var fileName = params["filename"]

		body1, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		exportedFile := outputDirPath + fileName
		appName, _, _ := getAppFileInfo(exportedFile)

		modifiedFile := utils.AddKeywords(body1, exportedFile, appName)
		ioutil.WriteFile(exportedFile, modifiedFile, 0644)
		log.Println("Successfully created the export file : " + exportedFile)
	}
}
