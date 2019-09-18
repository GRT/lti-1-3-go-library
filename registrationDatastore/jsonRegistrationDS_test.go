package registrationDatastore_test

import (
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"
	"testing"
)

const (
	jsonPath     = "registrations.json"
	issuer       = "http://imsglobal.org"
	deploymentID = "1234"
)

var (
	regDS registrationDatastore.RegistrationDatastore
)

func init() {
	regDS, _ = registrationDatastore.NewJsonRegistrationDatastore(jsonPath)
}

func TestJsonParse(t *testing.T) {
	ds, err := registrationDatastore.NewJsonRegistrationDatastore(jsonPath)
	if err != nil {
		t.Fatalf("failed to create the json reg datastore: %v", err)
	}
	if ds == nil {
		t.Fatalf("json reg datastore should not be nil")
	}
	// fmt.Printf("DS: %+v", ds)

}

func TestFindRegYes(t *testing.T) {
	myReg, _ := regDS.FindRegistration(issuer)
	if myReg == nil {
		t.Fatalf("Could not find registration object for issuer (%q)", issuer)
	}
}

func TestFindRegNo(t *testing.T) {
	myReg, _ := regDS.FindRegistration("2F1D06B2-1859-4589-9CB2-6C04AABCB4C2")
	if myReg != nil {
		t.Fatalf("Found a registration when nil should have been returned")
	}
}

func TestFindDepYes(t *testing.T) {
	myDep, _ := regDS.FindDeployment(issuer, deploymentID)
	if myDep == nil {
		t.Fatalf("Could not find depoyment object for issuer (%q), deploymentId (%q)", issuer, deploymentID)
	}
}

func TestFindDepNo1(t *testing.T) {
	myDep, _ := regDS.FindDeployment(issuer, "__Doesnotexist")
	if myDep != nil {
		t.Fatalf("Found a Deployment when none should be found (missing deployment)")
	}
}

func TestFindDepNo2(t *testing.T) {
	myDep, _ := regDS.FindDeployment("__Doesnotexist", deploymentID)
	if myDep != nil {
		t.Fatalf("Found a Deployment when none should be found (missing issuer)")
	}
}
