package auth

import (
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"edgeproxy/client/clientauth"
	"edgeproxy/config"
	b64 "encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type spireAuthorizer struct {
	ctx         context.Context
	rootPath    string
	roots       *x509.CertPool
	trustDomain string
	pathConfig  config.PathsConfig
}

// HardFail determines whether the failure to check the revocation
// status of a certificate (i.e. due to network failure) causes
// verification to fail (a hard failure).
var HardFail = false

var HTTPClient = http.DefaultClient
var crlRead = ioutil.ReadAll
var remoteRead = ioutil.ReadAll

// CRLSet associates a PKIX certificate list with the URL the CRL is
// fetched from.
var CRLSet = map[string]*pkix.CertificateList{}
var crlLock = new(sync.Mutex)

type SpiffeSubject struct {
	cert *x509.Certificate
}

func (s SpiffeSubject) GetSubject() string {
	//TODO implement me
	return s.cert.URIs[0].String()
}

func NewSpireAuthorizer(ctx context.Context, caConfig config.ServerAuthCaConfig) Authenticate {
	authorizer := &spireAuthorizer{
		ctx:         ctx,
		rootPath:    caConfig.TrustedRoot,
		trustDomain: caConfig.SpireTrustDomain,
		pathConfig:  caConfig.Paths,
	}

	buf, err := ioutil.ReadFile(caConfig.TrustedRoot)
	if err != nil {
		log.Fatalf("error loading root certs: %v", err)
		return nil
	}

	authorizer.roots = x509.NewCertPool()
	ok := authorizer.roots.AppendCertsFromPEM(buf)
	if !ok {
		log.Fatalf("error loaded root cert pem")
	}

	return authorizer
}

func IsValidToken(token string, pub interface{}) bool {
	var extractToken = regexp.MustCompile(`^Bearer (.*)$`)

	bearerMatch := extractToken.FindStringSubmatch(token)
	if len(bearerMatch) != 2 {
		//not in bearer token format
		log.Debugf("bad auth header %s", token)
		return false
	}

	bearerToken := strings.TrimSpace(bearerMatch[1])

	// TODO: allow more than one signing key
	parsedToken, err := jwt.ParseWithClaims(bearerToken, &clientauth.ClientAuthorizationClaims{}, func(token *jwt.Token) (interface{}, error) {
		return pub, nil
	})
	if err != nil {
		log.Debugf("error validting authentication client: %v", err)
		return false
	}
	claims := parsedToken.Claims.(*clientauth.ClientAuthorizationClaims)
	// didn't blow up, meaning it's signed by the right key and not expired
	if claims.StandardClaims.VerifyAudience("edgeproxy", true) {
		return true
	} else {
		log.Debugf("bad audience: %s", claims.StandardClaims.Audience)
		return false
	}
}

func (f *spireAuthorizer) Authenticate(w http.ResponseWriter, r *http.Request) (bool, Subject) {
	//var port int
	//var err error
	token := r.Header.Get(clientauth.HeaderAuthorization)
	cert := r.Header.Get(clientauth.HeaderCertificate)
	//dstAddr := r.Header.Get(transport.HeaderDstAddress)
	if token == "" {
		return false, nil
	}

	if cert == "" {
		return false, nil
	}

	validatedCert, certErr := f.validateClientCertificate(cert)
	if certErr != nil {
		log.Debugf("error validating cert: %v", certErr)
		return false, nil
	}

	// make sure jwt is signed by same key as cert
	validToken := IsValidToken(token, validatedCert.PublicKey)
	if !validToken {
		log.Debug("invalid token")
		return false, nil
	}

	subj := NewSpiffeSubject(validatedCert)
	return true, subj
}

func NewSpiffeSubject(cert *x509.Certificate) Subject {
	subj := SpiffeSubject{
		cert: cert,
	}
	return subj
}

// shamelessly stolen from https://github.com/cloudflare/cfssl/blob/master/revoke/revoke.go
// revCheck should check the certificate for any revocations. It
// returns a pair of booleans: the first indicates whether the certificate
// is revoked, the second indicates whether the revocations were
// successfully checked.. This leads to the following combinations:
//
//  false, false: an error was encountered while checking revocations.
//
//  false, true:  the certificate was checked successfully and
//                  it is not revoked.
//
//  true, true:   the certificate was checked successfully and
//                  it is revoked.
//
//  true, false:  failure to check revocation status causes
//                  verification to fail
func revCheck(cert *x509.Certificate) (revoked, ok bool, err error) {
	for _, url := range cert.CRLDistributionPoints {
		if revoked, ok, err := certIsRevokedCRL(cert, url); !ok {
			log.Warning("error checking revocation via CRL")
			if HardFail {
				return true, false, err
			}
			return false, false, err
		} else if revoked {
			log.Info("certificate is revoked via CRL")
			return true, true, err
		}
	}

	// TODO: OCSP support
	//if revoked, ok, err := certIsRevokedOCSP(cert, HardFail); !ok {
	//	log.Warning("error checking revocation via OCSP")
	//	if HardFail {
	//		return true, false, err
	//	}
	//	return false, false, err
	//} else if revoked {
	//	log.Info("certificate is revoked via OCSP")
	//	return true, true, err
	//}

	return false, true, nil
}

// fetchCRL fetches and parses a CRL.
func fetchCRL(url string) (*pkix.CertificateList, error) {
	resp, err := HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, errors.New("failed to retrieve CRL")
	}

	body, err := crlRead(resp.Body)
	if err != nil {
		return nil, err
	}
	return x509.ParseCRL(body)
}

