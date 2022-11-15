package client

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/GoKillers/libsodium-go/cryptobox"
	sodium "github.com/gokillers/libsodium-go/cryptosign"
	"golang.org/x/crypto/nacl/sign"
)

// sign a msg using public key
func signMsg(message []byte, privateKey []byte) ([]byte, int) {
	return sodium.CryptoSign(message, privateKey)
}

// sign a msg then encode it
func SignEncode(payload map[string]interface{}, privateKey []byte) (string, error) {
	message, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	signed, _ := signMsg(message, privateKey)

	return base64.StdEncoding.EncodeToString(signed), nil
}

// verify the signed data (value) of the set request body
func verifySignedData(data string, pk []byte) ([]byte, error) {

	// pk in bytes
	verifyPk := [32]byte{}
	copy(verifyPk[:], pk)

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return []byte{}, err
	}

	decodedDataOut := []byte{}
	verifiedContent, verified := sign.Open(decodedDataOut, decodedData, &verifyPk)

	if !verified {
		return []byte{}, errors.New("verifying data failed, " + fmt.Sprint(verified))
	}

	return verifiedContent, nil
}

func encrypt(payload string, publicKey []byte) (string, error) {
	message, err := json.Marshal(payload)

	if err != nil {
		return "", errors.New("marshal payload failed with error, " + fmt.Sprint(err))
	}

	curvePublicKey, _ := sodium.CryptoSignEd25519PkToCurve25519(publicKey)
	encryptedMessage, _ := cryptobox.CryptoBoxSeal(message, curvePublicKey)

	return base64.StdEncoding.EncodeToString(encryptedMessage), nil
}

func decrypt(cipher string, publicKey []byte, privateKey []byte) (string, error) {
	decodedCipher, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", errors.New("decoding cipher text failed with error, " + fmt.Sprint(err))
	}

	curvePublicKey, _ := sodium.CryptoSignEd25519PkToCurve25519(publicKey)
	curvePrivateKey, _ := sodium.CryptoSignEd25519SkToCurve25519(privateKey)

	decrypted, _ := cryptobox.CryptoBoxSealOpen(decodedCipher, curvePublicKey, curvePrivateKey)

	return string(decrypted), nil
}
