package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

// runEditor opens the specified file in the user's preferred editor.
// It gets the editor from the EDITOR environment variable.
// If EDITOR is not set, it falls back to "nano".
// If EDITOR is set to an empty string, it fails with an error.
// Returns an error if the editor cannot be found or executed.
func runEditor(filePath string) error {
	editor, exists := os.LookupEnv("EDITOR")

	if exists && editor == "" {
		return fmt.Errorf("EDITOR environment variable is set to empty string")
	}

	if !exists {
		if *debug {
			fmt.Fprintln(os.Stderr, "No editor set. Set $EDITOR to your preferred editor")
		}
		editor = "nano"
	}

	editorFromPath, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("could not find %s in path: %w", editor, err)
	}

	cmd := exec.Command(editorFromPath, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run %s: %w", editorFromPath, err)
	}

	return nil
}

const AutoApplyWarningMessage = "# WARNING: Your file will be applied automatically once saved. If you do not want to apply anything, save an  empty file.\n"
