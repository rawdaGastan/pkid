package client

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	sodium "github.com/gokillers/libsodium-go/cryptosign"
)

type PkidClient struct {
	client     http.Client
	port       int
	privateKey []byte
	publicKey  []byte
}

// create a new instance from the pkid client
func NewPkidClient(privateKey []byte, publicKey []byte, port int) PkidClient {
	client := http.Client{Timeout: 5 * time.Second}

	return PkidClient{
		client:     client,
		port:       port,
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
		decryptedValue, err := encrypt(value, pc.publicKey)
		if err != nil {
			return errors.New("encryption failed with error: " + fmt.Sprint(err))
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

	signedBody, err := SignEncode(payload, pc.privateKey)
	if err != nil {
		return errors.New("error sign body: " + fmt.Sprint(err))
	}

	signedHeader, err := SignEncode(header, pc.privateKey)
	if err != nil {
		return errors.New("error sign header: " + fmt.Sprint(err))
	}

	// set request
	jsonBody := []byte(signedBody)
	bodyReader := bytes.NewReader(jsonBody)

	requestURL := fmt.Sprintf("http://localhost:%v/set/%v/%v/%v", pc.port, hex.EncodeToString(pc.publicKey), project, key)
	request, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		return errors.New("set request failed with error: " + fmt.Sprint(err))
	}

	request.Header.Set("Authorization", signedHeader)
	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return errors.New("set response failed with error: " + fmt.Sprint(err))
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.New("read response body failed: " + fmt.Sprint(err))
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return errors.New("unmarshal response body failed: " + fmt.Sprint(err))
	}

	msg := data["msg"].(string)
	fmt.Println(msg)

	return err
}

// get a value for a key inside a project
func (pc *PkidClient) Get(project string, key string) (string, error) {

	requestURL := fmt.Sprintf("http://localhost:%v/get/%v/%v/%v", pc.port, hex.EncodeToString(pc.publicKey), project, key)
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return "", errors.New("get request failed with error: " + fmt.Sprint(err))
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return "", errors.New("get response failed with error: " + fmt.Sprint(err))
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.New("read response body failed: " + fmt.Sprint(err))
	}

	var data map[string]string
	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", errors.New("unmarshal response body failed: " + fmt.Sprint(err))
	}

	msg := data["msg"]
	fmt.Println(msg)

	signedPayload := data["data"]

	payload, err := verifySignedData(signedPayload, pc.publicKey)
	if err != nil {
		return "", errors.New("verifying data failed: " + fmt.Sprint(err))
	}

	var jsonPayload map[string]interface{}
	err = json.Unmarshal(payload, &jsonPayload)

	if err != nil {
		return "", errors.New("unmarshal payload failed: " + fmt.Sprint(err))
	}

	is_encrypted := jsonPayload["is_encrypted"].(bool)
	value := jsonPayload["payload"].(string)

	if is_encrypted {
		value, err = decrypt(value, pc.publicKey, pc.privateKey)
		if err != nil {
			return "", errors.New("decrypting value failed with error, " + fmt.Sprint(err))
		}
	}

	return value, nil
}

// list all keys for a project
func (pc *PkidClient) List(project string) ([]string, error) {

	requestURL := fmt.Sprintf("http://localhost:%v/list/%v/%v", pc.port, hex.EncodeToString(pc.publicKey), project)
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return []string{}, errors.New("get request failed with error: " + fmt.Sprint(err))
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return []string{}, errors.New("get response failed with error: " + fmt.Sprint(err))
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []string{}, errors.New("read response body failed: " + fmt.Sprint(err))
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return []string{}, errors.New("unmarshal response body failed: " + fmt.Sprint(err))
	}

	interfaceKeys := data["data"].([]interface{})
	keys := make([]string, len(interfaceKeys))
	for i, v := range interfaceKeys {
		keys[i] = v.(string)
	}

	return keys, nil
}

// delete a key with its value inside a project
func (pc *PkidClient) Delete(project string, key string) error {

	requestURL := fmt.Sprintf("http://localhost:%v/delete/%v/%v/%v", pc.port, hex.EncodeToString(pc.publicKey), project, key)
	request, err := http.NewRequest(http.MethodDelete, requestURL, nil)
	if err != nil {
		return errors.New("delete request failed with error: " + fmt.Sprint(err))
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := pc.client.Do(request)
	if err != nil {
		return errors.New("delete response failed with error: " + fmt.Sprint(err))
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.New("read response body failed: " + fmt.Sprint(err))
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return errors.New("unmarshal response body failed: " + fmt.Sprint(err))
	}

	msg := data["msg"].(string)
	fmt.Println(msg)

	return err
}
