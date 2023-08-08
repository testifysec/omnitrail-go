package omnitrail

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"sort"
	"strings"
	"testing"
)

func TestEmpty(t *testing.T) {
	// TODO, use a tempdir instead of making one in ./test
	if _, err := os.Stat("./test/empty"); os.IsNotExist(err) {
		err := os.Mkdir("./test/empty", 0755)
		assert.NoError(t, err)
	}
	name := "empty"
	testAdd(t, name)
}

func TestOneFiles(t *testing.T) {
	name := "one-file"
	testAdd(t, name)
}

func TestTwoFiles(t *testing.T) {
	name := "two-files"
	testAdd(t, name)
}

func TestDeepStructure(t *testing.T) {
	name := "deep"
	testAdd(t, name)
}

func testAdd(t *testing.T, name string) {
	mapping := NewTrail()

	err := mapping.Add("./test/" + name)
	assert.NoError(t, err)

	expectedBytes, err := os.ReadFile("./test/" + name + ".json")
	assert.NoError(t, err)

	var expectedEnvelope Envelope
	err = json.Unmarshal(expectedBytes, &expectedEnvelope)
	assert.NoError(t, err)

	shortestExpectedKey := getShortestKey(&expectedEnvelope)
	shortestActualKey := getShortestKey(mapping.Envelope())

	for oldKey, val := range expectedEnvelope.Mapping {
		newKey := strings.Replace(oldKey, shortestExpectedKey, shortestActualKey, 1)
		delete(expectedEnvelope.Mapping, oldKey)
		expectedEnvelope.Mapping[newKey] = val
	}

	// get current username
	currentUser, err := user.Current()
	assert.NoError(t, err)
	uid := currentUser.Uid
	gid := currentUser.Gid

	for _, v := range expectedEnvelope.Mapping {
		v.Posix.OwnerUID = uid
		v.Posix.OwnerGID = gid
	}

	assert.Equal(t, &expectedEnvelope, mapping.Envelope())

	res := FormatADGString(mapping)

	expectedBytes, err = os.ReadFile("./test/" + name + ".adg")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedBytes), res)
}

func getShortestKey(expectedEnvelope *Envelope) string {
	// get map keys
	keys := make([]string, 0, len(expectedEnvelope.Mapping))
	for key := range expectedEnvelope.Mapping {
		keys = append(keys, key)
	}

	// sort keys from shortest to longest
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) < len(keys[j])
	})

	shortestKey := keys[0]
	return shortestKey
}
