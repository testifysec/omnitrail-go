package omnitrail

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmpty(t *testing.T) {
	mapping := New(WithSha1(), WithSha256())

	err := mapping.Add("./test/empty")
	assert.NoError(t, err)

	result, err := json.MarshalIndent(mapping, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(result))
	printADGs(mapping)
}

func TestOneFiles(t *testing.T) {
	mapping := New(WithSha1(), WithSha256())

	err := mapping.Add("./test/one-file")
	assert.NoError(t, err)

	result, err := json.MarshalIndent(mapping, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(result))

	fmt.Println("sha1")
	printADGs(mapping)
}

func printADGs(mapping *Envelope) {
	for k, v := range mapping.Sha1ADGs() {
		fmt.Println(k)
		fmt.Println(v)
		fmt.Println()
	}
	fmt.Println("sha256")
	for k, v := range mapping.Sha256ADGs() {
		fmt.Println(k)
		fmt.Println(v)
		fmt.Println("--")
	}
}

func TestTwoFiles(t *testing.T) {
	mapping := New(WithSha1(), WithSha256())

	err := mapping.Add("./test/two-files")
	assert.NoError(t, err)

	result, err := json.MarshalIndent(mapping, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(result))

	printADGs(mapping)
}

func TestDeepStructure(t *testing.T) {
	mapping := New(WithSha1(), WithSha256())

	err := mapping.Add("./test/deep")
	assert.NoError(t, err)

	result, err := json.MarshalIndent(mapping, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(result))
	printADGs(mapping)
}
