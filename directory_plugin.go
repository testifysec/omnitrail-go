package omnitrail

import (
	"github.com/omnibor/omnibor-go"
	"os"
	"path/filepath"
	"sort"
)

type DirectoryPlugin struct {
	algorithms []string
	// algorithm -> path -> hash
	directories map[string]bool
	sha1adgs    map[string]omnibor.ArtifactTree
	sha256adgs  map[string]omnibor.ArtifactTree
}

func (plug *DirectoryPlugin) Sha1ADG(m map[string]string) {
	for _, v := range plug.sha1adgs {
		m[v.Identity()] = v.String()
	}
}

func (plug *DirectoryPlugin) Sha256ADG(m map[string]string) {
	for _, v := range plug.sha256adgs {
		m[v.Identity()] = v.String()
	}
}

func (plug *DirectoryPlugin) Add(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		plug.directories[path] = true
	}

	return nil
}

func (plug *DirectoryPlugin) Store(envelope *Envelope) error {
	envelope.Header.Features["directory"] = Feature{Algorithms: plug.algorithms}
	// get a list of all keys from plug.directories
	keys := make([]string, 0, len(plug.directories))
	for path := range plug.directories {
		keys = append(keys, path)
	}

	var sha1tree map[string]omnibor.ArtifactTree
	var sha256tree map[string]omnibor.ArtifactTree

	for _, algorithm := range plug.algorithms {
		switch algorithm {
		case "gitoid:sha1":
			sha1tree = make(map[string]omnibor.ArtifactTree)
		case "gitoid:sha256":
			sha256tree = make(map[string]omnibor.ArtifactTree)
		default:
			continue
		}
	}
	//sha1tree := make(map[string]omnibor.ArtifactTree)
	//sha256tree := make(map[string]omnibor.ArtifactTree)

	for _, key := range keys {
		if sha1tree != nil {
			sha1tree[key] = omnibor.NewSha1OmniBOR()
		}
		if sha256tree != nil {
			sha256tree[key] = omnibor.NewSha256OmniBOR()
		}
	}

	for path, element := range envelope.Mapping {
		dir := filepath.Dir(path)
		if _, ok := sha1tree[dir]; ok {
			err := sha1tree[dir].AddExistingReference(element.Sha1Gitoid)
			if err != nil {
				return err
			}
			err = sha256tree[dir].AddExistingReference(element.Sha256Gitoid)
			if err != nil {
				return err
			}
		}
	}

	// sort the keys from the longest length to the shortest length
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	if sha1tree != nil {
		err := plug.addKeysToTree(keys, sha1tree)
		if err != nil {
			return err
		}
	}
	if sha256tree != nil {
		err := plug.addKeysToTree(keys, sha256tree)
		if err != nil {
			return err
		}
	}

	for key, value := range sha1tree {
		if _, ok := envelope.Mapping[key]; !ok {
			envelope.Mapping[key] = &Element{
				Type: "directory",
			}
		}
		e := envelope.Mapping[key]
		e.Sha1Gitoid = value.Identity()
		envelope.Mapping[key] = e
	}

	for key, value := range sha256tree {
		if _, ok := envelope.Mapping[key]; !ok {
			envelope.Mapping[key] = &Element{
				Type: "directory",
			}
		}
		e := envelope.Mapping[key]
		e.Sha256Gitoid = value.Identity()
		envelope.Mapping[key] = e
	}

	for k, v := range sha1tree {
		plug.sha1adgs[k] = v
	}

	for k, v := range sha256tree {
		plug.sha256adgs[k] = v
	}

	return nil
}

func (plug *DirectoryPlugin) addKeysToTree(keys []string, tree map[string]omnibor.ArtifactTree) error {
	for _, key := range keys {
		dir := filepath.Dir(key)
		if _, ok := tree[dir]; ok {
			err := tree[dir].AddExistingReference(tree[key].Identity())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func NewDirectoryPlugin() Plugin {
	algorithms := []string{"gitoid:sha1", "gitoid:sha256"}
	sort.Strings(algorithms)
	return &DirectoryPlugin{
		algorithms:  algorithms,
		directories: make(map[string]bool),
		sha1adgs:    make(map[string]omnibor.ArtifactTree),
		sha256adgs:  make(map[string]omnibor.ArtifactTree),
	}
}
