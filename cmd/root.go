package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/frawleyskid/w3s-upload/cfg"
	"github.com/frawleyskid/w3s-upload/dir"
	"github.com/frawleyskid/w3s-upload/logger"
	"github.com/frawleyskid/w3s-upload/output"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/frawleyskid/w3s-upload/pin"
	"github.com/frawleyskid/w3s-upload/w3s"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrInvalidMode               = errors.New("invalid mode parameter")
	ErrOverloadedPinningServices = errors.New("can not upload to both Web3.Storage and a pinning service")
)

func init() {
	rootCmd.Flags().StringP("repo", "r", ".", "The `path` of the git repo containing the artifact files")
	rootCmd.Flags().StringP("mode", "m", string(cfg.DefaultMode), "The mode to operate in, either \"validate\" or \"history\"")
	rootCmd.Flags().String("path", cfg.DefaultPath, "The `path` of the artifact files in the repository")
	rootCmd.Flags().String("w3s-token", "", "The secret API `token` for Web3.Storage")
	rootCmd.Flags().String("ipfs-api", "", "The `multiaddr` of your IPFS node")
	rootCmd.Flags().String("pin-endpoint", "", "The `url` of the IPFS pinning service API endpoint to use")
	rootCmd.Flags().String("pin-token", "", "The secret bearer `token` for the configured IPFS pinning service")
	rootCmd.Flags().StringP("output", "o", "summary", "The output to produce, either \"artifacts\", \"cids\", or \"summary\"")
	rootCmd.Flags().Bool("action", false, "Run this tool as a GitHub Action")

	if err := rootCmd.Flags().MarkHidden("action"); err != nil {
		panic(err)
	}

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		panic(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "artifact-action",
	Long:  "Host content from Ace Archive on the IPFS network.\n\nTo upload content to Web3.Storage, you must specify `--ipfs-api` and\n`--w3s-token`.\n\nTo pin content with an IPFS pinning service, you must specify `--ipfs-api`,\n`--pin-endpoint`, and `--pin-token`.\n\nThe multiaddr of your local IPFS node is most likely\n`/dns/localhost/tcp/5001/http` by default.",
	Short: "Host content from Ace Archive on the IPFS network",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if cfg.PinningServicesOverloaded() {
			return ErrOverloadedPinningServices
		}

		var (
			artifacts []parse.Artifact
			err       error
		)

		switch mode := cfg.Mode(); mode {
		case cfg.ModeValidate:
			artifacts, err = parse.Tree(cfg.Repo(), cfg.Path())
			if err != nil {
				return err
			}
		case cfg.ModeHistory:
			artifacts, err = parse.History(cfg.Repo(), cfg.Path())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("%w: %s", ErrInvalidMode, mode)
		}

		cidsInArtifacts, err := parse.ExtractCids(artifacts)
		if err != nil {
			return err
		}

		actionOutput := output.Output{
			Artifacts: artifacts,
			RootCid:   nil,
		}

		if cfg.DoIpfs() {
			cidsToUpload := make([]cid.Cid, len(cidsInArtifacts), len(cidsInArtifacts)+1)
			copy(cidsToUpload, cidsInArtifacts)

			if cfg.Mode() == cfg.ModeValidate {
				rootCid, err := dir.Build(ctx, artifacts)
				if err != nil {
					return err
				}

				cidsToUpload = append(cidsToUpload, rootCid)

				rootCidStr := rootCid.String()
				actionOutput.RootCid = &rootCidStr
			}

			if cfg.DoW3S() {
				if err := w3s.Upload(ctx, cfg.W3SToken(), cidsToUpload); err != nil {
					return err
				}
			}

			if cfg.DoPin() {
				if err := pin.Pin(ctx, cfg.PinEndpoint(), cfg.PinToken(), cidsToUpload); err != nil {
					return err
				}
			}
		}

		return output.Print(actionOutput, cidsInArtifacts)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.LogError(err)
		os.Exit(1)
	}
}
