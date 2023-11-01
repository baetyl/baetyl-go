package utils

import "github.com/denisbrodbeck/machineid"

const DefaultAppID = "hBmNlyAWkmrKfqwCFWoSiJiTsZJWksvv"

func GetFingerprint(appID string) (string, error) {
	if appID == "" {
		appID = DefaultAppID
	}
	return machineid.ProtectedID(appID)
}
