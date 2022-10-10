package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type testEncodeStruct struct {
	Others  string             `yaml:"others" json:"others"`
	Modules []testEncodeModule `yaml:"modules" json:"modules" default:"[]"`
}

type testEncodeModule struct {
	Name   string   `yaml:"name" json:"name"  validate:"regexp=^(m1|m2)$"`
	Params []string `yaml:"params" json:"params" default:"[\"-c\", \"conf.yml\"]"`
}

func TestUnmarshal(t *testing.T) {
	confString := `
id: id
name: name
others: others
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	cfg := testEncodeStruct{
		Others: "others",
		Modules: []testEncodeModule{
			{
				Name:   "m1",
				Params: []string{"-c", "conf.yml"},
			},
			{
				Name:   "m2",
				Params: []string{"arg1", "arg2"},
			},
		},
	}
	var cfg2 testEncodeStruct
	err := UnmarshalYAML([]byte(confString), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)

	err = UnmarshalYAML([]byte("-{}-"), &cfg2)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-{}-` into utils.testEncodeStruct")

	confString2 := `{
    "id": "id",
    "name": "name",
    "others": "others",
    "modules": [
        {
            "name": "k9"
        },
		{
            "name": "m2",
            "params": [
                "arg1",
                "arg2"
            ]
        }
    ]
}`
	err = UnmarshalYAML([]byte(confString2), &cfg2)
	assert.Error(t, err)
}

func TestParseEnv(t *testing.T) {
	const EnvHostIDKey = "BAETYL_HOST_ID"
	hostID := "test_host_id"
	err := os.Setenv(EnvHostIDKey, hostID)
	assert.NoError(t, err)
	confString := `
id: id
name: name
others: others
env: {{.BAETYL_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	expectedString := strings.Replace(confString, "{{.BAETYL_HOST_ID}}", hostID, 1)
	res, err := ParseEnv([]byte(confString))
	resString := string(res)
	assert.Equal(t, expectedString, resString)
	assert.NoError(t, err)

	// env not exist, env of parsed string would be empty
	confString2 := `
id: id
name: name
others: others
env: {{.BAETYL_NOT_EXIST}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	_, err2 := ParseEnv([]byte(confString2))
	assert.Error(t, err2)

	// syntax error
	confString3 := `
id: id
name: name
others: others
env: {{BAETYL_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	res3, err3 := ParseEnv([]byte(confString3))
	assert.Equal(t, []byte(nil), res3)
	assert.Error(t, err3)
}

func TestUnmarshalJSON(t *testing.T) {
	confString := `{
    "id": "id",
    "name": "name",
    "others": "others",
    "modules": [
        {
            "name": "m1"
        },
		{
            "name": "m2",
            "params": [
                "arg1",
                "arg2"
            ]
        }
    ]
}`
	cfg := testEncodeStruct{
		Others: "others",
		Modules: []testEncodeModule{
			{
				Name:   "m1",
				Params: []string{"-c", "conf.yml"},
			},
			{
				Name:   "m2",
				Params: []string{"arg1", "arg2"},
			},
		},
	}
	var cfg2 testEncodeStruct
	err := UnmarshalJSON([]byte(confString), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)

	err = UnmarshalJSON([]byte("{"), &cfg2)
	assert.Error(t, err)

	confString2 := `{
    "id": "id",
    "name": "name",
    "others": "others",
    "modules": [
        {
            "name": "k9"
        },
		{
            "name": "m2",
            "params": [
                "arg1",
                "arg2"
            ]
        }
    ]
}`
	err = UnmarshalJSON([]byte(confString2), &cfg2)
	assert.Error(t, err)
}

func TestLoadYAML(t *testing.T) {
	dir, err := ioutil.TempDir("", "template")
	assert.NoError(t, err)
	fileName := "template_test"
	f, err := os.Create(filepath.Join(dir, fileName))
	defer f.Close()
	confString := `
id: id
name: name
others: others
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	_, err = io.WriteString(f, confString)
	assert.NoError(t, err)

	cfg := testEncodeStruct{
		Others: "others",
		Modules: []testEncodeModule{
			{
				Name:   "m1",
				Params: []string{"-c", "conf.yml"},
			},
			{
				Name:   "m2",
				Params: []string{"arg1", "arg2"},
			},
		},
	}
	var cfg2 testEncodeStruct
	err = LoadYAML(filepath.Join(dir, fileName), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)

	fakeFileName := "fake"
	err = LoadYAML(filepath.Join(dir, fakeFileName), &cfg2)
	assert.Error(t, err)

	confString2 := `
id: id
name: name
others: others
env: {{BAETYL_HOST_ID}}
modules:
  - name: m1
  - name: m2
    params:
      - arg1
      - arg2
`
	fileName2 := "template_test2"
	f2, err := os.Create(filepath.Join(dir, fileName2))
	assert.NoError(t, err)
	_, err = io.WriteString(f2, confString2)
	err = LoadYAML(filepath.Join(dir, fileName2), &cfg2)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)
}

