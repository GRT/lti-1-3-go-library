package registrationDatastore

import (
	"encoding/json"
	"fmt"
	"os"
)

type jsonRegistrationDatastore struct {
	regMap map[string]Registration
}

func NewJsonRegistrationDatastore(jsonPath string) (RegistrationDatastore, error) {
	var regs []Registration
	file, err := os.Open(jsonPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	parser := json.NewDecoder(file)
	if err := parser.Decode(&regs); err != nil {
		return nil, err
	}
	regMap := convertRegsToRegMap(regs)
	return &jsonRegistrationDatastore{regMap: regMap}, nil
}

func (ds *jsonRegistrationDatastore) FindRegistration(issuer string) (*Registration, error) {
	if reg, exists := ds.findRegistration(issuer); exists {
		return &reg, nil
	}
	return nil, fmt.Errorf("Issuer not found")
}

func (ds *jsonRegistrationDatastore) FindDeployment(issuer, deploymentID string) (*Deployment, error) {
	if reg, exists := ds.findRegistration(issuer); exists {
		for _, id := range reg.DeploymentIds {
			if id == deploymentID {
				return &Deployment{id}, nil
			}
		}
	}
	return nil, nil
}

func (ds *jsonRegistrationDatastore) findRegistration(issuer string) (Registration, bool) {
	v, ok := ds.regMap[issuer]
	return v, ok
}

func convertRegsToRegMap(regs []Registration) map[string]Registration {
	m := make(map[string]Registration)
	for _, reg := range regs {
		m[reg.Issuer] = reg
	}
	return m
}
