package lti

import (
	"fmt"
	"log"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	gocache "github.com/patrickmn/go-cache"
)

const (
	cookieStatePrefix = "lti1_3_"
	// TODO: this should be from elsewhere?
	toolIssuer = "grt-go-test-platform"
)

var (
	// a cache for the public keys (by url)
	keysetCache *gocache.Cache
)

type ltiBase struct {
	regDS       registrationDatastore.RegistrationDatastore
	cache       ltiCache.Cache
	store       sessions.Store
	sessionName string
}

func init() {
	keysetCache = gocache.New(15*time.Minute, 60*time.Minute)
}

// fetches the public key used to validate the jwt token by finding the issuer in our registration datastore
func (lti ltiBase) getValidationKey(token *jwt.Token) (interface{}, error) {
	var keyset *jwk.Set
	const debug = true
	claims := token.Claims.(jwt.MapClaims)
	issuer := claims["iss"].(string)
	if issuer == "" {
		return nil, fmt.Errorf("The issuer cannot be blank")
	}
	reg, err := lti.regDS.FindRegistration(issuer)
	if err != nil {
		return nil, err
	}
	url := reg.KeySetURL
	log.Printf("pubkey url: %s", url)
	// fetch the keyset for this tool client
	if ks, found := keysetCache.Get(url); found {
		keyset = ks.(*jwk.Set)
		if debug {
			log.Printf("Cache HIT for key: %q\n", url)
		}
	} else {
		if debug {
			log.Printf("Cache MISS for key: %q\n", url)
		}
		keyset, err = jwk.Fetch(url)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed fetching keyset from endpoint: %q", url)
		}
		// save it to cache
		keysetCache.Set(url, keyset, gocache.DefaultExpiration)
	}
	// figure out which key in the keyset to use
	kid := token.Header["kid"].(string)
	if debug {
		log.Printf("Looking for token kid: %q", kid)
	}
	kset := keyset.LookupKeyID(kid)
	if kset == nil || len(kset) < 1 {
		return nil, fmt.Errorf("Token validation key not found for kid: %q", kid)
	} else if len(kset) > 1 {
		log.Printf("Multiple validation keys found for kid value: %q (using first one)", kid)
	}
	// TODO: get the the matching key by checking the algorithm - is that necessary?  Maybe not.
	// var key jwk.Key = kset[0]
	key := kset[0]
	// for _, k := range kset {
	// 	if k.Algorithm() == token.Header["alg"].(string)
	// }

	// get the key itself
	materializedKey, err := key.Materialize()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not materialize key for kid: %q", kid)
	}
	// log.Printf("Returning materialized Key: %+v\n", key)
	return materializedKey, nil
}
