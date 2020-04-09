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
			name:      "nil-1",
			desire:    Desire{},
			report:    nil,
			wantDelta: Desire{},
		},
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
		{
			name:      "apps-2",
			desire:    Desire{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			report:    Report{"apps": nil},
			wantDelta: Desire{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
		},
		{
			name:      "apps-3",
			desire:    Desire{"apps": nil},
			report:    Report{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			wantDelta: Desire{},
		},
		{
			name:      "apps-4",
			desire:    Desire{"apps": []interface{}{}},
			report:    Report{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			wantDelta: Desire{"apps": []interface{}{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDelta, err := tt.desire.Diff(tt.report)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantDelta, gotDelta)
			assert.Equal(t, tt.desire.AppInfos(), gotDelta.AppInfos())
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
			name:     "nil-1",
			oldData:  map[string]interface{}{},
			newData:  nil,
			wantData: map[string]interface{}{},
		},
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
			name:     "err",
			oldData:  map[string]interface{}{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"5": map[string]interface{}{"6": "y"}}}}}},
			newData:  map[string]interface{}{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"n": nil, "5": map[string]interface{}{"n": nil, "6": "x"}}}}}},
			wantData: map[string]interface{}{"1": map[string]interface{}{"2": map[string]interface{}{"3": map[string]interface{}{"4": map[string]interface{}{"5": map[string]interface{}{"6": "y"}}}}}},
			wantErr:  ErrJSONLevelExceedsLimit,
		},
		{
			name:     "apps-1",
			oldData:  map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			newData:  map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "b", "version": "2"}, map[string]interface{}{"name": "c", "version": "2"}}},
			wantData: map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "b", "version": "2"}, map[string]interface{}{"name": "c", "version": "2"}}},
		},
		{
			name:     "apps-2",
			oldData:  map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			newData:  map[string]interface{}{"apps": nil},
			wantData: map[string]interface{}{"apps": nil},
		},
		{
			name:     "apps-3",
			oldData:  map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			newData:  map[string]interface{}{"apps": []interface{}{}},
			wantData: map[string]interface{}{"apps": []interface{}{}},
		},
		{
			name:     "apps-4",
			oldData:  map[string]interface{}{"apps": nil},
			newData:  map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
			wantData: map[string]interface{}{"apps": []interface{}{map[string]interface{}{"name": "a", "version": "1"}, map[string]interface{}{"name": "b", "version": "1"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			or, nr := Report(tt.oldData), Report(tt.newData)
			err := or.Merge(nr)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, Report(tt.wantData), or)

			if tt.name == "err" {
				od, nd := Desire(tt.oldData), Desire(tt.newData)
				err = od.Merge(nd)
				assert.Equal(t, tt.wantErr, err)
				assert.Equal(t, Desire(tt.wantData), od)
			} else {
				assert.Equal(t, nr.AppInfos(), or.AppInfos())
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

func TestAppInfos(t *testing.T) {
	assert.Nil(t, Report{}.AppInfos())
	assert.Nil(t, Report{}.SysAppInfos())
	assert.Nil(t, Report{"apps": nil}.AppInfos())
	assert.Nil(t, Report{"sysapps": nil}.SysAppInfos())
	assert.Nil(t, Report{"apps": []string{}}.AppInfos())
	assert.Nil(t, Report{"sysapps": []string{}}.SysAppInfos())
	assert.Equal(t, []AppInfo{}, Report{"apps": []AppInfo{}}.AppInfos())
	assert.Equal(t, []AppInfo{}, Report{"sysapps": []AppInfo{}}.SysAppInfos())
	assert.Equal(t, []AppInfo{}, Report{"apps": []interface{}{}}.AppInfos())
	assert.Equal(t, []AppInfo{}, Report{"sysapps": []interface{}{}}.SysAppInfos())

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

	r := Report{
		"apps": []interface{}{
			map[string]interface{}{"name": "app1", "version": "1"},
			map[string]interface{}{"name": "app2", "version": "2"},
		},
	}
	assert.Equal(t, expectApps, r.AppInfos())

	r = Report{
		"sysapps": []AppInfo{
			AppInfo{Name: "app1", Version: "1"},
			AppInfo{Name: "app2", Version: "2"},
		},
	}

	assert.Equal(t, expectApps, r.SysAppInfos())
}
