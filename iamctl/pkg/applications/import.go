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
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func ImportAll(inputDirPath string, format string) {

	var importFilePath = "."
	if inputDirPath != "" {
		importFilePath = inputDirPath
	}
	importFilePath = importFilePath + "/Applications/"

	files, err := ioutil.ReadDir(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var appFilePath string
	for _, file := range files {
		appFilePath = importFilePath + file.Name()
		log.Println("Importing app: " + file.Name())
		importApp(appFilePath, format)
	}
}

func importApp(importFilePath string, format string) {

	var reqUrl = utils.SERVER_CONFIGS.ServerUrl + "/t/" + utils.SERVER_CONFIGS.TenantDomain + "/api/server/v1/applications/import/" + format
	var err error

	fmt.Println(reqUrl)
	fileBytes, err := ioutil.ReadFile(importFilePath)
	if err != nil {
		log.Fatal(err)
	}

	extraParams := map[string]string{
		"file": string(fileBytes),
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range extraParams {
		err := writer.WriteField(key, val)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer writer.Close()

	request, err := http.NewRequest("POST", reqUrl, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+utils.SERVER_CONFIGS.Token)
	defer request.Body.Close()

	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	statusCode := resp.StatusCode
	fmt.Println(statusCode)
	switch statusCode {
	case 401:
		log.Println("Unauthorized access.\nPlease check your Username and password.")
	case 400:
		log.Println("Provided parameters are not in correct format.")
	case 403:
		log.Println("Forbidden request.")
	case 409:
		log.Println("An application with the same name already exists.")
	case 500:
		log.Println("Internal server error.")
	case 201:
		log.Println("Application imported successfully.")
	}
}
