package omnitrail

import (
	"fmt"
	"github.com/edwarnicke/gitoid"
	"github.com/omnibor/omnibor-go"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type Envelope struct {
	Header  Header                   `json:"header"`
	Mapping map[string][]interface{} `json:"mapping"`
	options *options
}

type Header struct {
	Features []string `json:"features"`
}

type File struct {
	Type   string `json:"type"`
	Gitoid string `json:"gitoid"`
}

type Directory struct {
	Type   string `json:"type"`
	Gitoid string `json:"gitoid"`
}

type Option func(o *options)

type options struct {
	hashName string
}

func WithSha1() Option {
	return func(o *options) {
		o.hashName = "sha1"
	}
}

func WithSha256() Option {
	return func(o *options) {
		o.hashName = "sha256"
	}
}

func New(option ...Option) *Envelope {
	o := &options{
		hashName: "sha1",
	}
	for _, opt := range option {
		opt(o)
	}
	return &Envelope{
		Header:  Header{Features: []string{"directory", "file"}},
		Mapping: make(map[string][]interface{}),
		options: o,
	}
}

func (envelope *Envelope) Add(path string) error {
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
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
			defer reader.Close()

			options := make([]gitoid.Option, 0)
			if envelope.options.hashName == "sha256" {
				options = append(options, gitoid.WithSha256())
			}
			// TODO get the length of the file and pass it in as a gitoid option
			g, err := gitoid.New(reader, options...)
			if err != nil {
				return err
			}
			f := File{
				Type:   "file",
				Gitoid: g.String(),
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
		return len(keys[i]) < len(keys[j])
	})

	adgs := make(map[string]omnibor.ArtifactTree)
	// iterate over keys
	for _, k := range keys {
		// iterate over values
		for _, v := range envelope.Mapping[k] {
			// check if value is a directory
			if _, ok := v.(Directory); ok {
				switch envelope.options.hashName {
				case "sha1":
					adgs[k] = omnibor.NewSha1OmniBOR()
				case "sha256":
					adgs[k] = omnibor.NewSha256OmniBOR()
				default:
					return fmt.Errorf("unknown hash name: %s", envelope.options.hashName)
				}
			}
		}
	}

	// iterate over keys backwards, starting from the longest
	for i := len(keys) - 1; i >= 0; i-- {
		k := keys[i]
		// iterate over values
		for _, v := range envelope.Mapping[k] {
			// check if value is a file
			if _, ok := v.(File); ok {
				// get parent directory
				parentDir := filepath.Dir(k)
				// add file to parent directory

				err := func() error {
					reader, err := os.Open(path)
					if err != nil {
						return err
					}
					defer reader.Close()

					if k := adgs[parentDir]; k == nil {
						switch envelope.options.hashName {
						case "sha1":
							adgs[parentDir] = omnibor.NewSha1OmniBOR()
						case "sha256":
							adgs[parentDir] = omnibor.NewSha256OmniBOR()
						default:
							return fmt.Errorf("unknown hash name: %s", envelope.options.hashName)
						}
					}
					// get the length of the file at path
					adgs[parentDir].AddExistingReference(envelope.Mapping[k][0].(File).Gitoid)
					return nil
				}()
				if err != nil {
					return err
				}
			}
			if entry, ok := v.(Directory); ok {
				// check if the directory mapping has an omnibor object
				if exists := adgs[k]; exists == nil {
					// create new omnibor document and store it in the map
					switch envelope.options.hashName {
					case "sha1":
						adgs[k] = omnibor.NewSha1OmniBOR()
					case "sha256":
						adgs[k] = omnibor.NewSha256OmniBOR()
					default:
						return fmt.Errorf("unknown hash name: %s", envelope.options.hashName)
					}
				}

				// Generate the string representation of the current omnibor document
				entry.Gitoid = adgs[k].Identity()
				envelope.Mapping[k][0] = entry

				// get parent directory
				parentDir := filepath.Dir(k)
				// add identity to parent directory
				if k := adgs[parentDir]; k == nil {
					switch envelope.options.hashName {
					case "sha1":
						adgs[parentDir] = omnibor.NewSha1OmniBOR()
					case "sha256":
						adgs[parentDir] = omnibor.NewSha256OmniBOR()
					default:
						return fmt.Errorf("unknown hash name: %s", envelope.options.hashName)
					}
				}
				adgs[parentDir].AddExistingReference(entry.Gitoid)

				// print the identity and string representation of the current omnibor document
				fmt.Printf("%s\n%s\n", entry.Gitoid, adgs[k].String())
			}
		}
	}
	return nil
}
