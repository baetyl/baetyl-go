package context

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHostPathLib(t *testing.T) {
	hostPathLib, err := HostPathLib()
	assert.NoError(t, err)
	assert.Equal(t, defaultHostPathLib, hostPathLib)
	os.Setenv(KeyBaetylHostPathLib, "/var/data")
	hostPathLib, err = HostPathLib()
	assert.NoError(t, err)
	assert.Equal(t, "/var/data", hostPathLib)
}

func TestDetectRunMode(t *testing.T) {
	os.Setenv(KeyRunMode, "native")
	assert.Equal(t, "native", RunMode())
	os.Setenv(KeyRunMode, "xxx")
	assert.Equal(t, "kube", RunMode())
}

func TestConst(t *testing.T) {
	assert.Equal(t, baetylEdgeNamespace, EdgeNamespace())
	assert.Equal(t, baetylEdgeSystemNamespace, EdgeSystemNamespace())
	assert.Equal(t, baetylBrokerSystemPort, BrokerPort())
	assert.Equal(t, baetylFunctionSystemHttpPort, FunctionHttpPort())
	assert.Equal(t, fmt.Sprintf("%s.%s", "baetyl-broker", baetylEdgeNamespace), BrokerHost())
	assert.Equal(t, fmt.Sprintf("%s.%s", "baetyl-function", baetylEdgeSystemNamespace), FunctionHost())
}
