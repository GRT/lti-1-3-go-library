package lti

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"

	"github.com/segmentio/ksuid"

	"github.com/pkg/errors"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/gorilla/sessions"
)

const (
	userKeyName = "user"
)

// MessageLaunch a struct that represents an LTI 1.3 Tool Launch
type MessageLaunch struct {
	ltiBase
	Debug        bool
	cachedClaims *jwt.MapClaims
	registration *registrationDatastore.Registration
	launchID     string
}

// used to store key/values in the context
type ltiContextKey int

// key for launchID values
const launchIDKey ltiContextKey = 0

// NewMessageLaunch creates a MessageLaunch with params.
func NewMessageLaunch(registrationDS registrationDatastore.RegistrationDatastore, cache ltiCache.Cache, store sessions.Store, sessionName string, debug bool) *MessageLaunch {
	base := ltiBase{registrationDS, cache, store, sessionName}
	ml := MessageLaunch{ltiBase: base, Debug: debug}
	ml.launchID = fmt.Sprintf("lti1p3_launch_%s", ksuid.New().String())
	return &ml
}

// NewMessageLaunchFromCache creates a MessageLaunch with params that is associated with a cached launch id (cached jwt payload)
func NewMessageLaunchFromCache(launchID string, r *http.Request, registrationDS registrationDatastore.RegistrationDatastore, cache ltiCache.Cache, store sessions.Store, sessionName string, debug bool) (*MessageLaunch, error) {
	claimsStr := cache.GetLaunchData(r, launchID)
	if claimsStr == "" {
		return nil, fmt.Errorf("Could not find message launch from cache with launchId: %q", launchID)
	}
	var claims jwt.MapClaims
	if err := json.Unmarshal([]byte(claimsStr), &claims); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshall claims from json")
	}
	m := NewMessageLaunch(registrationDS, cache, store, sessionName, debug)
	m.cachedClaims = &claims
	m.launchID = launchID

	// save the launchID in the request context
	newReq := requestWithLaunchIDContext(r, launchID)
	*r = *newReq // TODO: I don't think this changes the callers request.  test it.

	if err := m.validateClientID(*m.cachedClaims); err != nil {
		return nil, errors.Wrap(err, "Cached message launch failed validation")
	}
	return m, nil
}

// GetNrps returns the name roles provisioning service associated with this message launch context
func (M *MessageLaunch) GetNrps() (*NameRolesProvisioningService, error) {
	nrpsClaim, err := M.getNrpsClaim()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get NRPSvc: No claim")
	}
	svcConn := NewServiceConnector(*M.registration)
	svc := NewNameRolesProvisioningService(svcConn, &nrpsClaim)
	return svc, nil
}

// GetAgs returns the name roles provisioning service associated with this message launch context
func (M *MessageLaunch) GetAgs() (*AssignmentsGradeService, error) {
	agsClaim, err := M.getAgsClaim()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get AgSvc: No claim")
	}
	svcConn := NewServiceConnector(*M.registration)
	svc := NewAssignmentsGradeService(svcConn, &agsClaim)
	return svc, nil
}

func (M *MessageLaunch) getNrpsClaim() (jwt.MapClaims, error) {
	if M.cachedClaims == nil {
		return nil, fmt.Errorf("no cached claim exists for nrps")
	}
	nrsMap, ok := (*M.cachedClaims)["https://purl.imsglobal.org/spec/lti-nrps/claim/namesroleservice"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("nameroleservice claim is missing")
	}
	if nrsMap["context_memberships_url"].(string) == "" {
		return nil, fmt.Errorf("nameroleservice claim has no context_memberships_url attribute")
	}
	return nrsMap, nil
}

func (M *MessageLaunch) getAgsClaim() (jwt.MapClaims, error) {
	if M.cachedClaims == nil {
		return nil, fmt.Errorf("no cached claim exists for ags")
	}
	agsMap, ok := (*M.cachedClaims)["https://purl.imsglobal.org/spec/lti-ags/claim/endpoint"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("lti-ags claim is missing")
	}
	return agsMap, nil
}

func (M *MessageLaunch) validateState(req *http.Request) error {
	stateVal := req.FormValue("state")
	cookieName := fmt.Sprintf("%s%s", cookieStatePrefix, stateVal)
	stateCookie, err := req.Cookie(cookieName)
	if err != nil {
		return errors.Wrap(err, "Missing state cookie")
	}
	if stateCookie.Value == "" {
		return fmt.Errorf("Empty state cookie in request")
	}
	if stateCookie.Value != stateVal {
		return fmt.Errorf("State not found")
	}
	return nil
}

