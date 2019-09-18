package ltiCache_test

import (
	"fmt"
	"net/http/httptest"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/gorilla/securecookie"
)

const (
	launchData        = "LD-7934604A-2A70-4FA3-8D17-64D373467D19"
	nonce             = "nonce-4ED53209-2AC8-452B-9420-D2F34EEA385C"
	sessionCookieName = "sessiondoo"
)

var (
	store sessions.Store
	cache ltiCache.Cache
)

func init() {
	store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
	cache = ltiCache.NewSessionStoreCache(store, sessionCookieName)
}

func TestPutFetchRequests(t *testing.T) {
	req1 := httptest.NewRequest("GET", "http://localhost", nil)
	rec1 := httptest.NewRecorder()

	session1, err := store.Get(req1, sessionCookieName)
	if err != nil {
		t.Fatalf("failed to create session from empty request: %v", err)
	}

	cache.PutLaunchData(req1, "launchKey", launchData)
	cache.PutNonce(req1, nonce)
	session1.Save(req1, rec1)

	req2 := httptest.NewRequest("GET", "http://localhost", nil)
	myCookie1 := rec1.Result().Cookies()[0]
	req2.AddCookie(myCookie1)
	fmt.Printf("cookie: %+v", myCookie1)

	gotLaunchData := cache.GetLaunchData(req2, "launchKey")
	if gotLaunchData != launchData {
		t.Fatalf("expecting launchData of %q but got: %q", launchData, gotLaunchData)
	}

	goodNonce := cache.CheckNonce(req2, nonce)
	if !goodNonce {
		t.Fatalf("The Nonce did not check out!")
	}
}
