package omnitrail

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"github.com/edwarnicke/gitoid"
	"io"
	"os"
	"sort"
	"strings"
)

type FilePlugin struct {
	algorithms []string
	files      map[string]map[string]string
}

func (plug *FilePlugin) Sha1ADG(m map[string]string) {
	for algo, files := range plug.files {
		if algo == "gitoid:sha1" {
			for _, adg := range files {
				m[adg] = ""
			}
		}
	}
}

func (plug *FilePlugin) Sha256ADG(m map[string]string) {
	for algo, files := range plug.files {
		if algo == "gitoid:sha256" {
			for _, adg := range files {
				m[adg] = ""
			}
		}
	}
}

func NewFilePlugin() Plugin {
	algorithms := []string{"sha1", "sha256", "gitoid:sha1", "gitoid:sha256"}
	sort.Strings(algorithms)
	files := make(map[string]map[string]string)
	for _, algorithms := range algorithms {
		files[algorithms] = make(map[string]string)
	}
	return &FilePlugin{
		algorithms: algorithms,
		files:      files,
	}
}

func (plug *FilePlugin) Add(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	// explicitly ignore error from closing file
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	for _, algorithm := range plug.algorithms {
		_, err := f.Seek(0, 0)
		if err != nil {
			return err
		}

		if strings.HasPrefix(algorithm, "gitoid:") {
			var res *gitoid.GitOID
			switch algorithm {
			case "gitoid:sha1":
				res, err = gitoid.New(f, gitoid.WithContentLength(stat.Size()))
			case "gitoid:sha256":
				res, err = gitoid.New(f, gitoid.WithContentLength(stat.Size()), gitoid.WithSha256())
			}
			if err != nil {
				return err
			}
			plug.files[algorithm][path] = res.String()
		} else {
			switch algorithm {
			case "sha1":
				hasher := sha1.New()
				_, err = io.Copy(hasher, f)
				res := hasher.Sum([]byte{})
				plug.files[algorithm][path] = fmt.Sprintf("%x", res)
			case "sha256":
				hasher := sha256.New()
				_, err = io.Copy(hasher, f)
				res := hasher.Sum([]byte{})
				plug.files[algorithm][path] = fmt.Sprintf("%x", res)
			}
		}
	}

	return nil
}

func (plug *FilePlugin) Store(envelope *Envelope) error {
	envelope.Header.Features["file"] = Feature{Algorithms: plug.algorithms}
	for algorithm, paths := range plug.files {
		for path, hash := range paths {
			if _, ok := envelope.Mapping[path]; !ok {
				envelope.Mapping[path] = &Element{
					Type: fmt.Sprintf("%s", "file"),
				}
			}
			{
				e := envelope.Mapping[path]
				switch algorithm {
				case "sha1":
					e.Sha1 = hash
				case "sha256":
					e.Sha256 = hash
				case "gitoid:sha1":
					e.Sha1Gitoid = hash
				case "gitoid:sha256":
					e.Sha256Gitoid = hash
				}
				envelope.Mapping[path] = e
			}
		}
	}
	return nil
}
