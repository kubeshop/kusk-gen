package prompt

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

type Prompter interface {
	SelectOneOf(label string, variants []string, withAdd bool) string
	Input(label, defaultString string) string
	InputNonEmpty(label, defaultString string) string
	InputMany(label string) []string
	FilePath(label, defaultPath string, shouldExist bool) string
	Confirm(question string) bool
}

type prompter struct{}

func New() Prompter {
	return prompter{}
}

func (pr prompter) SelectOneOf(label string, variants []string, withAdd bool) string {
	if len(variants) == 0 {
		// it's better to show a prompt
		return pr.InputNonEmpty(label, "")
	}

	if withAdd {
		p := promptui.SelectWithAdd{
			Label:  label,
			Stdout: os.Stderr,
			Items:  variants,
		}

		_, res, err := p.Run()
		if errors.Is(err, promptui.ErrInterrupt) {
			exit()
		}
		return res
	}

	p := promptui.Select{
		Label:  label,
		Stdout: os.Stderr,
		Items:  variants,
	}

	_, res, err := p.Run()
	if errors.Is(err, promptui.ErrInterrupt) {
		exit()
	}
	return res
}

func (_ prompter) Input(label, defaultString string) string {
	p := promptui.Prompt{
		Label:  label,
		Stdout: os.Stderr,
		Validate: func(s string) error {
			return nil
		},
		Default: defaultString,
	}

	res, err := p.Run()
	if errors.Is(err, promptui.ErrInterrupt) {
		exit()
	}

	return res
}

func (_ prompter) InputNonEmpty(label, defaultString string) string {
	p := promptui.Prompt{
		Label:  label,
		Stdout: os.Stderr,
		Validate: func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("should not be empty")
			}

			return nil
		},
		Default: defaultString,
	}

	res, err := p.Run()
	if errors.Is(err, promptui.ErrInterrupt) {
		exit()
	}

	return res
}

func (pr prompter) InputMany(label string) []string {
	endPromptInstruction := "(press return without entering anything to finish input)"
	completeInputLabel := fmt.Sprintf("%s %s", label, endPromptInstruction)

	var collection []string
	for {
		if in := pr.Input(completeInputLabel, ""); len(in) > 0 {
			collection = append(collection, in)
		} else {
			break
		}
	}

	return collection
}

func (_ prompter) FilePath(label, defaultPath string, shouldExist bool) string {
	p := promptui.Prompt{
		Label:   label,
		Stdout:  os.Stderr,
		Default: defaultPath,
		Validate: func(fp string) error {
			if strings.TrimSpace(fp) == "" {
				return errors.New("should not be empty")
			}

			if !shouldExist {
				return nil
			}

			if fileExists(fp) {
				return nil
			}

			return errors.New("should be an existing file")
		},
	}

	res, err := p.Run()
	if errors.Is(err, promptui.ErrInterrupt) {
		exit()
	}

	return res
}

func (_ prompter) Confirm(question string) bool {
	p := promptui.Prompt{
		Label:     question,
		Stdout:    os.Stderr,
		IsConfirm: true,
	}

	_, err := p.Run()
	if err != nil {
		switch err {
		case promptui.ErrAbort:
			return false
		case promptui.ErrInterrupt:
			exit()
		}
	}

	return true
}

func exit() {
	fmt.Println("Exiting.")
	os.Exit(1)
}

func fileExists(path string) bool {
	// check if file exists
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}
