package v1

import "time"

type STSRequest struct {
	STSType     string        `json:"stsType,omitempty" default:"minio"`
	ExpiredTime time.Duration `json:"expiredTime,omitempty" default:"12h"`
}

type STSResponse struct {
	AK         string    `json:"ak"`
	SK         string    `json:"sk"`
	Token      string    `json:"token"`
	Endpoint   string    `json:"endpoint"`
	Bucket     string    `json:"bucket"`
	Namespace  string    `json:"namespace"`
	NodeName   string    `json:"nodeName"`
	Expiration time.Time `json:"expiration"`
}
