package mqtt

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	counter := NewCounter()

	assert.Equal(t, ID(1), counter.NextID())
	assert.Equal(t, ID(2), counter.NextID())

	for i := 0; i < math.MaxUint16-3; i++ {
		counter.NextID()
	}

	assert.Equal(t, ID(math.MaxUint16), counter.NextID())
	assert.Equal(t, ID(1), counter.NextID())

	counter.Reset()

	assert.Equal(t, ID(1), counter.NextID())
	assert.Equal(t, counter.GetNextID(), counter.NextID())
	assert.Equal(t, ID(3), counter.NextID())

	i := ID(1)
	assert.Equal(t, ID(2), NextCounterID(i))

	i = ID(math.MaxUint16)
	assert.Equal(t, ID(1), NextCounterID(i))
}
