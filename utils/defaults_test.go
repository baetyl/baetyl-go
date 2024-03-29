package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testDefaultsModule struct {
	Name   string   `yaml:"name"`
	Params []string `yaml:"params" default:"[\"-c\", \"conf.yml\"]"`
}

type testDefaultsStruct struct {
	Others   string                        `yaml:"others"`
	Timeout  time.Duration                 `yaml:"timeout" default:"1m"`
	Modules  []testDefaultsModule          `yaml:"modules" default:"[]"`
	Services map[string]testDefaultsModule `yaml:"modules" default:"{}"`
}

func TestSetDefaults(t *testing.T) {
	err := SetDefaults("")
	assert.NotNil(t, err)
	assert.Equal(t, ": not a struct pointer", err.Error())

	tests := []struct {
		name    string
		args    *testDefaultsStruct
		want    *testDefaultsStruct
		wantErr bool
	}{
		{
			name: "defaults-struct-slice",
			args: &testDefaultsStruct{
				Others: "others",
				Modules: []testDefaultsModule{
					{
						Name: "m1",
					},
					{
						Name:   "m2",
						Params: []string{"arg1", "arg2"},
					},
				},
				Services: map[string]testDefaultsModule{
					"m1": {},
					"m2": {
						Params: []string{"arg1", "arg2"},
					},
				},
			},
			want: &testDefaultsStruct{
				Others:  "others",
				Timeout: time.Minute,
				Modules: []testDefaultsModule{
					{
						Name:   "m1",
						Params: []string{"-c", "conf.yml"},
					},
					{
						Name:   "m2",
						Params: []string{"arg1", "arg2"},
					},
				},
				Services: map[string]testDefaultsModule{
					"m1": {
						Params: []string{"-c", "conf.yml"},
					},
					"m2": {
						Params: []string{"arg1", "arg2"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetDefaults(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("SetDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.args)
		})
	}
}
