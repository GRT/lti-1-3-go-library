package lti

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	scopeKey         = "scope"
	scoreScopeKey    = "https://purl.imsglobal.org/spec/lti-ags/scope/score"
	lineItemScopeKey = "https://purl.imsglobal.org/spec/lti-ags/scope/lineitem"
	minScore         = 0
)

// AssignmentsGradeService offers the endpoints as specified in lti13
type AssignmentsGradeService struct {
	svcConn *ServiceConnector
	svcData *jwt.MapClaims
}

// LineItem represents a resource's item which can be assigned and graded
type LineItem struct {
	ID            string    `json:"id"`
	ScoreMax      int       `json:"scoreMaximum"`
	Label         string    `json:"label"`
	ResourceID    string    `json:"resourceid"`
	Tag           string    `json:"tag"`
	StartDateTime time.Time `json:"startDateTime"`
	EndDateTime   time.Time `json:"endDateTime"`
}

// Grade contains attributes for a grade
type Grade struct {
	ScoreGiven       int       `json:"scoreGiven"`
	ScoreMax         int       `json:"scoreMaximum"`
	ActivityProgress int       `json:"activityProgress"`
	GradingProgress  int       `json:"gradingProgress"`
	Timestamp        time.Time `json:"timestamp"`
	UserID           string    `json:"userId"`
}

// Result contains attributes about a particular users grade
type Result struct {
	UserID        string `json:"userId"`
	ResultScore   string `json:"resultScore"`
	ResultMaximum int    `json:"resultMaximum"`
	Comment       string `json:"comment"`
	ID            string `json:"id"`
	ScoreOf       string `json:"scoreOf"`
}

// NewAssignmentsGradeService creates a new AGS with (JWT) data from the claim
func NewAssignmentsGradeService(conn *ServiceConnector, data *jwt.MapClaims) *AssignmentsGradeService {
	return &AssignmentsGradeService{svcConn: conn, svcData: data}
}

