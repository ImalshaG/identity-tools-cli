package applications

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
)

func getAppFileInfo(filePath string) (string, string, string) {

	filename := filepath.Base(filePath)
	fileExtension := filepath.Ext(filename)
	appName := strings.TrimSuffix(filename, fileExtension)

	return appName, filename, fileExtension
}

func isAppExcluded(appName string) bool {

	// Include only the applications added to INCLUDE_ONLY config
	includeOnlyApps, ok := utils.TOOL_CONFIGS.ApplicationConfigs["INCLUDE_ONLY"].([]interface{})
	if ok {
		for _, app := range includeOnlyApps {
			if app.(string) == appName {
				return false
			}
		}
		log.Println("Application " + appName + " is excluded.")
		return true
	} else {
		// Exclude applications added to EXCLUDE_APPLICATIONS config
		appsToExclude, ok := utils.TOOL_CONFIGS.ApplicationConfigs["EXCLUDE"].([]interface{})
		if ok {
			for _, app := range appsToExclude {
				if app.(string) == appName {
					log.Println("Application " + appName + " is excluded.")
					return true
				}
			}
		}
		return false
	}
}

func getDeployedAppNames() []string {

	apps := getAppList()

	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}

	return appNames
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
