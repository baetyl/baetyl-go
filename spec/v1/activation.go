package v1

type Activation struct {
	FingerprintValue string                 `json:"fingerprintValue,omitempty"`
	PenetrateData    map[string]interface{} `json:"penetrateData,omitempty"`
}