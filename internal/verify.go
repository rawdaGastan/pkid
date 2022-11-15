package internal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/nacl/sign"
)

// verify the signed data (value) of the set request body
func verifySignedData(logger zerolog.Logger, data string, pk []byte) (bool, error) {

	logger.Debug().Msg("start data verification")

	// pk in bytes
	verifyPk := [32]byte{}
	copy(verifyPk[:], pk)

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false, err
	}
	logger.Debug().Msg("data is decoded")

	decodedDataOut := []byte{}
	_, verified := sign.Open(decodedDataOut, decodedData, &verifyPk)
	logger.Debug().Msg("signed data is verified: " + fmt.Sprint(verified))

	logger.Debug().Msg("end data verification")

	return verified, nil
}

// verify the signed header of the set request
func verifySignedHeader(logger zerolog.Logger, header string, pk []byte) (bool, error) {

	logger.Debug().Msg("verifying header")

	// pk in bytes
	verifyPk := [32]byte{}
	copy(verifyPk[:], pk)

	decodedHeader, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return false, err
	}
	logger.Debug().Msg("header is decoded")

	decodedHeaderOut := []byte{}

	verifiedSignedHeader, verified := sign.Open(decodedHeaderOut, decodedHeader, &verifyPk)
	logger.Debug().Msg("signed header is verified: " + fmt.Sprint(verified))

	jsonHeader := map[string]any{}
	err = json.Unmarshal(verifiedSignedHeader, &jsonHeader)
	if err != nil {
		return false, err
	}

	milliseconds := time.Now().Unix()
	diff := milliseconds - int64(jsonHeader["timestamp"].(float64))
	logger.Debug().Msg("timestamp difference is: " + fmt.Sprint(diff) + " seconds")

	if diff > 5 || jsonHeader["intent"].(string) != "pkid.store" {
		return false, errors.New("timestamp difference exceeded 5 seconds, " + fmt.Sprint(diff))
	}

	logger.Debug().Msg("end header verification")

	return verified, nil
}
