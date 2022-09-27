package cfg

import "github.com/spf13/viper"

type OperatingMode string

const (
	ModeValidate OperatingMode = "validate"
	ModeHistory  OperatingMode = "history"
)

type OutputType string

const (
	OutputArtifacts OutputType = "artifacts"
	OutputCids      OutputType = "cids"
	OutputSummary   OutputType = "summary"
)

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

	if err := viper.BindEnv("w3s-token", "INPUT_W3S-TOKEN"); err != nil {
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

func W3SToken() string {
	return viper.GetString("w3s-token")
}

func IpfsApi() string {
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

func PinningServicesOverloaded() bool {
	return viper.IsSet("w3s-token") && (viper.IsSet("pin-endpoint") || viper.IsSet("pin-token"))
}

func DoIpfs() bool {
	return viper.IsSet("ipfs-api")
}

func DoPin() bool {
	return viper.IsSet("pin-endpoint") && viper.IsSet("pin-token")
}

func DoW3S() bool {
	return viper.IsSet("w3s-token")
}
