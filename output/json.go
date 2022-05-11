package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/ipfs/go-cid"
	"github.com/spf13/viper"
)

const prettyJsonIndent = "  "

var ErrInvalidOutput = errors.New("invalid output parameter")

type Output struct {
	Artifacts []parse.Artifact `json:"artifacts"`
}

func marshalArtifact(artifacts []parse.Artifact, pretty bool) (string, error) {
	output := Output{Artifacts: artifacts}

	var (
		marshalledOutput []byte
		err              error
	)

	if pretty {
		marshalledOutput, err = json.MarshalIndent(output, "", prettyJsonIndent)
	} else {
		marshalledOutput, err = json.Marshal(output)
	}

	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}

func marshalCid(cids []cid.Cid, pretty bool) (string, error) {
	var (
		marshalledOutput []byte
		err              error
	)

	marshalledCids := make([]string, len(cids))

	for i, id := range cids {
		marshalledCids[i] = id.String()
	}

	if pretty {
		marshalledOutput, err = json.MarshalIndent(marshalledCids, "", prettyJsonIndent)
	} else {
		marshalledOutput, err = json.Marshal(marshalledCids)
	}

	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}

func Print(entries []parse.Artifact, cidList []cid.Cid) error {
	if viper.GetBool("action") {
		artifactOutput, err := marshalArtifact(entries, false)
		if err != nil {
			return err
		}

		fmt.Printf("::set-output name=artifacts::%s\n", artifactOutput)

		cidOutput, err := marshalCid(cidList, false)
		if err != nil {
			return err
		}

		fmt.Printf("::set-output name=cids::%s\n", cidOutput)

		return nil
	}

	switch outputMode := viper.GetString("output"); outputMode {
	case "artifacts":
		artifactOutput, err := marshalArtifact(entries, true)
		if err != nil {
			return err
		}

		fmt.Println(artifactOutput)
	case "cids":
		cidOutput, err := marshalCid(cidList, true)
		if err != nil {
			return err
		}

		fmt.Println(cidOutput)
	case "summary":
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOutput, outputMode)
	}

	return nil
}
