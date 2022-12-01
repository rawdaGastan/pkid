package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	sodium "github.com/GoKillers/libsodium-go/cryptosign"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/sign"
)

// sign a msg using public key
func signMsg(message []byte, privateKey []byte) ([]byte, int) {
	return sodium.CryptoSign(message, privateKey)
}

// SignEncode signs a msg then encode it
func SignEncode(payload map[string]interface{}, privateKey []byte) (string, error) {
	message, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	signed, _ := signMsg(message, privateKey)

	return base64.StdEncoding.EncodeToString(signed), nil
}

// VerifySignedData verifies the signed data (value) of the set request body
func VerifySignedData(data string, pk []byte) ([]byte, error) {

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
		return []byte{}, fmt.Errorf("verifying data failed, %v", verified)
	}

	return verifiedContent, nil
}

// Encrypt encrypts a payload with the public key
func Encrypt(payload string, publicKey []byte) (string, error) {
	message, err := json.Marshal(payload)

	// marshal string will new cause an error
	if err != nil {
		return "", err
	}

	curvePublicKey, _ := sodium.CryptoSignEd25519PkToCurve25519(publicKey)
	var encryptedMessage []byte
	encryptedMessage, err = box.SealAnonymous(encryptedMessage, message, (*[32]byte)(curvePublicKey), nil)
	if err != nil {
		return "", err
	}
	//encryptedMessage, _ := cryptobox.CryptoBoxSeal(message, curvePublicKey)

	return base64.StdEncoding.EncodeToString(encryptedMessage), nil
}

// Decrypt decrypts a cipher with the public key and private key
func Decrypt(cipher string, publicKey []byte, privateKey []byte) (string, error) {
	decodedCipher, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", fmt.Errorf("decoding cipher text failed with error: %w", err)
	}

	curvePublicKey, _ := sodium.CryptoSignEd25519PkToCurve25519(publicKey)
	curvePrivateKey, _ := sodium.CryptoSignEd25519SkToCurve25519(privateKey)

	var decrypted []byte
	decrypted, ok := box.OpenAnonymous(decrypted, decodedCipher, (*[32]byte)(curvePublicKey), (*[32]byte)(curvePrivateKey))
	if !ok {
		return "", fmt.Errorf("decrypting failed")
	}
	//decrypted, _ := cryptobox.CryptoBoxSealOpen(decodedCipher, curvePublicKey, curvePrivateKey)

	return string(decrypted), nil
}
