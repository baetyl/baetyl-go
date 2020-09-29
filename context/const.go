package context

import (
	"fmt"
)

// EdgeNamespace return namespace of edge.
func EdgeNamespace() string {
	return baetylEdgeNamespace
}

// EdgeSystemNamespace return system namespace of edge.
func EdgeSystemNamespace() string {
	return baetylEdgeSystemNamespace
}

// BrokerPort return broker port.
func BrokerPort() string {
	return baetylBrokerSystemPort
}

// FunctionPort return http port of function.
func FunctionHttpPort() string {
	return baetylFunctionSystemHttpPort
}

// BrokerHost return broker host.
func BrokerHost() string {
	if RunMode() == RunModeNative {
		return "127.0.0.1"
	}
	return fmt.Sprintf("%s.%s", "baetyl-broker", baetylEdgeNamespace)
}

// FunctionHost return function host.
func FunctionHost() string {
	if RunMode() == RunModeNative {
		return "127.0.0.1"
	}
	return fmt.Sprintf("%s.%s", "baetyl-function", baetylEdgeSystemNamespace)
}
