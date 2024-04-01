package pemkey

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func SerializePublicKeyToPEM(pubKey *rsa.PublicKey) string {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM)
}

func GeneratePEMFile() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateFile, err := os.Create("private_key.pem")
	if err != nil {
		return err
	}
	defer privateFile.Close()

	privatePEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateFile, privatePEM); err != nil {
		return err
	}

	publicKey := &privateKey.PublicKey

	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	publicPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	}

	publicFile, err := os.Create("public_key.pem")
	if err != nil {
		return err
	}

	defer publicFile.Close()

	if err := pem.Encode(publicFile, publicPEM); err != nil {
		return err
	}

	return nil
}

func LoadPrivateKeyFile(file string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func LoadPublicKeyFile() []byte {
	key, err := os.ReadFile("public_key.pem")
	if err != nil {
		log.Fatal(err)
	}

	return key
}
