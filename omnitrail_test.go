package omnitrail

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	// TODO, use a tempdir instead of making one in ./test
	if _, err := os.Stat("./test/empty"); os.IsNotExist(err) {
		err := os.Mkdir("./test/empty", 0755)
		assert.NoError(t, err)
	}
	name := "empty"
	if err := testAdd(t, name); err != nil {
		t.Fatalf("TestEmpty failed: %v", err)
	}
}

func TestOneFiles(t *testing.T) {
	name := "one-file"
	if err := testAdd(t, name); err != nil {
		t.Fatalf("TestOneFiles failed: %v", err)
	}
}

func TestTwoFiles(t *testing.T) {
	name := "two-files"
	if err := testAdd(t, name); err != nil {
		t.Fatalf("TestTwoFiles failed: %v", err)
	}
}

func TestDeepStructure(t *testing.T) {
	name := "deep"
	if err := testAdd(t, name); err != nil {
		t.Fatalf("TestDeepStructure failed: %v", err)
	}
}

func TestSymlinkGood(t *testing.T) {
	name := "symlink-good"
	if err := testAdd(t, name); err != nil {
		t.Fatalf("TestSymlinkGood failed: %v", err)
	}
}

func TestSymlinkBroken(t *testing.T) {
	name := "symlink-broken"
	if err := testAdd(t, name); err != nil {
		t.Fatalf("should ignore a bad symlink: %v", err)
	}
}

func TestSymlinkOutOfBounds(t *testing.T) {
	name := "symlink-out-of-bounds"
	err := os.WriteFile("/tmp/omnitrail-well-known-file", []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("unable to write temporary file: %v", err)
	}
	defer os.Remove("/tmp/omnitrail-well-known-file")
	err = testAdd(t, name)
	if !strings.Contains(err.Error(), "not in the allow list") {
		t.Fatalf("unexpected error: %v", err)

	}
	if err == nil {
		t.Fatalf("TestSymlinkOutOfBounds failed: should report a symlik out of bounds")
	}
}

func testAdd(t *testing.T, name string) error {
	mapping := NewTrail()

	err := mapping.Add("./test/" + name)
	if err != nil {
		return err
	}

	// WARNING: these are only for generating new test cases easily
	// file, err := json.MarshalIndent(mapping.Envelope(), "", "  ")
	// os.WriteFile("./test/"+name+".json", file, 0644)
	// res := FormatADGString(mapping)
	// os.WriteFile("./test/"+name+".adg", []byte(res), 0644)
	// END WARNING

	expectedBytes, err := os.ReadFile("./test/" + name + ".json")
	if err != nil {
		return err
	}

	var expectedEnvelope Envelope
	err = json.Unmarshal(expectedBytes, &expectedEnvelope)
	if err != nil {
		return err
	}

	shortestExpectedKey := getShortestKey(&expectedEnvelope)
	shortestActualKey := getShortestKey(mapping.Envelope())

	for oldKey, val := range expectedEnvelope.Mapping {
		newKey := strings.Replace(oldKey, shortestExpectedKey, shortestActualKey, 1)
		delete(expectedEnvelope.Mapping, oldKey)
		expectedEnvelope.Mapping[newKey] = val
	}

	// get current username
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	uid := currentUser.Uid
	gid := currentUser.Gid

	for _, v := range expectedEnvelope.Mapping {
		v.Posix.OwnerUID = uid
		v.Posix.OwnerGID = gid
	}

	assert.Equal(t, &expectedEnvelope, mapping.Envelope())

	if !reflect.DeepEqual(&expectedEnvelope, mapping.Envelope()) {
		return fmt.Errorf("expected envelope does not match actual envelope")
	}

	res := FormatADGString(mapping)

	expectedBytes, err = os.ReadFile("./test/" + name + ".adg")
	if err != nil {
		return err
	}
	if string(expectedBytes) != res {
		return fmt.Errorf("expected ADG string does not match actual ADG string")
	}
	return nil
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
