package omnitrail

import (
	"github.com/edwarnicke/gitoid"
	"github.com/omnibor/omnibor-go"
	"github.com/testifysec/go-witness/log"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Envelope struct {
	Header     Header                   `json:"header"`
	Mapping    map[string][]interface{} `json:"mapping"`
	options    *options
	adgsSha1   map[string]omnibor.ArtifactTree
	adgsSha256 map[string]omnibor.ArtifactTree
}

type Header struct {
	Features []string `json:"features"`
}

type File struct {
	Type         string `json:"type"`
	Sha1Gitoid   string `json:"gitoid:sha1,omitempty"`
	Sha256Gitoid string `json:"gitoid:sha256,omitempty"`
}

type Directory struct {
	Type         string `json:"type"`
	Sha1Gitoid   string `json:"gitoid:sha1,omitempty"`
	Sha256Gitoid string `json:"gitoid:sha256,omitempty"`
}

type Option func(o *options)

type options struct {
	sha1Enabled   bool
	sha256Enabled bool
}

func WithSha1() Option {
	return func(o *options) {
		o.sha1Enabled = true
	}
}

func WithSha256() Option {
	return func(o *options) {
		o.sha256Enabled = true
	}
}

func New(option ...Option) *Envelope {
	o := &options{}
	for _, opt := range option {
		opt(o)
	}
	if o.sha1Enabled == false && o.sha256Enabled == false {
		o.sha1Enabled = true
	}
	return &Envelope{
		Header:     Header{Features: []string{"directory", "file"}},
		Mapping:    make(map[string][]interface{}),
		options:    o,
		adgsSha1:   make(map[string]omnibor.ArtifactTree),
		adgsSha256: make(map[string]omnibor.ArtifactTree),
	}
}

func (envelope *Envelope) Add(originalPath string) error {
	originalPath, err := filepath.Abs(originalPath)
	if err != nil {
		return err
	}

	// check if path already exists in the envelope, if so, return
	if _, ok := envelope.Mapping[originalPath]; ok {
		return nil
	}
	err = filepath.WalkDir(originalPath, func(path string, d fs.DirEntry, err error) error {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Creates a new Directory struct and adds it to the envelope
			directory := Directory{Type: "directory"}
			// TODO - check if path is already in the map to avoid duplicates
			envelope.Mapping[path] = append(envelope.Mapping[path], directory)
		} else {
			reader, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func(reader *os.File) {
				err := reader.Close()
				if err != nil {
					log.Error(err)
				}
			}(reader)

			// TODO get the length of the file and pass it in as a gitoid option
			g1, err := gitoid.New(reader)
			_, err = reader.Seek(0, 0)
			if err != nil {
				return err
			}
			g256, err := gitoid.New(reader, gitoid.WithSha256())
			if err != nil {
				return err
			}
			sha1Gitoid := g1.String()

			sha256Gitoid := g256.String()
			if envelope.options.sha1Enabled {
				sha1Gitoid = g1.String()
			}
			if envelope.options.sha256Enabled {
				sha1Gitoid = g1.String()
			}
			f := File{
				Type:         "file",
				Sha1Gitoid:   sha1Gitoid,
				Sha256Gitoid: sha256Gitoid,
			}
			// TODO - check if path is already in the map to avoid duplicates
			envelope.Mapping[path] = append(envelope.Mapping[path], f)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Generate GitOID for directories

	// collect all keys in the map
	var keys []string
	for k := range envelope.Mapping {
		keys = append(keys, k)
	}
	// sort keys by lexical order
	sort.Strings(keys)

	// stable sort keys by length
	sort.SliceStable(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	// iterate over keys
	for _, k := range keys {
		// iterate over values
		for _, v := range envelope.Mapping[k] {
			// check if value is a directory
			if _, ok := v.(Directory); ok {
				if envelope.options.sha1Enabled {
					envelope.adgsSha1[k] = omnibor.NewSha1OmniBOR()
				}
				if envelope.options.sha256Enabled {
					envelope.adgsSha256[k] = omnibor.NewSha256OmniBOR()
				}
			}
		}
	}

	// sort keys from longest to shortest
	sort.SliceStable(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	// iterate over keys backwards, starting from the longest
	//for i := len(keys) - 1; i >= 0; i-- {
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		// iterate over values
		for _, v := range envelope.Mapping[k] {
			// check if value is a file
			if _, ok := v.(File); ok {
				// get parent directory
				parentDir := filepath.Dir(k)
				// add file to parent directory

				err := func() error {

					// TODO abs might escape to outside of the current working directory
					// make sure we handle this. For now, we just check if it has the same
					// prefix.
					if !strings.HasPrefix(parentDir, originalPath) {
						return nil
					}

					if k := envelope.adgsSha1[parentDir]; k == nil {
						if envelope.options.sha1Enabled {
							envelope.adgsSha1[parentDir] = omnibor.NewSha1OmniBOR()
						}
					}
					if k := envelope.adgsSha256[parentDir]; k == nil {
						if envelope.options.sha256Enabled {
							envelope.adgsSha256[parentDir] = omnibor.NewSha256OmniBOR()
						}
					}

					// get the length of the file at path
					if envelope.options.sha1Enabled {
						if exists := envelope.adgsSha1[k]; exists == nil {
							// create new omnibor document and store it in the map
							if envelope.options.sha1Enabled {
								envelope.adgsSha1[k] = omnibor.NewSha1OmniBOR()
							}
						}
						err := envelope.adgsSha1[parentDir].AddExistingReference(envelope.Mapping[k][0].(File).Sha1Gitoid)
						if err != nil {
							return err
						}
					}
					if envelope.options.sha256Enabled {
						if exists := envelope.adgsSha256[k]; exists == nil {
							if envelope.options.sha256Enabled {
								envelope.adgsSha256[k] = omnibor.NewSha256OmniBOR()
							}
						}
						err := envelope.adgsSha256[parentDir].AddExistingReference(envelope.Mapping[k][0].(File).Sha256Gitoid)
						if err != nil {
							return err
						}
					}
					return nil
				}()
				if err != nil {
					return err
				}
			}
			if entry, ok := v.(Directory); ok {
				// check if the directory mapping has an omnibor object
				if exists := envelope.adgsSha1[k]; exists == nil {
					// create new omnibor document and store it in the map
					if envelope.options.sha1Enabled {
						envelope.adgsSha1[k] = omnibor.NewSha1OmniBOR()
					}
				}
				if exists := envelope.adgsSha256[k]; exists == nil {
					if envelope.options.sha256Enabled {
						envelope.adgsSha256[k] = omnibor.NewSha256OmniBOR()
					}
				}

				// Generate the string representation of the current omnibor document
				if envelope.options.sha1Enabled {
					entry.Sha1Gitoid = envelope.adgsSha1[k].Identity()
				}
				if envelope.options.sha256Enabled {
					entry.Sha256Gitoid = envelope.adgsSha256[k].Identity()
				}
				envelope.Mapping[k][0] = entry

				// get parent directory
				parentDir := filepath.Dir(k)
				if !strings.HasPrefix(parentDir, originalPath) {
					continue
				}
				// add identity to parent directory
				if k := envelope.adgsSha1[parentDir]; k == nil {
					if envelope.options.sha1Enabled {
						envelope.adgsSha1[parentDir] = omnibor.NewSha1OmniBOR()
					}
				}
				if k := envelope.adgsSha256[parentDir]; k == nil {
					if envelope.options.sha256Enabled {
						envelope.adgsSha256[parentDir] = omnibor.NewSha256OmniBOR()
					}
				}

				if envelope.options.sha1Enabled {
					err := envelope.adgsSha1[parentDir].AddExistingReference(entry.Sha1Gitoid)
					if err != nil {
						return err
					}
				}
				if envelope.options.sha256Enabled {
					err := envelope.adgsSha256[parentDir].AddExistingReference(entry.Sha256Gitoid)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Sha1ADGs return sha1 omnibor objects
func (envelope *Envelope) Sha1ADGs() map[string]string {
	adgs := make(map[string]string)
	for _, v := range envelope.adgsSha1 {
		if v.String() == "" {
			continue
		}
		adgs[v.Identity()] = v.String()
	}
	return adgs
}

// Sha256ADGs return sha256 omnibor objects
func (envelope *Envelope) Sha256ADGs() map[string]string {
	adgs := make(map[string]string)
	for _, v := range envelope.adgsSha256 {
		if v.String() == "" {
			continue
		}
		//fmt.Println(k)
		//fmt.Println(v.Identity())
		adgs[v.Identity()] = v.String()
		//fmt.Println("-")
	}
	return adgs
}
