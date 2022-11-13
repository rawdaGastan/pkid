package internal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/nacl/sign"
)

// verify the signed data (value) of the set request body
func verifySignedData(w http.ResponseWriter, data string, pk []byte) (bool, error) {

	fmt.Fprintf(w, "start data verification\n")

	// pk in bytes
	verifyPk := [32]byte{}
	copy(verifyPk[:], pk)

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false, err
	}
	fmt.Fprintf(w, "data is decoded\n")

	_, verified := sign.Open(decodedData, decodedData, &verifyPk)
	fmt.Fprintf(w, "signed data is verified: %v \n", verified)

	fmt.Fprintf(w, "end data verification\n")

	return verified, nil
}

// verify the signed header of the set request
func verifySignedHeader(w http.ResponseWriter, header string, pk []byte) (bool, error) {

	fmt.Fprintf(w, "verifying header\n")

	// pk in bytes
	verifyPk := [32]byte{}
	copy(verifyPk[:], pk)

	decodedHeader, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return false, err
	}
	fmt.Fprintf(w, "header is decoded\n")

	decodedHeaderOut := []byte{}

	verifiedSignedHeader, verified := sign.Open(decodedHeaderOut, decodedHeader, &verifyPk)
	fmt.Fprintf(w, "signed header is verified: %v\n", verified)

	jsonHeader := map[string]any{}
	err = json.Unmarshal(verifiedSignedHeader, &jsonHeader)
	if err != nil {
		return false, err
	}

	milliseconds := time.Now().Unix()
	diff := milliseconds - int64(jsonHeader["timestamp"].(float64))
	fmt.Fprintf(w, "timestamp difference is: %v seconds\n", diff)

	if diff > 5 || jsonHeader["intent"].(string) != "pkid.store" {
		return false, errors.New("timestamp difference exceeded 5 seconds, " + fmt.Sprint(diff) + "\n")
	}

	fmt.Fprintf(w, "end header verification\n")

	return verified, nil
}
