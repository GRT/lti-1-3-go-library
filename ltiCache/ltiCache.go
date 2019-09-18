package ltiCache

import (
	"net/http"
)

type Cache interface {
	// GetLaunchData returns the MapClaims bytes, key is the LaunchID
	GetLaunchData(r *http.Request, launchID string) string
	PutLaunchData(r *http.Request, launchID, jwtBody string)
	PutNonce(r *http.Request, nonce string)
	CheckNonce(r *http.Request, nonce string) bool
}
