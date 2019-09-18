package registrationDatastore

type Deployment struct {
	DeploymentID string `json:"deploymentId"`
}

type Registration struct {
	Issuer         string   `json:"issuer"`
	ClientID       string   `json:"clientId"`
	KeySetURL      string   `json:"keySetUrl"`
	AuthTokenURL   string   `json:"authTokenUrl"`
	AuthLoginURL   string   `json:"authLoginUrl"`
	ToolPrivateKey string   `json:"toolPrivateKey"`
	DeploymentIds  []string `json:"deploymentIds,omitempty"`
}

type RegistrationDatastore interface {
	FindRegistration(issuer string) (*Registration, error)
	FindDeployment(issuer, deploymentID string) (*Deployment, error)
}
