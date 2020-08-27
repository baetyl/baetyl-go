package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVar(t *testing.T) {
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
		Content: VariableValue{
			Value: dr,
		},
	}
	data, err := json.Marshal(v)
	assert.NoError(t, err)

	expData := "{\"kind\":\"report\",\"meta\":{\"1\":\"2\"},\"content\":{\"infos\":[{\"kind\":\"config\",\"name\":\"c082001\",\"version\":\"599944\"}]}}"
	assert.Equal(t, expData, string(data))

	msg := &Message{}
	err = json.Unmarshal(data, msg)
	assert.NoError(t, err)

	expContentData := "{\"infos\":[{\"kind\":\"config\",\"name\":\"c082001\",\"version\":\"599944\"}]}"
	assert.Equal(t, expContentData, string(msg.Content.data))

	assert.Nil(t, msg.Content.Value)
	msg.Content.Value = &DesireRequest{}
	if err := msg.Content.Unmarshal(msg.Content.Value); err != nil {
		assert.NoError(t, err)
	}
	assert.EqualValues(t, dr, msg.Content.Value.(*DesireRequest))
}
