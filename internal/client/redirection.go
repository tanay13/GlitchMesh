package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func makePostRedirection(redirectionUrl string, body string) error {

	requestBuffer, _ := json.Marshal(body)

	requestBody := bytes.NewBuffer(requestBuffer)

	_, err := http.Post(redirectionUrl, "application/json", requestBody)

	if err != nil {
		return fmt.Errorf("Error while making post call %v", err)
	}

	return nil

}

func MakeGetRedirection(redirectionURL string) ([]byte, error) {

	response, err := http.Get(redirectionURL)

	if err != nil {
		return nil, fmt.Errorf("error making GET call %v", err)
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return responseBody, fmt.Errorf("error while reading the response body %v", err)
	}

	return responseBody, nil

}
