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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ExportAll(outputDirPath string, format string) {

	var exportFilePath = "."
	if outputDirPath != "" {
		exportFilePath = outputDirPath
	}
	exportFilePath = exportFilePath + "/Applications/"
	os.MkdirAll(exportFilePath, 0700)

	apps := getAppList()
	for _, app := range apps {
		log.Println("Exporting app: " + app.Name)
		exportApp(app.Id, exportFilePath, format)
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
		fmt.Println("Writing to file: " + exportedFile)
		modifiedFile := utils.AddKeywords(body1, exportedFile)
		ioutil.WriteFile(exportedFile, modifiedFile, 0644)
		log.Println("Successfully created the export file : " + exportedFile)
	}
}

func getAppList() (spIdList []utils.Application) {

	var APPURL = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications"
	var list utils.List

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, _ := http.NewRequest("GET", APPURL, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	req.Header.Set("accept", "*/*")
	defer req.Body.Close()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer writer.Flush()

	err = json.Unmarshal(body, &list)
	if err != nil {
		log.Fatalln(err)
	}
	resp.Body.Close()

	return list.Applications
}
