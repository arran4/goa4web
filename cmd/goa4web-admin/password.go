package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// promptPassword reads a password interactively from stdin.
func promptPassword() (string, error) {
	fmt.Fprint(os.Stderr, "Password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}
