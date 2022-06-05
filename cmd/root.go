package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/frawleyskid/w3s-upload/logger"
	"github.com/frawleyskid/w3s-upload/output"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/frawleyskid/w3s-upload/pin"
	"github.com/frawleyskid/w3s-upload/w3s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	ErrInvalidMode               = errors.New("invalid mode parameter")
	ErrOverloadedPinningServices = errors.New("can not upload to both Web3.Storage and a pinning service")
)

type OperatingMode string

const (
	ModeValidate OperatingMode = "validate"
	ModeHistory  OperatingMode = "history"
)

func init() {
	rootCmd.Flags().StringP("repo", "r", ".", "The `path` of the git repo containing the artifact files")
	rootCmd.Flags().StringP("mode", "m", "validate", "The mode to operate in, either \"validate\" or \"history\"")
	rootCmd.Flags().String("path", "artifacts", "The path of the artifact files in the repository")
	rootCmd.Flags().String("w3s-token", "", "The secret API `token` for Web3.Storage")
	rootCmd.Flags().String("ipfs-api", "/dns/localhost/tcp/5001/http", "The `multiaddr` of your IPFS node")
	rootCmd.Flags().String("pin-endpoint", "", "The URL of the IPFS pinning service API `endpoint` to use")
	rootCmd.Flags().String("pin-token", "", "The bearer `token` for the configured IPFS pinning service")
	rootCmd.Flags().StringP("output", "o", "summary", "The output to produce, either \"artifacts\", \"cids\", or \"summary\"")
	rootCmd.Flags().Bool("action", false, "Run this tool as a GitHub Action")

	if err := rootCmd.Flags().MarkHidden("action"); err != nil {
		panic(err)
	}

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		panic(err)
	}

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

var rootCmd = &cobra.Command{
	Use:   "artifact-action",
	Long:  "Host content from Ace Archive on the IPFS network.\n\nTo upload content to Web3.Storage, you must specify `--w3s-token` and\n`--ipfs-api`.\n\nTo pin content with an IPFS pinning service, you must specify `--pin-endpoint`\nand `--pin-token`.",
	Short: "Host content from Ace Archive on the IPFS network",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if (viper.IsSet("w3s-token") || viper.IsSet("ipfs-api")) && (viper.IsSet("pin-endpoint") || viper.IsSet("pin-token")) {
			return ErrOverloadedPinningServices
		}

		var (
			artifacts []parse.Artifact
			err       error
		)

		switch mode := OperatingMode(viper.GetString("mode")); mode {
		case ModeValidate:
			artifacts, err = parse.Tree(viper.GetString("repo"), viper.GetString("path"))
			if err != nil {
				return err
			}
		case ModeHistory:
			artifacts, err = parse.History(viper.GetString("repo"), viper.GetString("path"))
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("%w: %s", ErrInvalidMode, mode)
		}

		cidList, err := parse.ExtractCids(artifacts)
		if err != nil {
			return err
		}

		if err := output.Print(artifacts, cidList); err != nil {
			return err
		}

		if viper.IsSet("w3s-token") && viper.IsSet("ipfs-api") {
			if err := w3s.Upload(ctx, viper.GetString("w3s-token"), viper.GetString("ipfs-api"), cidList); err != nil {
				return err
			}
		}

		if viper.IsSet("pin-endpoint") && viper.IsSet("pin-token") {
			if err := pin.Pin(ctx, viper.GetString("pin-endpoint"), viper.GetString("pin-token"), cidList); err != nil {
				return err
			}
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.LogError(err)
		os.Exit(1)
	}
}
