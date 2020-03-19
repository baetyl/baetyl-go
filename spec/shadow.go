package spec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
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
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Report            Report            `json:"report,omitempty"`
	Desire            Desire            `json:"desire,omitempty"`
}

// Report report data
type Report map[string]interface{}

// Desire desire data
type Desire map[string]interface{}

// Delta delat data
type Delta map[string]interface{}

// Merge merge new reported data
func (r Report) Merge(reported Report) error {
	return merge(r, reported, 1, maxJSONLevel)
}

// Merge merge new reported data
func (d Desire) Merge(desired Desire) error {
	return merge(d, desired, 1, maxJSONLevel)
}

// Diff diff with reported data, return the delta
func (d Desire) Diff(reported Report) (Delta, error) {
	return diff(d, reported)
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
