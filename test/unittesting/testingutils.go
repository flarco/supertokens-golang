/*
 * Copyright (c) 2021, VRAI Labs and/or its affiliates. All rights reserved.
 *
 * This software is licensed under the Apache License, Version 2.0 (the
 * "License") as published by the Apache Software Foundation.
 *
 * You may not use this file except in compliance with the License. You may
 * obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package unittesting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/supertokens/supertokens-golang/recipe/thirdparty/tpmodels"
	"gopkg.in/h2non/gock.v1"
)

func getListOfPids() []string {
	installationPath := getInstallationDir()
	pathOfDirToRead := installationPath + "/.started/"
	files, err := ioutil.ReadDir(pathOfDirToRead)
	if err != nil {
		return []string{}
	}
	var result []string
	for _, file := range files {
		pathOfFileToBeRead := installationPath + "/.started/" + file.Name()
		data, err := ioutil.ReadFile(pathOfFileToBeRead)
		if err != nil {
			log.Fatalf(err.Error(), "THIS IS GET-LIST-OF-PIDS")
		}
		if string(data) != "" {
			result = append(result, string(data))
		}
	}
	return result
}

func SetUpST() {
	installationPath := getInstallationDir()
	cmd := exec.Command("cp", "temp/config.yaml", "./config.yaml")
	cmd.Dir = installationPath
	err := cmd.Run()
	if err != nil {
		log.Fatalf(err.Error(), "THIS IS SETUP-ST")
	}
}

func StartUpST(host string, port string) string {
	pidsBefore := getListOfPids()
	command := fmt.Sprintf(`java -Djava.security.egd=file:/dev/urandom -classpath "./core/*:./plugin-interface/*" io.supertokens.Main ./ DEV host=%s port=%s test_mode`, host, port)
	startTime := getCurrTimeInMS()
	shellout(command)
	for getCurrTimeInMS()-startTime < 30000 {
		pidsAfter := getListOfPids()
		if len(pidsAfter) <= len(pidsBefore) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		nonIntersection := getNonIntersection(pidsAfter, pidsBefore)
		if len(nonIntersection) < 1 {
			panic("something went wrong while starting ST")
		} else {
			return nonIntersection[0]
		}
	}
	panic("could not start ST process")
}

func getNonIntersection(a1 []string, a2 []string) []string {
	var result = []string{}
	for i := 0; i < len(a1); i++ {
		there := false
		for y := 0; y < len(a2); y++ {
			if a1[i] == a2[y] {
				there = true
			}
		}
		if !there {
			result = append(result, a1[i])
		}
	}
	return result
}

func getCurrTimeInMS() uint64 {
	return uint64(time.Now().UnixNano() / 1000000)
}

//helper function to execute shell commands
func shellout(command string) {
	installationPath := getInstallationDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = installationPath
	err := cmd.Start()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func stopST(pid string) {
	installationPath := getInstallationDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	pidsBefore := getListOfPids()
	if len(pidsBefore) == 0 {
		return
	}
	if len(pidsBefore) == 1 {
		if pidsBefore[0] == "" {
			return
		}
	}
	cmd := exec.Command("kill", pid)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = installationPath
	err := cmd.Run()

	if err != nil {
		log.Fatal(err.Error(), "error in killSt")
	}

	startTime := getCurrTimeInMS()
	for getCurrTimeInMS()-startTime < 10000 {
		pidsAfter := getListOfPids()
		if itemExists(pidsAfter, pid) {
			time.Sleep(100 * time.Millisecond)
		} else {
			return
		}
	}
	panic("Could not stop ST")
}

func itemExists(arr []string, item string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func CleanST() {
	installationPath := getInstallationDir()

	cmd := exec.Command("rm", "config.yaml")
	cmd.Dir = installationPath

	err := cmd.Run()

	if err != nil {
		log.Fatalf(err.Error(), "could not delete the config yaml file [THIS IS CLEAN-ST]")
	}

	cmd = exec.Command("rm", "-rf", ".webserver-temp-*")
	cmd.Dir = installationPath
	err = cmd.Run()

	if err != nil {
		log.Fatalf(err.Error(), "could not delete the webserver-temp files [THIS IS CLEAN-ST]")
	}

	cmd = exec.Command("rm", "-rf", ".started")
	cmd.Dir = installationPath
	err = cmd.Run()

	if err != nil {
		log.Fatalf(err.Error(), "could not delete the .started file [THIS IS CLEAN-ST]")
	}

}

// MaxVersion returns max of v1 and v2
func MaxVersion(version1 string, version2 string) string {
	var splittedv1 = strings.Split(version1, ".")
	var splittedv2 = strings.Split(version2, ".")
	var minLength = len(splittedv1)
	if minLength > len(splittedv2) {
		minLength = len(splittedv2)
	}
	for i := 0; i < minLength; i++ {
		var v1, _ = strconv.Atoi(splittedv1[i])
		var v2, _ = strconv.Atoi(splittedv2[i])
		if v1 > v2 {
			return version1
		} else if v2 > v1 {
			return version2
		}
	}
	if len(splittedv1) >= len(splittedv2) {
		return version1
	}
	return version2
}

func KillAllST() {
	pids := getListOfPids()
	for i := 0; i < len(pids); i++ {
		stopST(pids[i])
	}
}

func ExtractInfoFromResponse(res *http.Response) map[string]string {
	antiCsrf := res.Header["Anti-Csrf"]
	cookies := res.Header["Set-Cookie"]
	var refreshToken string
	var refreshTokenExpiry string
	var refreshTokenDomain string
	var refreshTokenHttpOnly = "false"
	var idRefreshTokenFromCookie string
	var idRefreshTokenExpiry string
	var idRefreshTokenDomain string
	var idRefreshTokenHttpOnly = "false"
	var accessToken string
	var accessTokenExpiry string
	var accessTokenDomain string
	var accessTokenHttpOnly = "false"
	for _, cookie := range cookies {
		if strings.Split(strings.Split(cookie, ";")[0], "=")[0] == "sRefreshToken" {
			refreshToken = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
			if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " Expires" {
				refreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " expires" {
				refreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else {
				refreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[3], "=")[1]
			}
			if strings.Split(strings.Split(cookie, ";")[1], "=")[0] == " Path" {

			}
			for _, property := range strings.Split(cookie, ";") {
				if strings.Index(property, "HttpOnly") == 1 {
					refreshTokenHttpOnly = "true"
					break
				}
			}
		} else if strings.Split(strings.Split(cookie, ";")[0], "=")[0] == "sIdRefreshToken" {
			idRefreshTokenFromCookie = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
			if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " Expires" {
				idRefreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " expires" {
				idRefreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else {
				idRefreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[3], "=")[1]
			}
			if strings.Split(strings.Split(cookie, ";")[1], "=")[0] == " Path" {
			}
			for _, property := range strings.Split(cookie, ";") {
				if strings.Index(property, "HttpOnly") == 1 {
					idRefreshTokenHttpOnly = "true"
					break
				}
			}
		} else if strings.Split(strings.Split(cookie, ";")[0], "=")[0] == "sAccessToken" {
			accessToken = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
			if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " Expires" {
				accessTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " expires" {
				accessTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else {
				accessTokenExpiry = strings.Split(strings.Split(cookie, ";")[3], "=")[1]
			}
			if strings.Split(strings.Split(cookie, ";")[1], "=")[0] == " Path" {
			}
			for _, property := range strings.Split(cookie, ";") {
				if strings.Index(property, "HttpOnly") == 1 {
					accessTokenHttpOnly = "true"
					break
				}
			}
		}
	}
	return map[string]string{
		"antiCsrf":               antiCsrf[0],
		"sAccessToken":           accessToken,
		"sRefreshToken":          refreshToken,
		"sIdRefreshToken":        idRefreshTokenFromCookie,
		"refreshTokenExpiry":     refreshTokenExpiry,
		"refreshTokenDomain":     refreshTokenDomain,
		"refreshTokenHttpOnly":   refreshTokenHttpOnly,
		"idRefreshTokenExpiry":   idRefreshTokenExpiry,
		"idRefreshTokenDomain":   idRefreshTokenDomain,
		"idRefreshTokenHttpOnly": idRefreshTokenHttpOnly,
		"accessTokenExpiry":      accessTokenExpiry,
		"accessTokenDomain":      accessTokenDomain,
		"accessTokenHttpOnly":    accessTokenHttpOnly,
	}
}

func ExtractInfoFromResponseWhenAntiCSRFisNone(res *http.Response) map[string]string {
	cookies := res.Header["Set-Cookie"]
	var refreshToken string
	var refreshTokenExpiry string
	var refreshTokenDomain string
	var refreshTokenHttpOnly = "false"
	var idRefreshTokenFromCookie string
	var idRefreshTokenExpiry string
	var idRefreshTokenDomain string
	var idRefreshTokenHttpOnly = "false"
	var accessToken string
	var accessTokenExpiry string
	var accessTokenDomain string
	var accessTokenHttpOnly = "false"
	for _, cookie := range cookies {
		if strings.Split(strings.Split(cookie, ";")[0], "=")[0] == "sRefreshToken" {
			refreshToken = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
			if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " Expires" {
				refreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " expires" {
				refreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else {
				refreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[3], "=")[1]
			}
			if strings.Split(strings.Split(cookie, ";")[1], "=")[0] == " Path" {
			}
			for _, property := range strings.Split(cookie, ";") {
				if strings.Index(property, "HttpOnly") == 1 {
					refreshTokenHttpOnly = "true"
					break
				}
			}
		} else if strings.Split(strings.Split(cookie, ";")[0], "=")[0] == "sIdRefreshToken" {
			idRefreshTokenFromCookie = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
			if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " Expires" {
				idRefreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " expires" {
				idRefreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else {
				idRefreshTokenExpiry = strings.Split(strings.Split(cookie, ";")[3], "=")[1]
			}
			if strings.Split(strings.Split(cookie, ";")[1], "=")[0] == " Path" {
			}
			for _, property := range strings.Split(cookie, ";") {
				if strings.Index(property, "HttpOnly") == 1 {
					idRefreshTokenHttpOnly = "true"
					break
				}
			}
		} else if strings.Split(strings.Split(cookie, ";")[0], "=")[0] == "sAccessToken" {
			accessToken = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
			if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " Expires" {
				accessTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else if strings.Split(strings.Split(cookie, ";")[2], "=")[0] == " expires" {
				accessTokenExpiry = strings.Split(strings.Split(cookie, ";")[2], "=")[1]
			} else {
				accessTokenExpiry = strings.Split(strings.Split(cookie, ";")[3], "=")[1]
			}
			if strings.Split(strings.Split(cookie, ";")[1], "=")[0] == " Path" {
			}
			for _, property := range strings.Split(cookie, ";") {
				if strings.Index(property, "HttpOnly") == 1 {
					accessTokenHttpOnly = "true"
					break
				}
			}
		}
	}
	return map[string]string{
		"sAccessToken":           accessToken,
		"sRefreshToken":          refreshToken,
		"sIdRefreshToken":        idRefreshTokenFromCookie,
		"refreshTokenExpiry":     refreshTokenExpiry,
		"refreshTokenDomain":     refreshTokenDomain,
		"refreshTokenHttpOnly":   refreshTokenHttpOnly,
		"idRefreshTokenExpiry":   idRefreshTokenExpiry,
		"idRefreshTokenDomain":   idRefreshTokenDomain,
		"idRefreshTokenHttpOnly": idRefreshTokenHttpOnly,
		"accessTokenExpiry":      accessTokenExpiry,
		"accessTokenDomain":      accessTokenDomain,
		"accessTokenHttpOnly":    accessTokenHttpOnly,
	}
}

func getInstallationDir() string {
	installationDir := os.Getenv("INSTALL_DIR")
	installationDir = "../../" + installationDir
	return installationDir
}

func SetKeyValueInConfig(key string, value string) {
	installationPath := getInstallationDir()
	pathToConfigYamlFile := installationPath + "/config.yaml"
	f, err := os.OpenFile(pathToConfigYamlFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(key + ": " + value + "\n"); err != nil {
		panic(err)
	}
}

func SignupRequest(email string, password string, testUrl string) (*http.Response, error) {
	formFields := map[string][]map[string]string{
		"formFields": {
			{
				"id":    "email",
				"value": email,
			},
			{
				"id":    "password",
				"value": password,
			},
		},
	}

	postBody, err := json.Marshal(formFields)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	resp, err := http.Post(testUrl+"/auth/signup", "application/json", bytes.NewBuffer(postBody))

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return resp, nil
}

func SignInRequest(email string, password string, testUrl string) (*http.Response, error) {
	formFields := map[string][]map[string]string{
		"formFields": {
			{
				"id":    "email",
				"value": email,
			},
			{
				"id":    "password",
				"value": password,
			},
		},
	}

	postBody, err := json.Marshal(formFields)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	resp, err := http.Post(testUrl+"/auth/signin", "application/json", bytes.NewBuffer(postBody))

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return resp, nil
}

func EmailVerifyTokenRequest(testUrl string, userId string, accessToken string, idRefreshTokenFromCookie string, antiCsrf string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, testUrl+"/auth/user/email/verify/token", bytes.NewBuffer([]byte(userId)))
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Cookie", "sAccessToken="+accessToken+";"+"sIdRefreshToken="+idRefreshTokenFromCookie)
	req.Header.Add("anti-csrf", antiCsrf)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return resp, nil
}

func SignoutRequest(testUrl string, accessToken string, idRefreshTokenFromCookie string, antiCsrf string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, testUrl+"/auth/signout", nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Cookie", "sAccessToken="+accessToken+";"+"sIdRefreshToken="+idRefreshTokenFromCookie)
	req.Header.Add("anti-csrf", antiCsrf)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return resp, nil
}

func SessionRefresh(testUrl string, refreshToken string, idRefreshToken string, antiCsrf string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, testUrl+"/auth/session/refresh", nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Cookie", "sRefreshToken="+refreshToken+";"+"sIdRefreshToken="+idRefreshToken)
	req.Header.Add("anti-csrf", antiCsrf)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return resp, nil
}

func ReturnCustomProviderWithAuthRedirectParams() tpmodels.TypeProvider {
	return tpmodels.TypeProvider{
		ID: "custom",
		Get: func(redirectURI, authCodeFromRequest *string, userContext *map[string]interface{}) tpmodels.TypeProviderGetResponse {
			return tpmodels.TypeProviderGetResponse{
				AccessTokenAPI: tpmodels.AccessTokenAPI{
					URL: "https://test.com/oauth/token",
				},
				AuthorisationRedirect: tpmodels.AuthorisationRedirect{
					URL: "https://test.com/oauth/auth",
					Params: map[string]interface{}{
						"scope":     "test",
						"client_id": "supertokens",
						"dynamic": func(req *http.Request) string {
							return req.URL.Query().Get("dynamic")
						},
					},
				},
				GetProfileInfo: func(authCodeResponse interface{}, userContext *map[string]interface{}) (tpmodels.UserInfo, error) {
					return tpmodels.UserInfo{
						ID: "user",
						Email: &tpmodels.EmailStruct{
							ID:         "email@test.com",
							IsVerified: true,
						},
					}, nil
				},
				GetClientId: func(userContext *map[string]interface{}) string {
					return "supertokens"
				},
			}
		},
	}
}

func ReturnCustomProviderWithoutAuthRedirectParams() tpmodels.TypeProvider {
	return tpmodels.TypeProvider{
		ID: "custom",
		Get: func(redirectURI, authCodeFromRequest *string, userContext *map[string]interface{}) tpmodels.TypeProviderGetResponse {
			return tpmodels.TypeProviderGetResponse{
				AccessTokenAPI: tpmodels.AccessTokenAPI{
					URL: "https://test.com/oauth/token",
				},
				AuthorisationRedirect: tpmodels.AuthorisationRedirect{
					URL: "https://test.com/oauth/auth",
				},
				GetProfileInfo: func(authCodeResponse interface{}, userContext *map[string]interface{}) (tpmodels.UserInfo, error) {
					return tpmodels.UserInfo{
						ID: "user",
						Email: &tpmodels.EmailStruct{
							ID:         "email@test.com",
							IsVerified: true,
						},
					}, nil
				},
				GetClientId: func(userContext *map[string]interface{}) string {
					return "supertokens"
				},
			}
		},
	}
}

func SigninupCustomRequest(testServerUrl string, email string, id string) (*http.Response, error) {
	defer gock.OffAll()
	gock.New("https://test.com/").
		Post("oauth/token").
		Reply(200).
		JSON(map[string]interface{}{
			"email": email,
			"id":    id,
		})
	postData := map[string]string{
		"thirdPartyId": "custom",
		"code":         "32432432",
		"redirectURI":  "http://localhost.org",
	}

	postBody, err := json.Marshal(postData)
	if err != nil {
		return nil, err
	}

	gock.New(testServerUrl).EnableNetworking().Persist()
	gock.New("http://localhost:8080/").EnableNetworking().Persist()

	resp, err := http.Post(testServerUrl+"/auth/signinup", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}
	return resp, nil
}
