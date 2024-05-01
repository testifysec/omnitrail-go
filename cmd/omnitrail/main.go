package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fkautz/omnitrail-go"
	"log"
)

type OmniTrail struct {
	Sha1     map[string]string `json:"sha1"`
	Sha256   map[string]string `json:"sha256"`
	Envelope []byte            `json:"envelope"`
}

func main() {
	var (
		dataPath   string
		jsonOutput bool
	)
	flag.StringVar(&dataPath, "path", "", "Path to attest")
	flag.BoolVar(&jsonOutput, "json", false, "Enable JSON output")

	flag.Parse()

	if dataPath == "" {
		log.Fatal("Please specify a path to attest")
	}

	// Create a new trail.
	trail := omnitrail.NewTrail()
	// Add the input_path to the trail.
	err := trail.Add(dataPath)
	if err != nil {
		log.Fatalf("error adding %s: %v\n", dataPath, err)
	}

	if jsonOutput == true {
		jsonEnvelope, err := json.Marshal(trail.Envelope())
		if err != nil {
			log.Fatal(err)
		}
		output := OmniTrail{
			Sha1:     trail.Sha1ADGs(),
			Sha256:   trail.Sha256ADGs(),
			Envelope: jsonEnvelope,
		}
		marshal, err := json.Marshal(output)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(marshal))
	} else {
		jsonEnvelope, err := json.MarshalIndent(trail.Envelope(), "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		// Print the trail's omnibor ADG
		fmt.Println(omnitrail.FormatADGString(trail))

		// Print the trail's JSON representation
		fmt.Println(string(jsonEnvelope))
	}
}
