package omnitrail

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmpty(t *testing.T) {
}

func TestCurrent(t *testing.T) {
	mapping := New(WithSha256())

	err := mapping.Add("./test/two-files")
	assert.NoError(t, err)

	result, err := json.MarshalIndent(mapping, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(result))
}
