package client

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	sodium "github.com/gokillers/libsodium-go/cryptosign"
	"github.com/rawdaGastan/pkid/pkg"
)

type PkidClient struct {
	client     http.Client
	serverUrl  string
	privateKey []byte
	publicKey  []byte
}

// create a new instance from the pkid client
func NewPkidClient(privateKey []byte, publicKey []byte, url string, timeout time.Duration) PkidClient {
	client := http.Client{Timeout: timeout}

	return PkidClient{
		client:     client,
		serverUrl:  url,
		privateKey: privateKey,
		publicKey:  publicKey,
	}

}

// for testing with given client
func NewPkidClientWithHTTPClient(privateKey []byte, publicKey []byte, url string, client *http.Client) PkidClient {
	return PkidClient{
		client:     *client,
		serverUrl:  url,
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// generate a private key and public key for the client
func GenerateKeyPair() ([]byte, []byte) {
	privateKey, publicKey, _ := sodium.CryptoSignKeyPair()
	return privateKey, publicKey
}

// generate a private key and public key for the client using TF login seed
func GenerateKeyPairUsingSeed(seed string) ([]byte, []byte, error) {
	decodedSeed, err := base64.StdEncoding.DecodeString(seed)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	privateKey, publicKey, _ := sodium.CryptoSignSeedKeyPair(decodedSeed)
	return privateKey, publicKey, nil
}

// set a new value for a key inside a project
func (pc *PkidClient) Set(project string, key string, value string, willEncrypt bool) error {

	if willEncrypt {
		decryptedValue, err := pkg.Encrypt(value, pc.publicKey)
		if err != nil {
			return fmt.Errorf("encryption failed with error: %w", err)
		}

		value = decryptedValue
	}

	header := map[string]interface{}{
		"intent":    "pkid.store",
		"timestamp": time.Now().Unix(),
	}

	payload := map[string]interface{}{
		"is_encrypted": willEncrypt,
		"payload":      value,
		"data_version": 1,
	}

	signedBody, err := pkg.SignEncode(payload, pc.privateKey)
	if err != nil {
		return fmt.Errorf("error sign body: %w", err)
	}

	signedHeader, err := pkg.SignEncode(header, pc.privateKey)
	if err != nil {
		return fmt.Errorf("error sign header: %w", err)
	}

	// set request
	jsonBody := []byte(signedBody)
	bodyReader := bytes.NewReader(jsonBody)

	requestURL := fmt.Sprintf("%v/%v/%v/%v", pc.serverUrl, hex.EncodeToString(pc.publicKey), project, key)
	request, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		return fmt.Errorf("set request failed with error: %w", err)
	}

	request.Header.Set("Authorization", signedHeader)
	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return fmt.Errorf("set response failed with error: %w", err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return fmt.Errorf("unmarshal response body failed: %w", err)
	}

	return nil
}

// get a value for a key inside a project
func (pc *PkidClient) Get(project string, key string) (string, error) {

	requestURL := fmt.Sprintf("%v/%v/%v/%v", pc.serverUrl, hex.EncodeToString(pc.publicKey), project, key)
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("get request failed with error: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("get response failed with error: %w", err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read response body failed with error: %w", err)
	}

	var data map[string]string
	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", fmt.Errorf("unmarshal response body failed with error: %w", err)
	}

	signedPayload := data["data"]

	payload, err := pkg.VerifySignedData(signedPayload, pc.publicKey)
	if err != nil {
		return "", fmt.Errorf("verifying data failed with error: %w", err)
	}

	var jsonPayload map[string]interface{}
	err = json.Unmarshal(payload, &jsonPayload)

	if err != nil {
		return "", fmt.Errorf("unmarshal payload failed with error: %w", err)
	}

	is_encrypted := jsonPayload["is_encrypted"].(bool)
	value := jsonPayload["payload"].(string)

	if is_encrypted {
		value, err = pkg.Decrypt(value, pc.publicKey, pc.privateKey)
		if err != nil {
			return "", fmt.Errorf("decrypting value failed with error: %w", err)
		}
	}

	return value, nil
}

// list all keys for a project
func (pc *PkidClient) List(project string) ([]string, error) {

	requestURL := fmt.Sprintf("%v/%v/%v", pc.serverUrl, hex.EncodeToString(pc.publicKey), project)
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return []string{}, fmt.Errorf("get request failed with error: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return []string{}, fmt.Errorf("get response failed with error: %w", err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []string{}, fmt.Errorf("read response body failed with error: %w", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return []string{}, fmt.Errorf("unmarshal response body failed with error: %w", err)
	}

	interfaceKeys := data["data"].([]interface{})
	keys := make([]string, len(interfaceKeys))
	for i, v := range interfaceKeys {
		keys[i] = v.(string)
	}

	return keys, nil
}

// delete a key with its value inside a project
func (pc *PkidClient) DeleteProject(project string) error {

	requestURL := fmt.Sprintf("%v/%v/%v", pc.serverUrl, hex.EncodeToString(pc.publicKey), project)
	request, err := http.NewRequest(http.MethodDelete, requestURL, nil)
	if err != nil {
		return fmt.Errorf("delete request failed with error: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return fmt.Errorf("delete response failed with error: %w", err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response body failed with error: %w", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return fmt.Errorf("unmarshal response body failed with error: %w", err)
	}

	return nil
}

// delete a key with its value inside a project
func (pc *PkidClient) Delete(project string, key string) error {

	requestURL := fmt.Sprintf("%v/%v/%v/%v", pc.serverUrl, hex.EncodeToString(pc.publicKey), project, key)
	request, err := http.NewRequest(http.MethodDelete, requestURL, nil)
	if err != nil {
		return fmt.Errorf("delete request failed with error: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return fmt.Errorf("delete response failed with error: %w", err)
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response body failed with error: %w", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return fmt.Errorf("unmarshal response body failed with error: %w", err)
	}

	return nil
}
