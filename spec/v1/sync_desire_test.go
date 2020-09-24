package v1

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRDData(t *testing.T) {
	{
		// --- app
		desireddata := &ResourceValue{}
		desireddata.Name = "app"
		desireddata.Version = "123"
		desireddata.Kind = KindApplication
		desireddata.Value.Value = &Application{Name: "c"}

		appdata, err := json.Marshal(desireddata)
		assert.NoError(t, err)
		fmt.Printf(string(appdata))

		desireddata2 := &ResourceValue{}
		err = json.Unmarshal(appdata, desireddata2)
		assert.NoError(t, err)
		assert.Nil(t, desireddata2.Value.Value)

		// success
		app := desireddata2.App()
		assert.Equal(t, desireddata.Value.Value, app)

		desireddata.Kind = KindApp
		app = desireddata2.App()
		assert.Equal(t, desireddata.Value.Value, app)

		// failure
		cfg := desireddata2.Config()
		assert.Nil(t, cfg)

		// failure
		scr := desireddata2.Secret()
		assert.Nil(t, scr)
	}
	{
		// --- config
		desireddata := &ResourceValue{}
		desireddata.Name = "cfg"
		desireddata.Version = "123"
		desireddata.Kind = KindConfiguration
		desireddata.Value.Value = &Configuration{Name: "c"}

		appdata, err := json.Marshal(desireddata)
		assert.NoError(t, err)
		fmt.Printf(string(appdata))

		desireddata2 := &ResourceValue{}
		err = json.Unmarshal(appdata, desireddata2)
		assert.NoError(t, err)
		assert.Nil(t, desireddata2.Value.Value)

		// failure
		app := desireddata2.App()
		assert.Nil(t, app)
		assert.Nil(t, desireddata2.Value.Value)

		// sucees
		cfg := desireddata2.Config()
		assert.Equal(t, desireddata.Value.Value, cfg)

		desireddata.Kind = KindConfig
		cfg = desireddata2.Config()
		assert.Equal(t, desireddata.Value.Value, cfg)

		// failure
		scr := desireddata2.Secret()
		assert.Nil(t, scr)
	}
	{
		// --- secret
		desireddata := &ResourceValue{}
		desireddata.Name = "scr"
		desireddata.Version = "123"
		desireddata.Kind = KindSecret
		desireddata.Value.Value = &Secret{Name: "c"}

		appdata, err := json.Marshal(desireddata)
		assert.NoError(t, err)
		fmt.Printf(string(appdata))

		desireddata2 := &ResourceValue{}
		err = json.Unmarshal(appdata, desireddata2)
		assert.NoError(t, err)
		assert.Nil(t, desireddata2.Value.Value)

		// failure
		app := desireddata2.App()
		assert.Nil(t, app)
		assert.Nil(t, desireddata2.Value.Value)

		// failure
		cfg := desireddata2.Config()
		assert.Nil(t, cfg)
		assert.Nil(t, desireddata2.Value.Value)

		// failure
		scr := desireddata2.Secret()
		assert.Equal(t, desireddata.Value.Value, scr)
	}
}
