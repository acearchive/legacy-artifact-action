package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/parse"
)

const prettyJSONIndent = "  "

var ErrInvalidOutput = errors.New("invalid output parameter")

type Output struct {
	Artifacts []parse.Artifact `json:"artifacts"`
	RootCid   *string          `json:"rootCid"`
}

// initializeNilSlicesOfValue accepts a struct and initializes any nil slices in
// its fields to slices of len = 0 and cap = 0. If one of its fields is a
// struct or slice of structs, it does the same for those struct's fields
// recursively.
func initializeNilSlicesOfValue(value reflect.Value) {
	for fieldIndex := 0; fieldIndex < value.NumField(); fieldIndex++ {
		switch field := value.Field(fieldIndex); field.Kind() {
		case reflect.Struct:
			initializeNilSlicesOfValue(field)
		case reflect.Slice:
			if field.IsNil() {
				field.Set(reflect.MakeSlice(field.Type(), 0, 0))
			} else if field.Type().Elem().Kind() == reflect.Struct {
				for sliceIndex := 0; sliceIndex < field.Len(); sliceIndex++ {
					initializeNilSlicesOfValue(field.Index(sliceIndex))
				}
			}
		default:
			continue
		}
	}
}

func initializeNilSlices(value interface{}) {
	initializeNilSlicesOfValue(reflect.ValueOf(value).Elem())
}

func marshalArtifact(output Output, pretty bool) (string, error) {
	// To make the output more consistent and easier to parse, we want to
	// normalize any nil slices to empty slices before we serialize so that
	// they're serialized as `[]` and not `null`.
	initializeNilSlices(&output)

	var (
		marshalledOutput []byte
		err              error
	)

	if pretty {
		marshalledOutput, err = json.MarshalIndent(output, "", prettyJSONIndent)
	} else {
		marshalledOutput, err = json.Marshal(output)
	}

	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}

func marshalCid(cids parse.DeduplicatedCIDList, pretty bool) (string, error) {
	var (
		marshalledOutput []byte
		err              error
	)

	marshalledCids := make([]string, len(cids))

	for i, id := range cids {
		marshalledCids[i] = id.String()
	}

	if pretty {
		marshalledOutput, err = json.MarshalIndent(marshalledCids, "", prettyJSONIndent)
	} else {
		marshalledOutput, err = json.Marshal(marshalledCids)
	}

	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}

func Print(output Output, fileCids parse.DeduplicatedCIDList) error {
	if cfg.Action() {
		artifactOutput, err := marshalArtifact(output, false)
		if err != nil {
			return err
		}

		fmt.Printf("::set-output name=artifacts::%s\n", artifactOutput) //nolint:forbidigo

		cidOutput, err := marshalCid(fileCids, false)
		if err != nil {
			return err
		}

		fmt.Printf("::set-output name=cids::%s\n", cidOutput) //nolint:forbidigo

		if output.RootCid != nil {
			fmt.Printf("::set-output name=root::%s\n", *output.RootCid) //nolint:forbidigo
		}

		return nil
	}

	switch outputMode := cfg.Output(); outputMode {
	case cfg.OutputArtifacts:
		artifactOutput, err := marshalArtifact(output, true)
		if err != nil {
			return err
		}

		fmt.Println(artifactOutput) //nolint:forbidigo
	case cfg.OutputCids:
		cidOutput, err := marshalCid(fileCids, true)
		if err != nil {
			return err
		}

		fmt.Println(cidOutput) //nolint:forbidigo
	case cfg.OutputRoot:
		if output.RootCid != nil {
			fmt.Println(*output.RootCid) //nolint:forbidigo
		}
	case cfg.OutputSummary:
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOutput, outputMode)
	}

	return nil
}
