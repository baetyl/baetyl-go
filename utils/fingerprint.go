package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/denisbrodbeck/machineid"
	"github.com/super-l/machine-code/machine"
)

const DefaultAppID = "hBmNlyAWkmrKfqwCFWoSiJiTsZJWksvv"

func GetFingerprint(appID string) (string, error) {
	if appID == "" {
		appID = DefaultAppID
	}
	machineID, err := machineid.ProtectedID(appID)
	if err != nil {
		return "", err
	}
	uuid, err := machine.GetPlatformUUID()
	if err != nil {
		return "", err
	}
	macInfo, err := machine.GetMACAddress()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(machineID+uuid+macInfo))
	mac.Write([]byte(appID))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
