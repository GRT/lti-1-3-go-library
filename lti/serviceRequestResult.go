package lti

import "net/http"

// ServiceResult is a holder object for the results of a service call
type ServiceResult struct {
	Header http.Header
	Body   string
}
