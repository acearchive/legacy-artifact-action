package logger

import (
	"fmt"
	"os"

	"github.com/acearchive/artifact-action/cfg"
)

func Printf(format string, a ...interface{}) {
	if cfg.Action() || cfg.Output() == cfg.OutputSummary {
		fmt.Printf(format, a...) //nolint:forbidigo
	}
}

func Println(a ...interface{}) {
	if cfg.Action() || cfg.Output() == cfg.OutputSummary {
		fmt.Println(a...) //nolint:forbidigo
	}
}

func LogError(err error) {
	if cfg.Action() {
		fmt.Printf("::error::%s\n", err.Error()) //nolint:forbidigo
	} else {
		if _, err := fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error()); err != nil {
			panic(err)
		}
	}
}

func LogErrorGroup(name string, errList []error) {
	if cfg.Action() {
		fmt.Printf("::group::%s\n", name) //nolint:forbidigo

		for _, err := range errList {
			fmt.Println(err.Error()) //nolint:forbidigo
		}

		fmt.Println("::endgroup::") //nolint:forbidigo
	} else {
		if _, err := fmt.Fprintln(os.Stderr, name); err != nil {
			panic(err)
		}

		for _, err := range errList {
			if _, err := fmt.Fprintln(os.Stderr, err.Error()); err != nil {
				panic(err)
			}
		}
	}
}

func LogNotice(msg string) {
	if cfg.Action() {
		fmt.Printf("::notice::%s\n", msg) //nolint:forbidigo
	} else {
		if _, err := fmt.Fprintf(os.Stderr, "\n%s\n\n", msg); err != nil {
			panic(err)
		}
	}
}

func Exit() {
	os.Exit(1)
}
