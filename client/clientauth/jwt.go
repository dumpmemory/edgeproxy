package clientauth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"edgeproxy/config"
	b64 "encoding/base64"
	"encoding/pem"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

type ClientAuthorizationClaims struct {
	*jwt.StandardClaims
	Nonce string `json:"nonce"`
}

var (
	signKey        *rsa.PrivateKey
	signKeyEc      *ecdsa.PrivateKey
	certificate    *x509.Certificate
	certificatePem string
)

const (
	HeaderAuthorization = "Authorization"
	HeaderCertificate   = "X-Client-Certificate"
)

func SetSigningKey(pemPath string) {
	buf, err := ioutil.ReadFile(pemPath)
	if err != nil {
		log.Fatalf("error loading private key: %v", err)
		return
	}
	block, _ := pem.Decode(buf)
	if block == nil {
		log.Fatalln("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		privateKeyEDCSA, ecErr := x509.ParseECPrivateKey(block.Bytes)
		if ecErr != nil {
			log.Fatalf("error loading private key: %v", err)
			return
		}
		signKeyEc = privateKeyEDCSA
	} else {
		signKey = privateKey
	}

}

func SetCertificate(pemPath string) {
	buf, err := ioutil.ReadFile(pemPath)
	if err != nil {
		log.Fatalf("error loading private key: %v", err)
		return
	}
	block, _ := pem.Decode(buf)
	if block == nil {
		log.Fatalln("failed to parse PEM block containing the key")
	}

	_, certErr := x509.ParseCertificate(block.Bytes)
	if certErr != nil {
		log.Fatalf("error loading private key: %v", certErr)
		return
	}

	certificatePem = b64.StdEncoding.EncodeToString(buf)
}

func CreateClientToken() (string, error) {
	if signKey != nil {
		t := jwt.New(jwt.GetSigningMethod("RS256"))
		t.Claims = &ClientAuthorizationClaims{
			&jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
				Audience:  "edgeproxy",
			},
			strconv.Itoa(rand.Int()),
		}
		return t.SignedString(signKey)
	} else if signKeyEc != nil {
		t := jwt.New(jwt.GetSigningMethod("ES256"))
		t.Claims = &ClientAuthorizationClaims{
			&jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
				Audience:  "edgeproxy",
			},
			strconv.Itoa(rand.Int()),
		}
		return t.SignedString(signKeyEc)
	}
	return "", nil
}

func GetClientCertificate() (string, error) {

	return certificatePem, nil
}

type JwtAuthenticator struct {
}

func (receiver JwtAuthenticator) AddAuthenticationHeaders(headers *http.Header) {
	token, _ := CreateClientToken()
	cert, _ := GetClientCertificate()
	headers.Add(HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
	headers.Add(HeaderCertificate, cert)
}

func (receiver JwtAuthenticator) Load(config config.ClientAuthCaConfig) {
	SetCertificate(config.Certificate)
	SetSigningKey(config.Key)
}
