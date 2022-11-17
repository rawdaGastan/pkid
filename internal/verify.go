package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/nacl/sign"
)

// verify the signed data (value) of the set request body
func verifySignedData(data string, pk []byte) (bool, error) {

	// pk in bytes
	verifyPk := [32]byte{}
	copy(verifyPk[:], pk)

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false, err
	}

	decodedDataOut := []byte{}
	_, verified := sign.Open(decodedDataOut, decodedData, &verifyPk)

	return verified, nil
}

// verify the signed header of the set request
func verifySignedHeader(header string, pk []byte) (bool, error) {

	// pk in bytes
	verifyPk := [32]byte{}
	copy(verifyPk[:], pk)

	decodedHeader, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return false, err
	}

	decodedHeaderOut := []byte{}

	verifiedSignedHeader, verified := sign.Open(decodedHeaderOut, decodedHeader, &verifyPk)

	jsonHeader := map[string]any{}
	err = json.Unmarshal(verifiedSignedHeader, &jsonHeader)
	if err != nil {
		return false, err
	}

	milliseconds := time.Now().Unix()
	diff := milliseconds - int64(jsonHeader["timestamp"].(float64))

	if diff > 5 || jsonHeader["intent"].(string) != "pkid.store" {
		return false, fmt.Errorf("timestamp difference exceeded 5 seconds, %v", diff)
	}

	return verified, nil
}
