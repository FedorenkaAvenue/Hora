package tools

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ParseYamlFile[T interface{}](i T, filePath string) T {
	var pFile, err = os.ReadFile(filePath)

	if err != nil {
		panic(err)
	}

	e := yaml.Unmarshal(pFile, &i)

	if e != nil {
		panic(e)
	}

	return i
}
