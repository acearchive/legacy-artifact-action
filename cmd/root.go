package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/dir"
	"github.com/acearchive/artifact-action/logger"
	"github.com/acearchive/artifact-action/output"
	"github.com/acearchive/artifact-action/parse"
	"github.com/acearchive/artifact-action/pin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrInvalidMode = errors.New("invalid mode parameter")

func init() {
	rootCmd.Flags().StringP("repo", "r", ".", "The `path` of the git repo containing the artifact files")
	rootCmd.Flags().StringP("mode", "m", string(cfg.DefaultMode), "The mode to operate in")
	rootCmd.Flags().String("path", cfg.DefaultPath, "The `path` of the artifact files in the repository")
	rootCmd.Flags().String("ipfs-api", "", "The `multiaddr` of your IPFS node")
	rootCmd.Flags().String("pin-endpoint", "", "The `url` of the IPFS pinning service API endpoint to use")
	rootCmd.Flags().String("pin-token", "", "The secret bearer `token` for the configured IPFS pinning service")
	rootCmd.Flags().StringP("output", "o", "", "Print the given output type to stdout instead of summary statistics")
	rootCmd.Flags().Bool("dry-run", false, "Prevents uploading files when used in upload mode")
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
	Long:  "Host content from Ace Archive on the IPFS network.\n\nSee the README for details.",
	Short: "Host content from Ace Archive on the IPFS network",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if err := cfg.ValidateParams(); err != nil {
			return err
		}

		if cfg.DryRun() {
			if cfg.Mode() == cfg.ModePin {
				logger.LogNotice("This is a dry run. No files will actually be uploaded.")
			} else {
				logger.LogWarning(fmt.Sprintf("Using the %s option is pointless when not in `upload` mode.", cfg.StringifyInput("dry-run")))
			}
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
		case cfg.ModeHistory, cfg.ModePin:
			artifacts, err = parse.History(cfg.Repo(), cfg.Path())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("%w: %s", ErrInvalidMode, mode)
		}

		fileCids, err := parse.ExtractCids(artifacts)
		if err != nil {
			return err
		}

		actionOutput := output.Output{
			Artifacts: artifacts,
			RootCid:   nil,
		}

		if cfg.Mode() == cfg.ModePin {
			rootCid, err := dir.Build(ctx, artifacts)
			if err != nil {
				return err
			}

			rootCidStr := rootCid.String()
			actionOutput.RootCid = &rootCidStr

			if err := pin.Pin(ctx, cfg.PinEndpoint(), cfg.PinToken(), fileCids, rootCid); err != nil {
				return err
			}
		}

		if err := output.Print(actionOutput, fileCids); err != nil {
			return err
		}

		if cfg.DryRun() && cfg.Mode() == cfg.ModePin {
			logger.LogNotice("This was a dry run. No files were actually uploaded.")
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.LogError(err)
		logger.Exit()
	}
}
