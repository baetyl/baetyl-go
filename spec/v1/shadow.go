package v1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/evanphx/json-patch"
)

// maxJSONLevel the max level of json
const maxJSONLevel = 5

// ErrJSONLevelExceedsLimit the level of json exceeds the max limit
var ErrJSONLevelExceedsLimit = fmt.Errorf("the level of json exceeds the max limit (%d)", maxJSONLevel)

// Shadow the spec of shadow
type Shadow struct {
	Namespace         string            `json:"namespace,omitempty"`
	Name              string            `json:"name,omitempty"`
	Version           string            `json:"version,omitempty"`
	CreationTimestamp time.Time         `json:"createTime,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Report            Report            `json:"report,omitempty"`
	Desire            Desire            `json:"desire,omitempty"`
	Description       string            `json:"description,omitempty"`
}

// Report report data
type Report map[string]interface{}

// Desire desire data
type Desire map[string]interface{}

// AppInfos return app infos
func (r Report) AppInfos() []AppInfo {
	return getAppInfos("apps", r)
}

// AppInfos return sysapps infos
func (r Report) SysAppInfos() []AppInfo {
	return getAppInfos("sysapps", r)
}

// Merge merge new reported data
func (r Report) Merge(reported Report) error {
	return merge(r, reported, 1, maxJSONLevel)
}

// AppInfos return app infos
func (d Desire) AppInfos() []AppInfo {
	return getAppInfos("apps", d)
}

// AppInfos return sysapps infos
func (d Desire) SysAppInfos() []AppInfo {
	return getAppInfos("sysapps", d)
}

// Merge merge new reported data
func (d Desire) Merge(desired Desire) error {
	return merge(d, desired, 1, maxJSONLevel)
}

// Diff diff with reported data, return the delta fo desire
func (d Desire) Diff(reported Report) (Desire, error) {
	return diff(d, reported)
}

func getAppInfos(appType string, data map[string]interface{}) []AppInfo {
	if data == nil || data[appType] == nil {
		return nil
	}
	res, ok := data[appType].([]AppInfo)
	if ok {
		return res
	}
	ais, ok := data[appType].([]interface{})
	if !ok || len(ais) == 0 {
		return nil
	}
	res = []AppInfo{}
	for _, ai := range ais {
		aim := ai.(map[string]interface{})
		if aim == nil {
			return nil
		}
		res = append(res, AppInfo{Name: aim["name"].(string), Version: aim["version"].(string)})
	}
	return res
}

// merge right map into left map
func merge(left, right map[string]interface{}, depth, maxDepth int) error {
	if depth >= maxDepth {
		return ErrJSONLevelExceedsLimit
	}
	for rk, rv := range right {
		lv, ok := left[rk]
		if !ok || reflect.TypeOf(rv).Kind() != reflect.Map || reflect.TypeOf(lv).Kind() != reflect.Map {
			left[rk] = rv
			continue
		}
		if err := merge(lv.(map[string]interface{}), rv.(map[string]interface{}), depth+1, maxDepth); err != nil {
			return err
		}
	}
	return nil
}

func diff(desired, reported map[string]interface{}) (map[string]interface{}, error) {
	var delta map[string]interface{}
	r, err := json.Marshal(reported)
	if err != nil {
		return delta, err
	}
	d, err := json.Marshal(desired)
	if err != nil {
		return delta, err
	}
	patch, err := jsonpatch.CreateMergePatch(r, d)
	if err != nil {
		return delta, err
	}
	err = json.Unmarshal(patch, &delta)
	if err != nil {
		return delta, err
	}
	clean(delta)
	return delta, nil
}

func clean(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
			continue
		}
		bk := reflect.TypeOf(v).Kind()
		if bk != reflect.Map {
			continue
		}
		if vm, ok := v.(map[string]interface{}); ok {
			clean(vm)
		}
	}
}
