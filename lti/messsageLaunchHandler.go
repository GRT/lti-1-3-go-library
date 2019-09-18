package lti

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
)

// MessageLaunchHandlerCreator returns a function that creates http handler functions that handle the LTI 1.3 message launch.
// The function that the creator creates wraps a handler with middleware that grabs the JWT, validates it and
// checks that it is a valid LTI Message Launch request
func MessageLaunchHandlerCreator(registrationDS registrationDatastore.RegistrationDatastore, cache ltiCache.Cache, store sessions.Store, sessionName string, debug bool) func(http.Handler) http.Handler {
	base := ltiBase{registrationDS, cache, store, sessionName}
	return func(handla http.Handler) http.Handler {
		// initialize the jwt middleware
		opts := jwtmiddleware.Options{
			SigningMethod:       jwt.SigningMethodRS256,
			UserProperty:        userKeyName,
			Extractor:           FromAnyParameter("id_token"),
			Debug:               debug,
			ValidationKeyGetter: base.getValidationKey,
			ErrorHandler:        tokenMWErrorHandler,
		}
		jwtMW := jwtmiddleware.New(opts)
		// now create the inner handler
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// create this request handler's messageLaunch object
			msgL := NewMessageLaunch(registrationDS, cache, store, sessionName, debug)
			sess, _ := msgL.store.Get(req, msgL.ltiBase.sessionName)

			// get id_token claims (which was validated and placed in the context by jwt middleware)
			claims := GetClaims(req)

			// Note: token validity, security, expired handled by wrapper
			if err := msgL.validateState(req); err != nil {
				http.Error(w, err.Error(), 401)
				return
			}
			// validate the nonce
			tokNonce := claims["nonce"].(string)
			if err := msgL.validateNonce(req, tokNonce); err != nil {
				http.Error(w, err.Error(), 401)
				return
			}

			if err := msgL.validateClientID(GetClaims(req)); err != nil {
				http.Error(w, err.Error(), 401)
				return
			}

			if err := msgL.validateDeployment(GetClaims(req)); err != nil {
				http.Error(w, err.Error(), 401)
				return
			}

			if err := msgL.validateMessage(GetClaims(req)); err != nil {
				http.Error(w, err.Error(), 401)
				return
			}

			bytes, err := json.Marshal(claims)
			if err != nil {
				http.Error(w, "failed to cache claims", 500)
				return
			}
			// save the launch in our cache, for future use
			claimsStr := string(bytes)
			log.Printf("launchData length: %d", len(claimsStr))
			// log.Printf("launchData: %s", claimsStr)
			msgL.cache.PutLaunchData(req, msgL.launchID, string(bytes))
			// save the launchID in the request context
			req = requestWithLaunchIDContext(req, msgL.launchID)
			if err := sess.Save(req, w); err != nil {
				log.Printf("error while saving session: %v", err)
			}
			handla.ServeHTTP(w, req)
		})
		return jwtMW.Handler(handlerFunc)
	}
}

func tokenMWErrorHandler(w http.ResponseWriter, r *http.Request, err string) {
	http.Error(w, fmt.Sprintf("Token issue: %s", err), 401)
}

// GetClaims fetches the user's jwt claims from the context. It fetches it from the JWT that
// the jwt middleware stored in the request context
func GetClaims(req *http.Request) jwt.MapClaims {
	userToken := req.Context().Value(userKeyName)
	tok := userToken.(*jwt.Token)
	claims := tok.Claims.(jwt.MapClaims)
	return claims
}

// GetLaunchID fetches the launchID associated with this request.  It is stored in the request context if present.
func GetLaunchID(req *http.Request) string {
	if lid := req.Context().Value(launchIDKey); lid != nil {
		if launchID, ok := lid.(string); ok {
			return launchID
		}
	}
	return ""
}

// FromAnyParameter returns a TokenExtractor that fetches the jwt from the body of the post or the query param
func FromAnyParameter(param string) jwtmiddleware.TokenExtractor {
	return func(r *http.Request) (string, error) {
		return r.FormValue(param), nil
	}
}

// requestWithLaunchIDContext returns a request with a context that contains the launchID
func requestWithLaunchIDContext(r *http.Request, launchID string) *http.Request {
	return requestWithNewContextValue(r, launchIDKey, launchID)
}

// requestWithNewContextValue returns a new request identitcal to the input but with a context
//  that includes the given key-value mapping.
func requestWithNewContextValue(r *http.Request, key ltiContextKey, value interface{}) *http.Request {
	newRequest := r.WithContext(context.WithValue(r.Context(), key, value))
	return newRequest
}