// AgsPutGradeHandlerCreator returns a function which creates an http.Handler that uses a cached LTI Message launch's assessment grade service
//  to push a grade to the service.
// Expected method: Get, params: score, userId, launchId
func AgsPutGradeHandlerCreator(registrationDS registrationDatastore.RegistrationDatastore, cache ltiCache.Cache, store sessions.Store, sessionName string, debug bool, pLineItem *LineItem) func(http.Handler) http.Handler {
	lineitem := pLineItem
	if pLineItem == nil {
		lineitem = createDefaultLineItem()
	}
	return func(handla http.Handler) http.Handler {
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			userID := req.FormValue("userId")
			if userID == "" {
				http.Error(w, "missing userId param", 400)
				return
			}
			score, err := strconv.Atoi(req.FormValue("score"))
			if err != nil || score < minScore || score > lineitem.ScoreMax {
				http.Error(w, fmt.Sprintf("score param must be present and between %d and %d", minScore, lineitem.ScoreMax), 400)
				return
			}

			msgLaunch, err := NewMessageLaunchFromCache(req.FormValue("launchId"), req, registrationDS, cache, store, sessionName, debug)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			svc, err := msgLaunch.GetAgs()
			if err != nil {
				// could be an error or maybe the launch context doesn't provide ags
				http.Error(w, err.Error(), 404)
				return
			}
			grade := Grade{ScoreGiven: score, ScoreMax: lineitem.ScoreMax, UserID: userID}
			res, err := svc.PutGrade(grade, lineitem)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			log.Printf("response from PutGrade to be serialized: %+v", res)
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

// AgsGetGradesHandlerCreator returns a function which creates an http.Handler that uses a cached LTI Message launch's assessment grade service
//  to fetch the grades for a given lineitem from the service.
// Expected method: Get, params: launchId
func AgsGetGradesHandlerCreator(registrationDS registrationDatastore.RegistrationDatastore, cache ltiCache.Cache, store sessions.Store, sessionName string, debug bool, pLineItem *LineItem) func(http.Handler) http.Handler {
	lineitem := pLineItem
	if pLineItem == nil {
		lineitem = createDefaultLineItem()
	}
	return func(handla http.Handler) http.Handler {
		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			msgLaunch, err := NewMessageLaunchFromCache(req.FormValue("launchId"), req, registrationDS, cache, store, sessionName, debug)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			svc, err := msgLaunch.GetAgs()
			if err != nil {
				// could be an error or maybe the launch context doesn't provide ags
				http.Error(w, err.Error(), 404)
				return
			}
			res, err := svc.GetGrades(lineitem)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			log.Printf("get grades response to be serialized: %+v", res)
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

// ----------------------------------------------------------------------------
// Instance Public

// PutGrade is an lti1.3 specified AGS call.  It records the given grade against the given line item.
// If the line item is nil, then a default one is created.
func (s *AssignmentsGradeService) PutGrade(grade Grade, pLineItem *LineItem) (*Result, error) {
	log.Printf("PutGrade called, svcData: %+v", s.svcData)
	var scoreURL string
	inscope, err := s.hasScope(scoreScopeKey)
	if err != nil {
		return nil, errors.Wrap(err, "PutGrade failure due to inability to fetch scope")
	} else if !inscope {
		return nil, fmt.Errorf("missing scope: %q", scoreScopeKey)
	}

	if pLineItem != nil && pLineItem.ID == "" {
		lineitem, err := s.findOrCreateLineItem(pLineItem)
		if err != nil {
			return nil, errors.Wrap(err, "PutGrade failed to find or create lineitem from existing lineitem")
		}
		scoreURL = lineitem.ID
		log.Printf("Score url retrieved from existing lineitem (%+v), value: %q", pLineItem, scoreURL)
	} else if pLineItem == nil && (*s.svcData)["lineitem"] != nil {
		scoreURL = (*s.svcData)["lineitem"].(string)
		log.Printf("Score url retrieved from svcData, value: %q", scoreURL)
	} else {
		li := createDefaultLineItem()
		lineitem, err := s.findOrCreateLineItem(li)
		if err != nil {
			return nil, errors.Wrap(err, "PutGrade failed to find or create the lineitem from default")
		}
		scoreURL = lineitem.ID
		log.Printf("Score url retrieved from default lineitem (%+v), value: %q", li, scoreURL)
	}
	scoreURL = fmt.Sprintf("%s/scores", scoreURL)
	log.Printf("Final score url: %s", scoreURL)

	jsonBodyBytes, err := json.Marshal(grade)
	if err != nil {
		return nil, errors.Wrap(err, "PutGrade json failure")
	}

	res, err := s.svcConn.DoServiceRequest(s.getScopes(), scoreURL, "POST", string(jsonBodyBytes), "application/vnd.ims.lis.v1.score+json", "")
	if err != nil {
		return nil, errors.Wrap(err, "Failure executing service request for put grades")
	}
	log.Printf("put grades service request result: %+v", res)

	var retval *Result
	err = json.Unmarshal([]byte(res.Body), &retval)
	if err != nil {
		return nil, errors.Wrap(err, "put grade failed to create json from response")
	}
	return retval, nil
}

// GetGrades is an lti1.3 specified AGS call.  It returns the results of grades for the given line item.
// If the line item is nil, then a default one is assumed and created, if necessary.
func (s *AssignmentsGradeService) GetGrades(pLineItem *LineItem) ([]Result, error) {
	lineitem, err := s.findOrCreateLineItem(pLineItem)
	if err != nil {
		return nil, errors.Wrap(err, "GetGrades failed to find or create lineitem")
	}
	resultURL := fmt.Sprintf("%s/results", lineitem.ID)
	log.Printf("get grades result url: %s", resultURL)

	res, err := s.svcConn.DoServiceRequest(s.getScopes(), resultURL, "GET", "", "", "application/vnd.ims.lis.v2.resultcontainer+json")
	if err != nil {
		return nil, errors.Wrap(err, "Failure executing service request for get grades")
	}
	log.Printf("get grades service request result: %+v", res)

	var resList []Result
	err = json.Unmarshal([]byte(res.Body), &resList)
	if err != nil {
		return nil, errors.Wrap(err, "get grades failed to create json from response")
	}

	return resList, nil
}

// ----------------------------------------------------------------------------
// Instance Private

func (s *AssignmentsGradeService) findOrCreateLineItem(pLineItem *LineItem) (*LineItem, error) {
	log.Printf("findOrCreateLineItem: %+v", pLineItem)
	inscope, err := s.hasScope(lineItemScopeKey)
	if err != nil {
		return nil, errors.Wrapf(err, "Find/Create Lineitem failure due to inability to fetch scope: %q", lineItemScopeKey)
	} else if !inscope {
		return nil, fmt.Errorf("missing scope: %q", lineItemScopeKey)
	}

	lineitemsURL := (*s.svcData)["lineitems"].(string)
	log.Printf("calling GET on lineitems url: %q", lineitemsURL)
	res, err := s.svcConn.DoServiceRequest(s.getScopes(), lineitemsURL, "", "", "", "application/vnd.ims.lis.v2.lineitemcontainer+json")
	if err != nil {
		return nil, errors.Wrap(err, "Failure fetching existing lineitems")
	}
	log.Printf("lineitems initial lookup result: %+v", res)

	// find lineitem in existing list from provider,
	// if it exists, return it (tag should equal pLineItem.Tag if it's the same)
	var existingLineitems []LineItem
	if err := json.Unmarshal([]byte(res.Body), &existingLineitems); err != nil {
		return nil, errors.Wrap(err, "Failed to process lineitems")
	}
	for _, li := range existingLineitems {
		if li.Tag == pLineItem.Tag {
			log.Printf("Found lineitem amongst existing, returning: %+v", li)
			return &li, nil
		}
	}

	// since we didn't find one, create it and return it
	bodyBytes, err := json.Marshal(pLineItem)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize lineitem for sending")
	}
	log.Printf("calling POST on lineitems url: %q with body: %q", lineitemsURL, string(bodyBytes))

	res, err = s.svcConn.DoServiceRequest(s.getScopes(), lineitemsURL, "POST", string(bodyBytes), "application/vnd.ims.lis.v2.lineitem+json", "application/vnd.ims.lis.v2.lineitem+json")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new lineitem (1)")
	}
	log.Printf("result from lineitem post: %+v", res)
	var newLineItem *LineItem
	if err := json.Unmarshal([]byte(res.Body), &newLineItem); err != nil {
		log.Printf("Error during unmarshall: %+v", err)
		return nil, errors.Wrap(err, "failed to create new lineitem (2)")
	}
	return newLineItem, nil
}

func (s *AssignmentsGradeService) hasScope(pScope string) (bool, error) {
	scopes := s.getScopes()
	if len(scopes) == 0 {
		return false, fmt.Errorf("missing scopes in AGS service data")
	}
	for _, val := range scopes {
		valStr := fmt.Sprint(val)
		if pScope == valStr {
			return true, nil
		}
	}
	return false, nil
}

func (s *AssignmentsGradeService) getScopes() []string {
	scopes, ok := (*s.svcData)[scopeKey].([]interface{})
	if !ok {
		return make([]string, 0)
	}
	return stringSlicify(scopes)
}

// ----------------------------------------------------------------------------
// Helpers

func stringSlicify(t []interface{}) []string {
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func createDefaultLineItem() *LineItem {
	return &LineItem{Tag: "default", Label: "Default", ScoreMax: 100}
}