func getIssuer(cert *x509.Certificate) *x509.Certificate {
	var issuer *x509.Certificate
	var err error
	for _, issuingCert := range cert.IssuingCertificateURL {
		issuer, err = fetchRemote(issuingCert)
		if err != nil {
			continue
		}
		break
	}

	return issuer

}

func fetchRemote(url string) (*x509.Certificate, error) {
	resp, err := HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	in, err := remoteRead(resp.Body)
	if err != nil {
		return nil, err
	}

	p, _ := pem.Decode(in)
	if p != nil {
		return ParseCertificatePEM(in)
	}

	return x509.ParseCertificate(in)
}

func ParseCertificatePEM(certPEM []byte) (*x509.Certificate, error) {
	certPEM = bytes.TrimSpace(certPEM)
	cert, rest, err := ParseOneCertificateFromPEM(certPEM)
	if err != nil {
		// Log the actual parsing error but throw a default parse error message.
		log.Debugf("Certificate parsing error: %v", err)
		return nil, err
	} else if cert == nil {
		return nil, errors.New("cert decode failed")
	} else if len(rest) > 0 {
		return nil, errors.New("the PEM file should contain only one object")
	} else if len(cert) > 1 {
		return nil, errors.New("the PKCS7 object in the PEM file should contain only one certificate")
	}
	return cert[0], nil
}

func ParseCertificatesPEM(certsPEM []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	var err error
	certsPEM = bytes.TrimSpace(certsPEM)
	for len(certsPEM) > 0 {
		var cert []*x509.Certificate
		cert, certsPEM, err = ParseOneCertificateFromPEM(certsPEM)
		if err != nil {

			return nil, errors.New("error parsing pem")
		} else if cert == nil {
			break
		}

		certs = append(certs, cert...)
	}
	if len(certsPEM) > 0 {
		return nil, errors.New("decode failed")
	}
	return certs, nil
}

func ParseOneCertificateFromPEM(certsPEM []byte) ([]*x509.Certificate, []byte, error) {

	block, rest := pem.Decode(certsPEM)
	if block == nil {
		return nil, rest, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, rest, err
	}
	var certs = []*x509.Certificate{cert}
	return certs, rest, nil
}

func certIsRevokedCRL(cert *x509.Certificate, url string) (revoked, ok bool, err error) {
	crlLock.Lock()
	crl, ok := CRLSet[url]
	if ok && crl == nil {
		ok = false
		delete(CRLSet, url)
	}
	crlLock.Unlock()

	var shouldFetchCRL = true
	if ok {
		if !crl.HasExpired(time.Now()) {
			shouldFetchCRL = false
		}
	}

	issuer := getIssuer(cert)

	if shouldFetchCRL {
		var err error
		crl, err = fetchCRL(url)
		if err != nil {
			log.Warningf("failed to fetch CRL: %v", err)
			return false, false, err
		}

		// check CRL signature
		if issuer != nil {
			err = issuer.CheckCRLSignature(crl)
			if err != nil {
				log.Warningf("failed to verify CRL: %v", err)
				return false, false, err
			}
		}

		crlLock.Lock()
		CRLSet[url] = crl
		crlLock.Unlock()
	}

	for _, revoked := range crl.TBSCertList.RevokedCertificates {
		if cert.SerialNumber.Cmp(revoked.SerialNumber) == 0 {
			log.Info("Serial number match: intermediate is revoked.")
			return true, true, err
		}
	}

	return false, true, err
}

func (f *spireAuthorizer) validateClientCertificate(certificate string) (*x509.Certificate, error) {
	sDec, _ := b64.StdEncoding.DecodeString(certificate)
	block, _ := pem.Decode([]byte(sDec))
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatalf("error loading private key: %v", err)
		return nil, err
	}

	// TODO: think about intermediates in verify

	opts := x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		Roots:     f.roots,
	}
	validationChain, err := cert.Verify(opts)
	if err != nil {
		return nil, err
	}

	if !time.Now().Before(cert.NotAfter) {
		msg := fmt.Sprintf("Certificate expired %s\n", cert.NotAfter)
		log.Info(msg)
		return nil, fmt.Errorf(msg)
	} else if !time.Now().After(cert.NotBefore) {
		msg := fmt.Sprintf("Certificate isn't valid until %s\n", cert.NotBefore)
		log.Info(msg)
		return nil, fmt.Errorf(msg)
	}

	revoked, ok, revCheckErr := revCheck(cert)
	if revoked {
		return nil, errors.New(fmt.Sprintf("certificate has been revoked for %s, %v", cert.Subject, revCheckErr))
	}
	if !ok {
		// TODO: decide how strict we wanna be if we can't check CRL
		log.Info("error reading CRL")
	}
	if revCheckErr != nil {
		log.Errorf("Error with CRL: %s", revCheckErr)
	}

	validatedClientCert := validationChain[0][0]
	if len(validatedClientCert.URIs) > 0 {
		// TODO: think how to handle multiple SVIDs
		uri := validatedClientCert.URIs[0]
		if uri.Scheme == "spiffe" {
			trustDomain := uri.Host
			path := uri.Path
			if trustDomain == f.trustDomain {
				if f.pathConfig.AllowedPath(path) {
					log.Debugf("allowed subject '%s' claiming certificate from %v", uri, validationChain[0][len(validationChain[0])-1].Subject)
					return validatedClientCert, nil
				} else {
					log.Debugf("forbidden subject: %s", uri)
					return nil, errors.New("forbidden subject")
				}
			} else {
				log.Debugf("untrusted spiffe domain: %s", trustDomain)
				return nil, errors.New("untrusted spiffe domain")

			}
		}

	}

	return nil, errors.New("certificate does not look like a SVID")
}
