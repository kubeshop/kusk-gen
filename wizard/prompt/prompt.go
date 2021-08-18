package prompt

import (
	"errors"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

func SelectOneOf(label string, variants []string, withAdd bool) string {
	if len(variants) == 0 {
		// it's better to show a prompt
		return InputNonEmpty(label, "")
	}

	if withAdd {
		p := promptui.SelectWithAdd{
			Label:  label,
			Stdout: os.Stderr,
			Items:  variants,
		}

		_, res, _ := p.Run()
		return res
	}

	p := promptui.Select{
		Label:  label,
		Stdout: os.Stderr,
		Items:  variants,
	}

	_, res, _ := p.Run()
	return res
}

func Input(label, defaultString string) string {
	p := promptui.Prompt{
		Label:  label,
		Stdout: os.Stderr,
		Validate: func(s string) error {
			return nil
		},
		Default: defaultString,
	}

	res, _ := p.Run()

	return res
}

func InputNonEmpty(label, defaultString string) string {
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

	res, _ := p.Run()

	return res
}

func FilePath(label, defaultPath string, shouldExist bool) string {
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

	res, _ := p.Run()

	return res
}

func Confirm(question string) bool {
	p := promptui.Prompt{
		Label:     question,
		Stdout:    os.Stderr,
		IsConfirm: true,
	}

	_, err := p.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false
		}
	}

	return true
}

func fileExists(path string) bool {
	// check if file exists
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}
