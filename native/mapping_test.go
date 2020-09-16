package native

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl-go/v2/log"
)

func TestServiceMapping_SetServicePorts(t *testing.T) {
	mapping := NewServiceMapping()
	assert.NotNil(t, mapping)

	defer os.RemoveAll("var")

	err := mapping.SetServicePorts("serviceA", []int{50010, 50011, 50012})
	assert.NoError(t, err)

	data, err := ioutil.ReadFile(ServiceMappingFile)
	assert.NoError(t, err)

	expected := `serviceA:
  ports:
  - 50010
  - 50011
  - 50012
`
	assert.Equal(t, expected, string(data))

	err = mapping.SetServicePorts("serviceB", []int{50020, 50021, 50022})
	assert.NoError(t, err)

	data, err = ioutil.ReadFile(ServiceMappingFile)
	assert.NoError(t, err)

	expected = `serviceA:
  ports:
  - 50010
  - 50011
  - 50012
serviceB:
  ports:
  - 50020
  - 50021
  - 50022
`
	assert.Equal(t, expected, string(data))

	err = mapping.SetServicePorts("serviceA", []int{50030, 50031, 50032})
	assert.NoError(t, err)

	data, err = ioutil.ReadFile(ServiceMappingFile)
	assert.NoError(t, err)

	expected = `serviceA:
  ports:
  - 50030
  - 50031
  - 50032
serviceB:
  ports:
  - 50020
  - 50021
  - 50022
`
	assert.Equal(t, expected, string(data))
}

func TestServiceMapping_WatchFile(t *testing.T) {
	mapping := NewServiceMapping()
	assert.NotNil(t, mapping)

	defer os.RemoveAll("var")

	err := mapping.SetServicePorts("serviceA", []int{50010, 50011, 50012})
	assert.NoError(t, err)

	config := `serviceA:
  ports:
  - 50010
  - 50011
serviceB:
  ports:
  - 50020
  - 50021
`

	err = ioutil.WriteFile(ServiceMappingFile, []byte(config), 0755)
	assert.NoError(t, err)

	logger := log.L()
	err = mapping.WatchFile(logger)
	assert.NoError(t, err)

	port, err := mapping.GetServiceNextPort("serviceA")
	assert.NoError(t, err)
	assert.Equal(t, 50010, port)

	port, err = mapping.GetServiceNextPort("serviceA")
	assert.NoError(t, err)
	assert.Equal(t, 50011, port)

	port, err = mapping.GetServiceNextPort("serviceA")
	assert.NoError(t, err)
	assert.Equal(t, 50010, port)

	port, err = mapping.GetServiceNextPort("serviceB")
	assert.NoError(t, err)
	assert.Equal(t, 50020, port)

	port, err = mapping.GetServiceNextPort("serviceB")
	assert.NoError(t, err)
	assert.Equal(t, 50021, port)

	port, err = mapping.GetServiceNextPort("serviceB")
	assert.NoError(t, err)
	assert.Equal(t, 50020, port)

	config = `serviceA:
 ports:
 - 51010
 - 51011
serviceB:
 ports:
 - 51020
 - 51021
serviceC:
 ports:
 - 51030
 - 51031
`
	err = ioutil.WriteFile(ServiceMappingFile, []byte(config), 0755)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	port, err = mapping.GetServiceNextPort("serviceA")
	assert.NoError(t, err)
	assert.Equal(t, 51010, port)

	port, err = mapping.GetServiceNextPort("serviceA")
	assert.NoError(t, err)
	assert.Equal(t, 51011, port)

	port, err = mapping.GetServiceNextPort("serviceA")
	assert.NoError(t, err)
	assert.Equal(t, 51010, port)

	port, err = mapping.GetServiceNextPort("serviceB")
	assert.NoError(t, err)
	assert.Equal(t, 51020, port)

	port, err = mapping.GetServiceNextPort("serviceB")
	assert.NoError(t, err)
	assert.Equal(t, 51021, port)

	port, err = mapping.GetServiceNextPort("serviceB")
	assert.NoError(t, err)
	assert.Equal(t, 51020, port)

	port, err = mapping.GetServiceNextPort("serviceC")
	assert.NoError(t, err)
	assert.Equal(t, 51030, port)

	port, err = mapping.GetServiceNextPort("serviceC")
	assert.NoError(t, err)
	assert.Equal(t, 51031, port)

	port, err = mapping.GetServiceNextPort("serviceC")
	assert.NoError(t, err)
	assert.Equal(t, 51030, port)

	err = mapping.DeleteServicePorts("serviceC")
	assert.NoError(t, err)

	data, err := ioutil.ReadFile(ServiceMappingFile)
	assert.NoError(t, err)

	expected := `serviceA:
  ports:
  - 51010
  - 51011
serviceB:
  ports:
  - 51020
  - 51021
`
	assert.Equal(t, expected, string(data))

	mapping.Close()
}