func TestUnmarshalYAML(t *testing.T) {
	confString := "max: 2"
	l := Length{1}
	unmarshal := func(ls interface{}) error {
		err := UnmarshalYAML([]byte(confString), ls)
		if err != nil {
			return err
		}
		return nil
	}
	err := l.UnmarshalYAML(unmarshal)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), l.Max)

	confString2 := "max: \n2"
	unmarshal2 := func(ls interface{}) error {
		err := UnmarshalYAML([]byte(confString2), ls)
		if err != nil {
			return err
		}
		return nil
	}
	err = l.UnmarshalYAML(unmarshal2)
	assert.Error(t, err)
}

func TestMarshalYAML(t *testing.T) {
	type fields struct {
		Max int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    interface{}
		wantErr bool
	}{
		{
			name:    "test1",
			fields:  fields{Max: 2},
			want:    length{"2"},
			wantErr: false,
		},
		{
			name:    "test2",
			fields:  fields{Max: 2 * 1024},
			want:    length{"2k"},
			wantErr: false,
		},
		{
			name:    "test3",
			fields:  fields{Max: 2 * 1024 * 1024},
			want:    length{"2m"},
			wantErr: false,
		},
		{
			name:    "test4",
			fields:  fields{Max: 2 * 1024 * 1024 * 1024},
			want:    length{"2g"},
			wantErr: false,
		},
		{
			name:    "test5",
			fields:  fields{Max: 2 * 1024 * 1024 * 1024 * 1024},
			want:    length{"2t"},
			wantErr: false,
		},
		{
			name:    "test6",
			fields:  fields{Max: 2 * 1024 * 1024 * 1024 * 1024 * 1024},
			want:    length{"2p"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Length{
				Max: tt.fields.Max,
			}
			got, err := l.MarshalYAML()
			if (err != nil) != tt.wantErr {
				t.Errorf("Length.MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Length.MarshalYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSizeMarshal(t *testing.T) {
	type dummy struct {
		S Size `yaml:"s" json:"s"`
	}

	tests := []struct {
		name     string
		dummy    *dummy
		wantJSON string
		wantYAML string
	}{
		{
			name:     "1",
			dummy:    &dummy{S: 1},
			wantJSON: "{\"s\":1}",
			wantYAML: "s: 1\n",
		},
		{
			name:     "1024",
			dummy:    &dummy{S: 1024},
			wantJSON: "{\"s\":1024}",
			wantYAML: "s: 1024\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dd, err := json.Marshal(tt.dummy)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantJSON, string(dd))

			ddd := &dummy{}
			err = json.Unmarshal(dd, ddd)
			assert.NoError(t, err)
			assert.Equal(t, tt.dummy.S, ddd.S)

			dd, err = yaml.Marshal(tt.dummy)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantYAML, string(dd))

			ddd = &dummy{}
			err = yaml.Unmarshal(dd, ddd)
			assert.NoError(t, err)
			assert.Equal(t, tt.dummy.S, ddd.S)
		})
	}
}

func TestSizeUnmarshal(t *testing.T) {
	type dummy struct {
		S Size `yaml:"s" json:"s"`
	}

	tests := []struct {
		name     string
		json     string
		yaml     string
		wantSize Size
		wantErr  bool
	}{
		{
			name:     "1k",
			json:     "{\"s\":\"1k\"}",
			yaml:     "s: 1k\n",
			wantSize: Size(1024),
		},
		{
			name:     "1m",
			json:     "{\"s\":\"1M\"}",
			yaml:     "s: \"1M\"\n",
			wantSize: Size(1024 * 1024),
		},
		{
			name:     "1g",
			json:     "{\"s\":\"1gB\"}",
			yaml:     "s: \"1gB\"\n",
			wantSize: Size(1024 * 1024 * 1024),
		},
		{
			name:    "1x",
			json:    "{\"s\":\"1x\"}",
			yaml:    "s: \"1x\"\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := new(dummy)
			err := json.Unmarshal([]byte(tt.json), d)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSize, d.S)
			}

			d = new(dummy)
			err = yaml.Unmarshal([]byte(tt.yaml), d)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSize, d.S)
			}
		})
	}
}
