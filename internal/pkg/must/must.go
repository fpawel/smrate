package must

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// FatalIf will call fmt.Print(err) and os.Exit(1) in case given err is not nil.
func FatalIf(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

// PanicIf will call panic(err) in case given err is not nil.
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

// MarshalYaml is a wrapper for toml.Marshal.
func MarshalYaml(v interface{}) []byte {
	data, err := yaml.Marshal(v)
	PanicIf(err)
	return data
}