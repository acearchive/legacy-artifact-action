package cfg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var (
	ErrNotPinMode       = errors.New("these parameters are illegal when not pinning to IPFS")
	ErrMissingPinParams = errors.New("missing mandatory parameters for pinning to IPFS")
	ErrInvalidOutput    = errors.New("this is not a valid output type")
	ErrInvalidMode      = errors.New("this is not a valid mode")
)

type OperatingMode string

const (
	ModeValidate OperatingMode = "validate"
	ModeHistory  OperatingMode = "history"
	ModePin      OperatingMode = "pin"
)

var allModes = map[OperatingMode]struct{}{
	ModeValidate: {},
	ModeHistory:  {},
	ModePin:      {},
}

type OutputType string

const (
	OutputArtifacts OutputType = "artifacts"
	OutputCids      OutputType = "cids"
	OutputRoot      OutputType = "root"
	OutputSummary   OutputType = ""
)

var allOutputs = map[OutputType]struct{}{
	OutputArtifacts: {},
	OutputCids:      {},
	OutputRoot:      {},
	OutputSummary:   {},
}

const (
	DefaultMode = ModeValidate
	DefaultPath = "artifacts/"
)

func init() {
	viper.SetDefault("mode", string(DefaultMode))
	viper.SetDefault("path", string(DefaultPath))

	if err := viper.BindEnv("repo", "GITHUB_WORKSPACE"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("mode", "INPUT_MODE"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("path", "INPUT_PATH"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("ipfs-api", "INPUT_IPFS-API"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("pin-endpoint", "INPUT_PIN-ENDPOINT"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("pin-token", "INPUT_PIN-TOKEN"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("dry-run", "INPUT_DRY-RUN"); err != nil {
		panic(err)
	}
}

func Repo() string {
	return viper.GetString("repo")
}

func Mode() OperatingMode {
	return OperatingMode(viper.GetString("mode"))
}

func Path() string {
	return viper.GetString("path")
}

func IpfsAPI() string {
	return viper.GetString("ipfs-api")
}

func PinEndpoint() string {
	return viper.GetString("pin-endpoint")
}

func PinToken() string {
	return viper.GetString("pin-token")
}

func Output() OutputType {
	return OutputType(viper.GetString("output"))
}

func Action() bool {
	return viper.GetBool("action")
}

func DryRun() bool {
	return viper.GetBool("dry-run")
}

func StringifyInput(input string) string {
	if Action() {
		return fmt.Sprintf("`%s`", input)
	}

	return fmt.Sprintf("`--%s", input)
}

func ValidateParams() error {
	if _, isValid := allOutputs[Output()]; !isValid {
		return fmt.Errorf("%w: %s", ErrInvalidOutput, Output())
	}

	if _, isValid := allModes[Mode()]; !isValid {
		return fmt.Errorf("%w: %s", ErrInvalidMode, Mode())
	}

	hasIpfsAPI := viper.GetString("ipfs-api") != ""
	hasPinEndpoint := viper.GetString("pin-endpoint") != ""
	hasPinToken := viper.GetString("pin-token") != ""
	hasAnyPinParams := hasIpfsAPI || hasPinEndpoint || hasPinToken
	hasAllPinParams := hasIpfsAPI && hasPinEndpoint && hasPinToken

	if Mode() == ModePin && !hasAllPinParams {
		missingParams := make([]string, 0, 3)

		if !hasIpfsAPI {
			missingParams = append(missingParams, StringifyInput("ipfs-api"))
		}

		if !hasPinEndpoint {
			missingParams = append(missingParams, StringifyInput("pin-endpoint"))
		}

		if !hasPinToken {
			missingParams = append(missingParams, StringifyInput("pin-token"))
		}

		return fmt.Errorf("%w: %s", ErrMissingPinParams, strings.Join(missingParams, ", "))
	}

	if Mode() != ModePin && hasAnyPinParams {
		illegalParams := make([]string, 0, 3)

		if hasIpfsAPI {
			illegalParams = append(illegalParams, StringifyInput("ipfs-api"))
		}

		if hasPinEndpoint {
			illegalParams = append(illegalParams, StringifyInput("pin-endpoint"))
		}

		if hasPinToken {
			illegalParams = append(illegalParams, StringifyInput("pin-token"))
		}

		return fmt.Errorf("%w: %s", ErrNotPinMode, strings.Join(illegalParams, ", "))
	}

	return nil
}
