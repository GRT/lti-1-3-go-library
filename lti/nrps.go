package lti

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

// NameRolesProvisioningService offers the endpoints as specified in lti13
type NameRolesProvisioningService struct {
	svcConn *ServiceConnector
	svcData *jwt.MapClaims
}

// NewNameRolesProvisioningService creates a new NameRolesProvisioningService
func NewNameRolesProvisioningService(conn *ServiceConnector, data *jwt.MapClaims) *NameRolesProvisioningService {
	return &NameRolesProvisioningService{svcConn: conn, svcData: data}
}

// NrpsMemberResponse hold the response from a getMember call
type NrpsMemberResponse struct {
	ID      string       `json:"id"`
	Context NrpsContext  `json:"context"`
	Members []NrpsMember `json:"members"`
}

// NrpsContext contains context info for getMembers call
type NrpsContext struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Title string `json:"title"`
}

// NrpsMember contains attributes for a single member
type NrpsMember struct {
	Name               string   `json:"name"`
	Picture            string   `json:"picture"`
	GivenName          string   `json:"given_name"`
	FamilyName         string   `json:"family_name"`
	MiddleName         string   `json:"middle_name"`
	Email              string   `json:"email"`
	UserID             string   `json:"user_id"`
	LisPersonSourcedid string   `json:"lis_person_sourcedid"`
	Roles              []string `json:"roles"`
}

// GetMembers uses the Message Launches context and auth token to return a list of users associated with this launch
func (s *NameRolesProvisioningService) GetMembers() (*NrpsMemberResponse, error) {
	svcURL := (*s.svcData)["context_memberships_url"].(string)
	svcScopes := []string{"https://purl.imsglobal.org/spec/lti-nrps/scope/contextmembership.readonly"}
	retval := &NrpsMemberResponse{}
	linkRegex := regexp.MustCompile("^?<(.*)>; ?rel=\"next\"$")
	count := 0
	for svcURL != "" {
		count++
		res, err := s.svcConn.DoServiceRequest(svcScopes, svcURL, "GET", "", "", "")
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to fetch member fetch #%d", count)
		}
		log.Printf("------ nrps (%s) iteration %d success, body len: %d --------------", svcURL, count, len(res.Body))
		log.Printf("  body: %+v", res.Body)
		resp := &NrpsMemberResponse{}
		if err := json.Unmarshal([]byte(res.Body), resp); err != nil {
			return nil, errors.Wrapf(err, "failed to parse json, fetch #%d", count)
		}
		if count == 1 {
			retval.Context = resp.Context
			retval.ID = resp.ID
			retval.Members = make([]NrpsMember, 5)
		}
		retval.Members = append(retval.Members, resp.Members...)
		svcURL = ""

	HeaderLoop:
		for k, v := range res.Header {
			log.Printf("header: %s: %v", k, v)
			hKey := strings.ToLower(k)
			if hKey == "link" {
				log.Printf("link header for url(%q) found: %s: %v", svcURL, k, v)
				for _, link := range v {
					if res := linkRegex.FindSubmatch([]byte(link)); res != nil {
						nextURL := string(res[1])
						log.Printf("Next Url determined: %v", nextURL)
						svcURL = nextURL
						break HeaderLoop
					}
				}
			}
		}
	}
	return retval, nil
}

// NrpsGetMemberHandlerCreator returns a function which creates an http.Handler that uses a cached LTI Message launch's name role provisioning service
//  to fetch a list of users from that context.
func NrpsGetMemberHandlerCreator(registrationDS registrationDatastore.RegistrationDatastore, cache ltiCache.Cache, store sessions.Store, sessionName string, debug bool) func(http.Handler) http.Handler {
	return func(handla http.Handler) http.Handler {
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			msgLaunch, err := NewMessageLaunchFromCache(req.FormValue("launchId"), req, registrationDS, cache, store, sessionName, debug)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			svc, err := msgLaunch.GetNrps()
			if err != nil {
				// could be an error or maybe the launch context doesn't provide nrps
				http.Error(w, err.Error(), 404)
				return
			}
			res, err := svc.GetMembers()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			// log.Printf("res to be serialized: %+v", res)
			// serialize the result
			b, err := json.Marshal(res)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
			// Invoke the passed in handler if it's there
			// Not much for it to do at this point
			if handla != nil {
				handla.ServeHTTP(w, req)
			}
		})
		return handlerFunc
	}
}
