package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecV1_LazyValue(t *testing.T) {
	dr := &DesireRequest{
		Infos: []ResourceInfo{
			{
				Kind:    "config",
				Name:    "c082001",
				Version: "599944",
			},
		},
	}
	v := &Message{
		Kind:     MessageReport,
		Metadata: map[string]string{"1": "2"},
		Content: LazyValue{
			Value: dr,
		},
	}

	expData := "{\"kind\":\"report\",\"meta\":{\"1\":\"2\"},\"content\":{\"infos\":[{\"kind\":\"config\",\"name\":\"c082001\",\"version\":\"599944\"}]}}"

	desire := &DesireRequest{}
	err := v.Content.Unmarshal(desire)
	assert.NoError(t, err)
	assert.EqualValues(t, dr, desire)

	data, err := json.Marshal(v)
	assert.NoError(t, err)
	assert.Equal(t, expData, string(data))
	data, err = json.Marshal(v)
	assert.NoError(t, err)
	assert.Equal(t, expData, string(data))

	expContentData := "{\"infos\":[{\"kind\":\"config\",\"name\":\"c082001\",\"version\":\"599944\"}]}"

	msg := &Message{}
	err = json.Unmarshal(data, msg)
	assert.NoError(t, err)
	assert.Nil(t, msg.Content.Value)
	assert.Equal(t, expContentData, string(msg.Content.doc))
	err = json.Unmarshal(data, msg)
	assert.NoError(t, err)
	assert.Nil(t, msg.Content.Value)
	assert.Equal(t, expContentData, string(msg.Content.doc))

	desire1 := &DesireRequest{}
	err = msg.Content.Unmarshal(desire1)
	assert.NoError(t, err)
	assert.EqualValues(t, dr, desire1)
	desire2 := &DesireRequest{}
	err = msg.Content.Unmarshal(desire2)
	assert.NoError(t, err)
	assert.EqualValues(t, dr, desire2)

	data2, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Equal(t, expData, string(data2))
	msg2 := &Message{}
	err = json.Unmarshal(data2, msg2)
	assert.NoError(t, err)
	assert.Nil(t, msg2.Content.Value)
	assert.Equal(t, expContentData, string(msg2.Content.doc))

	msg3 := &Message{
		Kind:     MessageReport,
		Metadata: map[string]string{"1": "2"},
	}
	msg3.Content.SetJSON([]byte(expContentData))
	data4, err := json.Marshal(msg3)
	assert.NoError(t, err)
	assert.Equal(t, expData, string(data4))
}
