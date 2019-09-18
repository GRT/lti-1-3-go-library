package lti

import (
	"fmt"
	"log"
	"net/http"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/pkg/errors"

	"github.com/gorilla/sessions"
)

// OidcLogin class that is setup to handle LTI 1.3 logins
type OidcLogin struct {
	ltiBase
	launchURL string
}

// NewOidcLogin creates a new OidcLogin with given args
func NewOidcLogin(registrationDS registrationDatastore.RegistrationDatastore, cache ltiCache.Cache, store sessions.Store, launchURL, sessionName string) *OidcLogin {
	base := ltiBase{registrationDS, cache, store, sessionName}
	return &OidcLogin{ltiBase: base, launchURL: launchURL}
}

// LoginRedirectHandler creates http handler that processes login requests from an LTI 1.3 tool provider / platform
// example params (post or get) utf8=%E2%9C%93&iss=http%3A%2F%2Fimsglobal.org&login_hint=29922&target_link_uri=http%3A%2F%2Flocalhost%3A9001%2Fexample%2Flaunch.php&lti_message_hint=701&commit=Post+request
// The response is a redirect back to the tool platform that will launch the tool (via lti.MessageLaunch)
func (O *OidcLogin) LoginRedirectHandler() http.Handler {
	handlerFunc := http.HandlerFunc(O.handleLoginRedirect)
	return handlerFunc
}

func (O *OidcLogin) handleLoginRedirect(w http.ResponseWriter, req *http.Request) {
	sess, _ := O.store.Get(req, O.sessionName)

	if O.launchURL == "" {
		http.Error(w, "launch url is not configured.", 400)
		return
	}

	reg, err := O.validateOidcLogin(req)
	if err != nil {
		http.Error(w, errors.Wrap(err, "Oidc Login validation failure").Error(), 400)
		return
	}

	state := fmt.Sprintf("state-%s", ksuid.New().String())
	setStateCookie(w, state)

	nonce := fmt.Sprintf("nonce-%s", ksuid.New().String())
	O.cache.PutNonce(req, nonce)

	redirReq, err := http.NewRequest("GET", reg.AuthLoginURL, nil)
	if err != nil {
		http.Error(w, errors.Wrap(err, "Failed to construct redirect").Error(), 500)
		return
	}
	q := redirReq.URL.Query()
	q.Add("scope", "openid")
	q.Add("response_type", "id_token")
	q.Add("response_mode", "form_post")
	q.Add("prompt", "none")
	q.Add("client_id", reg.ClientID)
	q.Add("redirect_uri", O.launchURL)
	q.Add("state", state)
	q.Add("nonce", nonce)
	q.Add("login_hint", req.FormValue("login_hint"))
	if mh := req.FormValue("lti_message_hint"); mh != "" {
		q.Add("lti_message_hint", mh)
	}
	redirReq.URL.RawQuery = q.Encode()
	redirURL := redirReq.URL.String()
	log.Printf("OIDC Login Redir: %s", redirURL)

	sess.Save(req, w)
	http.Redirect(w, req, redirURL, 302)
}

func (O *OidcLogin) validateOidcLogin(req *http.Request) (*registrationDatastore.Registration, error) {
	iss := req.FormValue("iss")
	// log.Printf("iss: %q", iss)
	if iss == "" {
		return nil, fmt.Errorf("issuer not found")
	}
	loginHint := req.FormValue("login_hint")
	// log.Printf("login_hint: %q", loginHint)
	if loginHint == "" {
		return nil, fmt.Errorf("login hint not found")
	}
	return O.regDS.FindRegistration(iss)
}

func setStateCookie(w http.ResponseWriter, state string) {
	secs := 3600
	cookie := http.Cookie{
		Name:    fmt.Sprintf("%s%s", cookieStatePrefix, state),
		Value:   state,
		Expires: time.Now().Add(time.Second * time.Duration(secs)),
		MaxAge:  secs,
	}
	http.SetCookie(w, &cookie)
}
