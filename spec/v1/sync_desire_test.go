package v1

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/baetyl/baetyl-go/spec/crd"
	"github.com/stretchr/testify/assert"
)

func TestCRDData(t *testing.T) {
	{
		// --- app
		crddata := &CRDData{}
		crddata.Name = "app"
		crddata.Version = "123"
		crddata.Kind = crd.KindApplication
		crddata.Value.Value = &crd.Application{Name: "c"}
		expected := "{\"name\":\"c\",\"creationTimestamp\":\"0001-01-01T00:00:00Z\"}"

		appdata, err := json.Marshal(crddata)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(crddata.Value.Data))
		fmt.Printf(string(appdata))

		crddata2 := &CRDData{}
		err = json.Unmarshal(appdata, crddata2)
		assert.NoError(t, err)
		assert.Nil(t, crddata2.Value.Value)
		assert.Equal(t, expected, string(crddata.Value.Data))

		// success
		app := crddata2.App()
		assert.Equal(t, crddata.Value.Value, app)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)

		crddata.Kind = crd.KindApp
		app = crddata2.App()
		assert.Equal(t, crddata.Value.Value, app)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)

		// failure
		cfg := crddata2.Config()
		assert.Nil(t, cfg)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)

		// failure
		scr := crddata2.Secret()
		assert.Nil(t, scr)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)
	}
	{
		// --- config
		crddata := &CRDData{}
		crddata.Name = "cfg"
		crddata.Version = "123"
		crddata.Kind = crd.KindConfiguration
		crddata.Value.Value = &crd.Configuration{Name: "c"}
		expected := "{\"name\":\"c\",\"creationTimestamp\":\"0001-01-01T00:00:00Z\",\"updateTimestamp\":\"0001-01-01T00:00:00Z\"}"

		appdata, err := json.Marshal(crddata)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(crddata.Value.Data))
		fmt.Printf(string(appdata))

		crddata2 := &CRDData{}
		err = json.Unmarshal(appdata, crddata2)
		assert.NoError(t, err)
		assert.Nil(t, crddata2.Value.Value)
		assert.Equal(t, expected, string(crddata.Value.Data))

		// failure
		app := crddata2.App()
		assert.Nil(t, app)
		assert.Nil(t, crddata2.Value.Value)

		// sucees
		cfg := crddata2.Config()
		assert.Equal(t, crddata.Value.Value, cfg)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)

		crddata.Kind = crd.KindConfig
		cfg = crddata2.Config()
		assert.Equal(t, crddata.Value.Value, cfg)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)

		// failure
		scr := crddata2.Secret()
		assert.Nil(t, scr)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)
	}
	{
		// --- secret
		crddata := &CRDData{}
		crddata.Name = "scr"
		crddata.Version = "123"
		crddata.Kind = crd.KindSecret
		crddata.Value.Value = &crd.Secret{Name: "c"}
		expected := "{\"name\":\"c\",\"creationTimestamp\":\"0001-01-01T00:00:00Z\",\"updateTimestamp\":\"0001-01-01T00:00:00Z\"}"

		appdata, err := json.Marshal(crddata)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(crddata.Value.Data))
		fmt.Printf(string(appdata))

		crddata2 := &CRDData{}
		err = json.Unmarshal(appdata, crddata2)
		assert.NoError(t, err)
		assert.Nil(t, crddata2.Value.Value)
		assert.Equal(t, expected, string(crddata.Value.Data))

		// failure
		app := crddata2.App()
		assert.Nil(t, app)
		assert.Nil(t, crddata2.Value.Value)

		// failure
		cfg := crddata2.Config()
		assert.Nil(t, cfg)
		assert.Nil(t, crddata2.Value.Value)

		// failure
		scr := crddata2.Secret()
		assert.Equal(t, crddata.Value.Value, scr)
		assert.Equal(t, crddata.Value.Value, crddata2.Value.Value)
	}
}
