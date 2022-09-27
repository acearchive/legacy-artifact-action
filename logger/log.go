package logger

import (
	"fmt"
	"os"

	"github.com/frawleyskid/w3s-upload/cfg"
)

func Printf(format string, a ...interface{}) {
	if cfg.Action() || cfg.Output() == cfg.OutputSummary {
		fmt.Printf(format, a...)
	}
}

func Println(a ...interface{}) {
	if cfg.Action() || cfg.Output() == cfg.OutputSummary {
		fmt.Println(a...)
	}
}

func LogError(err error) {
	if cfg.Action() {
		fmt.Printf("::error::%s\n", err.Error())
	} else {
		if _, err := fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error()); err != nil {
			panic(err)
		}
	}
}

func LogErrorGroup(name string, errList []error) {
	if cfg.Action() {
		fmt.Printf("::group::%s\n", name)

		for _, err := range errList {
			fmt.Println(err.Error())
		}

		fmt.Println("::endgroup::")
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
