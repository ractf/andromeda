package challenge

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

type Spec struct {
	Name      string `json:"name"`
	Port      int    `json:"port"`
	MemLimit  int    `json:"mem_limit"`
	UserLimit int    `json:"user_limit"`
	Path      string
	ImageName string
}

var challengeSpecs = make(map[string]Spec)
var challengeSpecList = make([]Spec, 0)

func Create(path string) (Spec, error) {
	challengeSpec, err := unmarshallSpec(path)
	if err != nil {
		return Spec{}, err
	}

	challengeSpecs[challengeSpec.Name] = challengeSpec
	challengeSpecList = append(challengeSpecList, challengeSpec)

	return challengeSpec, nil
}
func unmarshallSpec(path string) (Spec, error) {
	jsonFile, err := os.Open(path + "/challenge.json")
	if err != nil {
		return Spec{}, err
	}

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return Spec{}, err
	}

	var challengeSpec Spec
	err = json.Unmarshal(bytes, &challengeSpec)
	_ = jsonFile.Close()
	if err != nil {
		return Spec{}, err
	}

	challengeSpec.Path = path
	challengeSpec.ImageName = "challenge-" + strings.ToLower(challengeSpec.Name)

	return challengeSpec, nil
}
