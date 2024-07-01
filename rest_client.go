package RadwareGoSDK

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type New_Client struct {
	HostIP string
	HTTPRequest *http.Request
	// Token      string
}

var Client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func Login(product string, host_ip string, username string, password string) (*New_Client, int, string, error) {
	new_client := New_Client{
		HostIP: host_ip,
	}
	if "ALTEON" == strings.ToUpper(product) {
		login_request, err := http.NewRequest("POST", "https://"+new_client.HostIP, nil)
		if err != nil {
			// fmt.Println("Error creating login request:", err)
			return &new_client, 0, "Error creating login request", err
		}
		login_request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
		login_request.Header.Set("Content-Type", "application/json")
		new_client.HTTPRequest = login_request
		return &new_client, 200, "Login Successful", nil
	} else if "CYBERCONTROLLER" == strings.ToUpper(product) {
		Data := map[string]string{"username": username, "password": password}
		Bytes, err := json.Marshal(Data)
		if err != nil {
			// fmt.Println("Error encoding JSON:", err)
			return nil, 0, "Error encoding JSON", err
		}
		login_request, err := http.NewRequest("POST", "https://"+new_client.HostIP+"/mgmt/system/user/login", bytes.NewBuffer(Bytes))
		if err != nil {
			// fmt.Println("Error creating login request:", err)
			return nil, 0, "Error creating login request", err
		}
		login_request.Header.Set("Content-Type", "application/json")
		login_response, err := Client.Do(login_request)
		if err != nil {
			// fmt.Println("Error making login request:", err)
			return &new_client, 0, "Error making login request", err
		}
		defer login_response.Body.Close()
		login_response_body, err := ioutil.ReadAll(login_response.Body)
		if err != nil {
			// fmt.Println("Error reading API response body:", err)
			return &new_client, login_response.StatusCode, string(login_response_body), err
		}
		if login_response.StatusCode == 401 {
			return nil, login_response.StatusCode, string(login_response_body), err
		}
		JsessionID, err := extractJSessionID(login_response)
		if err != nil {
			// fmt.Println("Error extracting JSESSIONID:", err)
			return &new_client, 0, "Error extracting JSESSIONID", err
		}
		
		
		new_client.HTTPRequest = login_request
		new_client.HTTPRequest.Header.Set("Content-Type", "application/json")
		new_client.HTTPRequest.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s", JsessionID))
		// login_request.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s", JsessionID))
		// c.Token := login_request.Header.Clone()
		return &new_client, login_response.StatusCode, string(login_response_body), nil
	}
	return &new_client, 0, "Provide proper product name, Ex. ALTEON or CYBERCONTROLLER", nil
}

func (new_client *New_Client) Request(method string, API string, Data []byte, additional_header map[string]string) (int, string, error) {
	URL := "https://" + new_client.HostIP + API
	New_Request, err := http.NewRequest(strings.ToUpper(method), URL, bytes.NewBuffer(Data))
	if err != nil {
		return 0, "Error creating API request", err
	}

	New_Request.Header = new_client.HTTPRequest.Header.Clone()
	fmt.Println(New_Request.Header)
	for key, value := range additional_header {
		New_Request.Header.Set(key, value)
	}
	APIResponse, err := Client.Do(New_Request)
	if err != nil {
		return 0, "Error making API request", err
	}
	defer APIResponse.Body.Close()

	// Read the response body of the API call
	APIResponseBody, err := ioutil.ReadAll(APIResponse.Body)
	if err != nil {
		return APIResponse.StatusCode, "Error reading API response body", err
	}

	return APIResponse.StatusCode, string(APIResponseBody), nil
}

func (new_client *New_Client) Logout(product string) (int, string, error) {
	url := ""
	if "ALTEON" == strings.ToUpper(product) {
		url = "https://" + new_client.HostIP + "/config/logout"
	} else if "CYBERCONTROLLER" == strings.ToUpper(product) {
		url = "https://" + new_client.HostIP + "/mgmt/system/user/logout"
	}
	New_Request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		// fmt.Println("Error creating login request:", err)
		return 0, "Error creating logout request", err
	}
	New_Request.Header = new_client.HTTPRequest.Header.Clone()
	APIResponse, err := Client.Do(New_Request)
	if err != nil {
		return 0, "Error making API request", err
	}
	defer APIResponse.Body.Close()

	// Read the response body of the API call
	APIResponseBody, err := ioutil.ReadAll(APIResponse.Body)
	if err != nil {
		return APIResponse.StatusCode, "Error reading API response body", err
	}

	return APIResponse.StatusCode, string(APIResponseBody), nil
}

func extractJSessionID(response *http.Response) (string, error) {
	for _, cookie := range response.Cookies() {
		if cookie.Name == "JSESSIONID" {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("JSESSIONID not found in cookies")
}
