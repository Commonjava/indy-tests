package common

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

const (
	TRACKING_SUFFIX = "+tracking"
)

func ValidateTargetIndyOrExit(targetIndy string) (string, bool) {
	targetIndyHost, validated := ValidateTargetIndy(targetIndy)
	if !validated {
		os.Exit(1)
	}
	return targetIndyHost, validated
}

func ValidateTargetIndy(targetIndy string) (string, bool) {
	indyHost := targetIndy
	if strings.HasPrefix(targetIndy, "http") {
		host := GetHost(targetIndy)
		port := GetPort(targetIndy)
		if IsEmptyString(port) || port == "80" {
			indyHost = host
		} else {
			indyHost = fmt.Sprintf("%s:%s", host, port)
		}
	} else {
		indyHost = strings.Split(targetIndy, "/")[0]
	}

	indyTest := ""
	var indyAPIBase string
	if strings.HasPrefix(indyHost, "http") {
		indyAPIBase = indyHost + "/api"
	} else {
		indyAPIBase = "http://" + indyHost + "/api"
	}

	fmt.Printf("Start testing target indy server %s\n", indyHost)
	_, err := url.ParseRequestURI(indyAPIBase)
	if err == nil {
		testPath := "/admin/stores/maven/remote/central"
		indyTest = indyAPIBase + testPath
		_, err = url.ParseRequestURI(indyTest)
		if err != nil {
			fmt.Printf("Error: not a valid indy server: %s because %s does not exist\n", targetIndy, testPath)
			return "", false
		}
	} else {
		fmt.Printf("Error: not a valid indy server: %s\n", targetIndy)
		return "", false
	}
	resp, err2 := http.Get(indyTest)
	if err2 != nil {
		fmt.Printf("Error: %s is not a valid indy server. Cause: %s\n", targetIndy, err2)
		return "", false
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		fmt.Printf("Error: %s returned bad status. Cause: %s\n", targetIndy, resp.Status)
		return "", false
	}
	resp.Body.Close()
	return indyHost, true
}

func StoreKeyToPath(storeKey string) string {
	return strings.ReplaceAll(storeKey, ":", "/")
}

func GetIndyContentUrl(indyBaseUrl, packageType, repoType, repoName, aPath string) string {
	return indyBaseUrl + path.Join("/api/content", packageType, repoType, repoName, aPath)
}
