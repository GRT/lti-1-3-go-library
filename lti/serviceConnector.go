package lti

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/segmentio/ksuid"

	"github.com/GRT/lti-1-3-go-library/registrationDatastore"
)

// ServiceConnector helper util that can hit endpoints associated with an lti1.3 launch
type ServiceConnector struct {
	registration registrationDatastore.Registration
	tokenMap     map[string]string
}

// NewServiceConnector creates a new ServiceConnector
func NewServiceConnector(reg registrationDatastore.Registration) *ServiceConnector {
	return &ServiceConnector{registration: reg, tokenMap: make(map[string]string)}
}

func (s *ServiceConnector) getAccessToken(scopes []string) (string, error) {
	sort.Strings(scopes)
	scopeStr := strings.Join(scopes, " ")
	if cachedToken, exists := s.tokenMap[scopeStr]; exists {
		return cachedToken, nil
	}

	privkey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(s.registration.ToolPrivateKey))
	if err != nil {
		return "", errors.Wrapf(err, "GetAccessToken: Error getting Tool Private Key for clientId: %q.", s.registration.ClientID)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": toolIssuer,
		"sub": s.registration.ClientID,
		"aud": s.registration.AuthTokenURL,
		"iat": time.Now().Unix(),
		"exp": time.Now().Unix() + 60,
		"jti": fmt.Sprintf("lti-service-token-%s", ksuid.New().String()),
	})
	tokenStr, err := token.SignedString(privkey)
	if err != nil {
		return "", errors.Wrapf(err, "GetAccessToken: Error signing token for clientId: %q.", s.registration.ClientID)
	}
	// log.Printf("jwt generated: %s", tokenStr)

	client := &http.Client{Timeout: time.Second * 30}
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	form.Add("client_assertion", tokenStr)
	form.Add("scope", scopeStr)
	log.Printf("Access Token Fetch Url: %s", s.registration.AuthTokenURL)
	// log.Printf("                  Form: %+v", form)

	req, err := http.NewRequest("POST", s.registration.AuthTokenURL, strings.NewReader(form.Encode()))
	// req, err := http.NewRequest("POST", "http://localhost:11112/goFromLocalhost", strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.Wrapf(err, "GetAccessToken: Error generating the token request url for clientId: %q.", s.registration.ClientID)
	}
	response, err := client.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "GetAccessToken: Error executing the form POST for clientId: %q.", s.registration.ClientID)
	}
	log.Printf("Access token response status: %s", response.Status)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// log.Printf("bad response code, response: %+v", response)
		return "", fmt.Errorf("getAccessToken: Error response from access token fetch (%q)", response.Status)
	}

	defer response.Body.Close()
	// log.Printf("returned headers: %v", response.Header)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrapf(err, "GetAccessToken: Error reading body of access token fetch response for clientId: %q.", s.registration.ClientID)
	}
	// log.Printf("response Body: %q", string(body))

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", errors.Wrapf(err, "GetAccessToken: Failed to parse json from body of access token fetch response for clientId: %q.", s.registration.ClientID)
	}
	accessToken := data["access_token"].(string)
	// log.Printf("access token retrieved: %s", access_token)
	s.tokenMap[scopeStr] = accessToken
	return accessToken, nil
}

// DoServiceRequest fetches an auth token for a service call, then makes and returns the results of that call
func (s *ServiceConnector) DoServiceRequest(scopes []string, url, pMethod, body, pContentType, pAccept string) (*ServiceResult, error) {
	var (
		method      = "GET"
		contentType = "application/json"
		accept      = "application/json"
		req         *http.Request
	)
	if pMethod != "" {
		method = pMethod
	}
	if pContentType != "" {
		contentType = pContentType
	}
	if pAccept != "" {
		accept = pAccept
	}
	accessToken, err := s.getAccessToken(scopes)
	if err != nil {
		return nil, err
	}
	// log.Printf("access token fetched: %s", accessToken)
	client := &http.Client{Timeout: time.Second * 30}
	if method == "POST" {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			return nil, errors.Wrapf(err, "DoServiceReq: Error Creating new request for POST to %q", url)
		}
		req.Header.Add("Content-Type", contentType)
	} else { // GET
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "DoServiceReq: Error Creating new request for method: %q to %q", method, url)
		}
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Accept", accept)
	log.Printf("About to make request for url, request: %+v", req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "DoServiceReq: Error Executing new request for method: %q to %q", method, url)
	}

	log.Printf("Response received for method: %q to %q: %q", method, url, resp.Status)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "DoServiceReq: Error reading the response body for method: %q to %q", method, url)
	}

	return &ServiceResult{Header: resp.Header, Body: string(bodyBytes)}, nil
}
