package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	sodium "github.com/gokillers/libsodium-go/cryptosign"
)

var client = http.Client{Timeout: 5 * time.Second}

func sign(message string, privateKey []byte) ([]byte, int) {
	return sodium.CryptoSign([]byte(message), privateKey)
}

func signEncode(payload map[string]interface{}, privateKey []byte) (string, error) {
	stringPayload, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	message := string(stringPayload)
	signed, _ := sign(message, privateKey)

	return base64.StdEncoding.EncodeToString(signed), nil
}

func setRequest(key string, value string, publicKey []byte, privateKey []byte) error {

	header := map[string]interface{}{
		"intent":    "pkid.store",
		"timestamp": time.Now().Unix(),
	}

	payload := map[string]interface{}{
		"is_encrypted": false,
		"payload":      value,
		"data_version": 1,
	}

	signedBody, _ := signEncode(payload, privateKey)
	signedHeader, err := signEncode(header, privateKey)
	if err != nil {
		return errors.New("error sign header: " + fmt.Sprint(err))
	}

	// set request
	jsonBody := []byte(signedBody)
	bodyReader := bytes.NewReader(jsonBody)

	requestURL := fmt.Sprintf("http://localhost:3000/set/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", key)
	request, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		return errors.New("set request failed with error: " + fmt.Sprint(err))
	}

	request.Header.Set("Authorization", signedHeader)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return errors.New("set response failed with error: " + fmt.Sprint(err))
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	return err
}

func getRequest(key string, publicKey []byte) error {

	requestURL := fmt.Sprintf("http://localhost:3000/get/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", key)
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return errors.New("get request failed with error: " + fmt.Sprint(err))
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return errors.New("get response failed with error: " + fmt.Sprint(err))
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	return err
}

func deleteRequest(key string, publicKey []byte) error {

	requestURL := fmt.Sprintf("http://localhost:3000/delete/%v/%v/%v", hex.EncodeToString(publicKey), "pkid", key)
	request, err := http.NewRequest(http.MethodDelete, requestURL, nil)
	if err != nil {
		return errors.New("delete request failed with error: " + fmt.Sprint(err))
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return errors.New("delete response failed with error: " + fmt.Sprint(err))
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	return err
}

func test() {

	privateKey, publicKey, _ := sodium.CryptoSignKeyPair()
	//fmt.Println(publicKey, privateKey)

	err := setRequest("key", "value", publicKey, privateKey)
	if err != nil {
		fmt.Printf("err set: %v\n", err)
	}

	err = getRequest("key", publicKey)
	if err != nil {
		fmt.Printf("err get: %v\n", err)
	}

	err = deleteRequest("key", publicKey)
	if err != nil {
		fmt.Printf("err delete: %v\n", err)
	}

}
