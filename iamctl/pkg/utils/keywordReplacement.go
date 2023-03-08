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

package utils

import (
	"html/template"
	"strings"
)

func ReplaceKeywords(fileData []byte, appName string) string {

	// Creating a template with the variable to replace
	tmpl, err := template.New("app").Parse(string(fileData))
	if err != nil {
		panic(err)
	}
	var buf strings.Builder

	appKeywordMap := getAppKeywordMappings(appName)

	// Replacing the values by executing the template
	if err := tmpl.Execute(&buf, appKeywordMap); err != nil {
		panic(err)
	}

	return buf.String()
}

func getAppKeywordMappings(appName string) (keywordMappings map[string]interface{}) {

	keywordMappings = SERVER_CONFIGS.KeywordMappings
	if SERVER_CONFIGS.ApplicationConfigs != nil {
		if appConfigs, ok := SERVER_CONFIGS.ApplicationConfigs[appName]; ok {
			if appKeywordMappings, ok := appConfigs.(map[string]interface{})["KEYWORD_MAPPINGS"]; ok {
				mergedKeywordMap := make(map[string]interface{})
				for key, value := range keywordMappings {
					mergedKeywordMap[key] = value.(string)
				}
				for key, value := range appKeywordMappings.(map[string]interface{}) {
					mergedKeywordMap[key] = value.(string)
				}
				return mergedKeywordMap
			}
		}
	}
	return keywordMappings
}
