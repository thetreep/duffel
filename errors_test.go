package duffel

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ErrorSource is polymorphic ༼ ºل͟º ༽.
func TestErrorSourceUnmarshal(t *testing.T) {
	var obj ErrorSource
	assert.NoError(t, json.Unmarshal([]byte(`{"field":"amount","pointer":"/data"}`), &obj))
	assert.Equal(t, "amount", obj.Field)
	assert.Equal(t, "/data", obj.Pointer)
	assert.Empty(t, obj.Airline)

	var str ErrorSource
	assert.NoError(t, json.Unmarshal([]byte(`"easyjet"`), &str))
	assert.Equal(t, "easyjet", str.Airline)
	assert.Empty(t, str.Field)
}
