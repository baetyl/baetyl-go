package context

import (
	"fmt"
	"os"
	"runtime"

	"github.com/baetyl/baetyl-go/v2/errors"
)

// All keys
const (
	KeyBaetyl             = "BAETYL"
	KeyConfFile           = "BAETYL_CONF_FILE"
	KeyNodeName           = "BAETYL_NODE_NAME"
	KeyAppName            = "BAETYL_APP_NAME"
	KeyAppVersion         = "BAETYL_APP_VERSION"
	KeySvcName            = "BAETYL_SERVICE_NAME"
	KeySysConf            = "BAETYL_SYSTEM_CONF"
	KeyRunMode            = "BAETYL_RUN_MODE"
	KeyServiceDynamicPort = "BAETYL_SERVICE_DYNAMIC_PORT"
	KeyBaetylHostPathLib  = "BAETYL_HOST_PATH_LIB"
)

const (
	RunModeKube    = "kube"
	RunModeNative  = "native"
	RunModeAndroid = "android"
)

const (
	localHost                    = "127.0.0.1"
	baetylEdgeNamespace          = "baetyl-edge"
	baetylEdgeSystemNamespace    = "baetyl-edge-system"
	baetylCoreNativeSystemPort   = "443"
	baetylCoreKubeSystemPort     = "8443"
	baetylBrokerSystemPort       = "50010"
	baetylFunctionSystemHttpPort = "50011"
	baetylFunctionSystemGrpcPort = "50012"
	DefaultHostPathLib           = "/var/lib/baetyl"
	DefaultWindowsHostPathLib    = "C:/baetyl"
)

// HostPathLib return HostPathLib
func HostPathLib() (string, error) {
	var hostPathLib string
	if val := os.Getenv(KeyBaetylHostPathLib); val == "" {
		val = DefaultHostPathLib
		if runtime.GOOS == "windows" {
			val = DefaultWindowsHostPathLib
		}
		err := os.Setenv(KeyBaetylHostPathLib, val)
		if err != nil {
			return "", errors.Trace(err)
		}
		hostPathLib = val
	} else {
		hostPathLib = val
	}
	return hostPathLib, nil
}

// RunMode return run mode of edge.
func RunMode() string {
	mode := os.Getenv(KeyRunMode)
	if mode != RunModeNative {
		mode = RunModeKube
	}
	return mode
}

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

// FunctionHttpPort return http port of function.
func FunctionHttpPort() string {
	return baetylFunctionSystemHttpPort
}

func CoreHttpPort() string {
	if RunMode() == RunModeNative {
		return baetylCoreNativeSystemPort
	}
	return baetylCoreKubeSystemPort
}

// BrokerHost return broker host.
func BrokerHost() string {
	if RunMode() == RunModeNative {
		return localHost
	}
	return fmt.Sprintf("%s.%s", "baetyl-broker", baetylEdgeSystemNamespace)
}

// CoreHost return cpre host.
func CoreHost() string {
	if RunMode() == RunModeNative {
		return localHost
	}
	return fmt.Sprintf("%s.%s", "baetyl-core", baetylEdgeSystemNamespace)
}

// FunctionHost return function host.
func FunctionHost() string {
	if RunMode() == RunModeNative {
		return localHost
	}
	return fmt.Sprintf("%s.%s", "baetyl-function", baetylEdgeSystemNamespace)
}

func getBrokerAddress() string {
	return fmt.Sprintf("%s://%s:%s", "ssl", BrokerHost(), BrokerPort())
}

func getFunctionAddress() string {
	return fmt.Sprintf("%s://%s:%s", "https", FunctionHost(), FunctionHttpPort())
}

func getCoreAddress() string {
	return fmt.Sprintf("%s://%s:%s", "https", CoreHost(), CoreHttpPort())
}
