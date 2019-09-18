package ltiCache

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

type sessionStoreCache struct {
	store sessions.Store
	// name of the session cookie, generally
	sessionName string
}

const (
	sessionKeyPrefix = "sessionStoreCache."
	nonceKey         = "nonce"
)

func NewSessionStoreCache(store sessions.Store, sessionName string) Cache {
	return &sessionStoreCache{store, sessionName}
}

func (c *sessionStoreCache) GetLaunchData(r *http.Request, key string) string {
	return c.fetchValueWithKey(r, key)
}

func (c *sessionStoreCache) PutLaunchData(r *http.Request, key, jwtBody string) {
	c.putValueWithKey(r, key, jwtBody)
}

func (c *sessionStoreCache) PutNonce(r *http.Request, nonce string) {
	c.putValueWithKey(r, nonceKey, nonce)
}

func (c *sessionStoreCache) CheckNonce(r *http.Request, nonce string) bool {
	cachedNonce := c.fetchValueWithKey(r, nonceKey)
	if cachedNonce == "" || cachedNonce != nonce {
		log.Printf("cached Nonce (%q) is not equal to request nonce (%q), reject.", cachedNonce, nonce)
		return false
	}
	return true
}

func (c *sessionStoreCache) fetchValueWithKey(r *http.Request, k string) string {
	key := fmt.Sprintf("%s%s", sessionKeyPrefix, k)
	// log.Printf("Looking for key %q in sessionName: %q", key, c.sessionName)
	session, err := c.store.Get(r, c.sessionName)
	if err != nil {
		log.Printf("fetch for session name %q failed since there is no session.  Assuming empty value for key %q.", c.sessionName, key)
	}
	// log.Printf("session values: %+v", session.Values)
	if v, ok := session.Values[key]; ok {
		return v.(string)
	}
	return ""
}

func (c *sessionStoreCache) putValueWithKey(r *http.Request, k string, v string) {
	key := fmt.Sprintf("%s%s", sessionKeyPrefix, k)
	// log.Printf("storing key %q in sessionName: %q", key, c.sessionName)

	session, _ := c.store.Get(r, c.sessionName)
	session.Values[key] = v
}
