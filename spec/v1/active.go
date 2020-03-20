package v1

import "github.com/baetyl/baetyl-go/utils"

// ActiveRequest body of active request
type ActiveRequest struct {
	BatchName        string            `yaml:"batchName,omitempty" json:"batchName,omitempty"`
	Namespace        string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	FingerprintValue string            `yaml:"fingerprintValue,omitempty" json:"fingerprintValue,omitempty"`
	SecurityType     string            `yaml:"securityType,omitempty" json:"securityType,omitempty"`
	SecurityValue    string            `yaml:"securityValue,omitempty" json:"securityValue,omitempty"`
	PenetrateData    map[string]string `yaml:"penetrateData,omitempty" json:"penetrateData,omitempty"`
}

// ActiveResponse body of active responce
type ActiveResponse struct {
	NodeName    string            `yaml:"nodeName,omitempty" json:"nodeName,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Certificate utils.Certificate `yaml:"certificate,omitempty" json:"certificate,omitempty"`
}
