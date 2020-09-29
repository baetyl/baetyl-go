package context

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConst(t *testing.T) {
	assert.Equal(t, baetylEdgeNamespace, EdgeNamespace())
	assert.Equal(t, baetylEdgeSystemNamespace, EdgeSystemNamespace())
	assert.Equal(t, baetylBrokerSystemPort, BrokerPort())
	assert.Equal(t, baetylFunctionSystemHttpPort, FunctionHttpPort())
	assert.Equal(t, fmt.Sprintf("%s.%s", "baetyl-broker", baetylEdgeNamespace), BrokerHost())
	assert.Equal(t, fmt.Sprintf("%s.%s", "baetyl-function", baetylEdgeSystemNamespace), FunctionHost())
}
