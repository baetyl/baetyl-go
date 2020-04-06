package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShadowDiff(t *testing.T) {
	tests := []struct {
		name      string
		desire    Desire
		report    Report
		wantDelta Desire
		wantErr   error
	}{
		{
			name:      "0",
			desire:    Desire{},
			report:    Report{},
			wantDelta: Desire{},
		},
		{
			name:      "1",
			desire:    Desire{"name": "module", "version": "45"},
			report:    Report{"name": "module", "version": "43"},
			wantDelta: Desire{"version": "45"},
		},
		{
			name:      "2",
			desire:    Desire{"name": "module", "module": map[string]interface{}{"image": "test:v2"}},
			report:    Report{"name": "module", "module": map[string]interface{}{"image": "test:v1"}},
			wantDelta: Desire{"module": map[string]interface{}{"image": "test:v2"}},
		},
		{
			name:      "3",
			desire:    Desire{"module": map[string]interface{}{"image": "test:v2", "array": []interface{}{}}},
			report:    Report{"module": map[string]interface{}{"image": "test:v1", "object": map[string]interface{}{"attr": "value"}}},
			wantDelta: Desire{"module": map[string]interface{}{"image": "test:v2", "array": []interface{}{}}},
		},
		{
			name:      "6",
			desire:    Desire{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"n": nil, "5": map[string]interface{}{"6": "x"}}}}}},
			report:    Report{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"5": map[string]interface{}{"n": nil, "6": "y"}}}}}},
			wantDelta: Desire{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"5": map[string]interface{}{"6": "x"}}}}}},
		},
		{
			name:      "apps",
			desire:    Desire{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			report:    Report{"apps": []interface{}{map[string]interface{}{"name": "b", "version": "2"}, map[string]interface{}{"name": "c", "version": "2"}}},
			wantDelta: Desire{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDelta, err := tt.desire.Diff(tt.report)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantDelta, gotDelta)

			ais := gotDelta.AppInfos()
			if tt.name == "apps" {
				assert.Len(t, ais, 2)
				assert.Equal(t, "a", ais[0].Name)
				assert.Equal(t, "1", ais[0].Version)
				assert.Equal(t, "b", ais[1].Name)
				assert.Equal(t, "1", ais[1].Version)
			} else {
				assert.Nil(t, ais)
			}
		})
	}
}

func TestShadowMerge(t *testing.T) {
	tests := []struct {
		name     string
		oldData  map[string]interface{}
		newData  map[string]interface{}
		wantData map[string]interface{}
		wantErr  error
	}{
		{
			name:     "0",
			oldData:  map[string]interface{}{},
			newData:  map[string]interface{}{},
			wantData: map[string]interface{}{},
		},
		{
			name:     "1",
			oldData:  map[string]interface{}{"name": "module", "version": "45"},
			newData:  map[string]interface{}{"name": "module", "version": "43"},
			wantData: map[string]interface{}{"name": "module", "version": "43"},
		},
		{
			name:     "2",
			oldData:  map[string]interface{}{"name": "module", "module": map[string]interface{}{"image": "test:v2"}},
			newData:  map[string]interface{}{"name": "module", "module": map[string]interface{}{"image": "test:v1"}},
			wantData: map[string]interface{}{"name": "module", "module": map[string]interface{}{"image": "test:v1"}},
		},
		{
			name:     "3",
			oldData:  map[string]interface{}{"module": map[string]interface{}{"image": "test:v2", "array": []interface{}{}}},
			newData:  map[string]interface{}{"module": map[string]interface{}{"image": "test:v1", "object": map[string]interface{}{"attr": "value"}}},
			wantData: map[string]interface{}{"module": map[string]interface{}{"image": "test:v1", "array": []interface{}{}, "object": map[string]interface{}{"attr": "value"}}},
		},
		{
			name:     "6",
			oldData:  map[string]interface{}{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"5": map[string]interface{}{"6": "y"}}}}}},
			newData:  map[string]interface{}{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"n": nil, "5": map[string]interface{}{"n": nil, "6": "x"}}}}}},
			wantData: map[string]interface{}{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"5": map[string]interface{}{"6": "y"}}}}}},
			wantErr:  ErrJSONLevelExceedsLimit,
		},
		{
			name:     "apps",
			oldData:  map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			newData:  map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "b", "version": "2"}, map[string]interface{}{"name": "c", "version": "2"}}},
			wantData: map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "b", "version": "2"}, map[string]interface{}{"name": "c", "version": "2"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			or, nr := Report(tt.oldData), Report(tt.newData)
			err := or.Merge(nr)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, Report(tt.wantData), or)

			if tt.name == "6" {
				od, nd := Desire(tt.oldData), Desire(tt.newData)
				err = od.Merge(nd)
				assert.Equal(t, tt.wantErr, err)
				assert.Equal(t, Desire(tt.wantData), od)
			}

			ais := nr.AppInfos()
			if tt.name == "apps" {
				assert.Len(t, ais, 2)
				assert.Equal(t, "b", ais[0].Name)
				assert.Equal(t, "2", ais[0].Version)
				assert.Equal(t, "c", ais[1].Name)
				assert.Equal(t, "2", ais[1].Version)
			} else {
				assert.Nil(t, ais)
			}
		})
	}
}

func TestDesireSysAppInfos(t *testing.T) {
	sysApps := Desire{
		"sysapps": []interface{}{
			map[string]interface{}{"name": "app1", "version": "1"},
			map[string]interface{}{"name": "app2", "version": "2"},
		},
	}

	expectApps := []AppInfo{
		{
			Name:    "app1",
			Version: "1",
		},
		{
			Name:    "app2",
			Version: "2",
		},
	}

	assert.Equal(t, expectApps, sysApps.SysAppInfos())
}

func TestReportSysAppInfos(t *testing.T) {
	sysApps := Report{
		"sysapps": []interface{}{
			map[string]interface{}{"name": "app1", "version": "1"},
			map[string]interface{}{"name": "app2", "version": "2"},
		},
	}

	expectApps := []AppInfo{
		{
			Name:    "app1",
			Version: "1",
		},
		{
			Name:    "app2",
			Version: "2",
		},
	}

	assert.Equal(t, expectApps, sysApps.SysAppInfos())
}