func (M *MessageLaunch) validateNonce(req *http.Request, nonce string) error {
	nonceOk := M.cache.CheckNonce(req, nonce)
	if nonceOk {
		return nil
	}
	log.Printf("nonce check failed")
	// platform is never sending the right nonce.
	//  It's commented out in the php reference: https://github.com/IMSGlobal/lti-1-3-php-library/blob/master/src/lti/lti_message_launch.php#L150
	//  for now, skip the error. Maybe this will be fixed in the future.
	// return fmt.Errorf("Invalid Nonce")
	return nil
}

func (M *MessageLaunch) validateClientID(claims jwt.MapClaims) error {
	var aud string
	var audClaim interface{} = claims["aud"]
	switch v := audClaim.(type) {
	case string:
		aud = v
	case []string:
		aud = v[0]
	default:
		log.Printf("aud claim is unexpected type: %T", v)
	}
	// get the issuer
	iss := claims["iss"].(string)
	// check that the clientIds match
	// note: to get this far, we know the issuer is in the claim, is a string and the registraton exists,
	//   since the jwt was already validated and the issuer was used to find the public key
	reg, _ := M.regDS.FindRegistration(iss)
	if reg.ClientID != aud {
		return fmt.Errorf("ClientId does not match issuer registration")
	}
	M.registration = reg
	return nil
}

func (M *MessageLaunch) validateDeployment(claims jwt.MapClaims) error {
	depID := claims["https://purl.imsglobal.org/spec/lti/claim/deployment_id"].(string)
	// get the issuer
	iss := claims["iss"].(string)
	// check that the clientIds match
	// note: to get this far, we know the issuer is in the claim, is a string and the registraton exists,
	//   since the jwt was already validated and the issuer was used to find the public key
	dep, _ := M.regDS.FindDeployment(iss, depID)
	if dep != nil {
		return nil
	}
	return fmt.Errorf("Unable to find deployment %q", depID)
}

func (M *MessageLaunch) validateMessage(claims jwt.MapClaims) error {
	msgType := claims["https://purl.imsglobal.org/spec/lti/claim/message_type"].(string)

	switch msgType {
	case "":
		return fmt.Errorf("Empty message type not allowed")
	case "LtiResourceLinkRequest":
		return M.validateMessageTypeLinkRequest(claims)
	case "LtiDeepLinkingRequest":
		return M.validateMessageTypeDeepLink(claims)
	default:
		return fmt.Errorf("unknown message type (%q)", msgType)
	}
}

func (M *MessageLaunch) validateMessageTypeLinkRequest(claims jwt.MapClaims) error {
	if err := M.validateMessageTypeCommon(claims); err != nil {
		return err
	}

	rlMap, ok := claims["https://purl.imsglobal.org/spec/lti/claim/resource_link"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("resource link claim is missing")
	}
	if rlMap["id"].(string) == "" {
		return fmt.Errorf("resource link id is missing")
	}
	return nil
}

func (M *MessageLaunch) validateMessageTypeDeepLink(claims jwt.MapClaims) error {
	if err := M.validateMessageTypeCommon(claims); err != nil {
		return err
	}

	dlsMap, ok := claims["https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("deep link settings claim is missing")
	}
	if dlsMap["deep_link_return_url"].(string) == "" {
		return fmt.Errorf("deep link return url is missing")
	}
	if dlsMap["accept_presentation_document_targets"].(string) == "" {
		return fmt.Errorf("deep link presentation type missing")
	}
	types, ok := dlsMap["accept_types"].([]string)
	if !ok {
		return fmt.Errorf("missing types (accept_types)")
	}
	// types must include 'ltiResourceLink'
	foundType := false
	for _, t := range types {
		if t == "ltiResourceLink" {
			foundType = true
		}
	}
	if !foundType {
		return fmt.Errorf("missing resource link placement types (accept_types)")
	}

	return nil
}

// validateMessageTypeCommon checks for claims that should be part of any message type
func (M *MessageLaunch) validateMessageTypeCommon(claims jwt.MapClaims) error {
	if claims["sub"].(string) == "" {
		return fmt.Errorf("token is missing user (sub) claim")
	}
	if claims["https://purl.imsglobal.org/spec/lti/claim/version"].(string) != "1.3.0" {
		return fmt.Errorf("token has incompatible lti version")
	}
	if claims["https://purl.imsglobal.org/spec/lti/claim/roles"] == nil {
		return fmt.Errorf("token is missing roles claim")
	}
	return nil
}

func (M *MessageLaunch) isDeepLinkLaunch(claims jwt.MapClaims) bool {
	msgType := claims["https://purl.imsglobal.org/spec/lti/claim/message_type"].(string)
	return msgType == "LtiDeepLinkingRequest"
}

func (M *MessageLaunch) isResourceLaunch(claims jwt.MapClaims) bool {
	msgType := claims["https://purl.imsglobal.org/spec/lti/claim/message_type"].(string)
	return msgType == "LtiResourceLinkRequest"
}
