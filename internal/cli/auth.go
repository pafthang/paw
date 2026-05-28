package cli

import (
	"fmt"
	"io"

	pawauth "github.com/pafthang/paw/internal/auth"
	"github.com/pafthang/paw/internal/config"
	"github.com/spf13/cobra"
)

func newAuthCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage local API access token", RunE: func(cmd *cobra.Command, args []string) error {
		return runAuthToken(out)
	}}
	cmd.AddCommand(
		&cobra.Command{Use: "token", Short: "Create/read and print the local access token", RunE: func(cmd *cobra.Command, args []string) error { return runAuthToken(out) }},
		&cobra.Command{Use: "path", Short: "Print access token path", RunE: func(cmd *cobra.Command, args []string) error { fmt.Fprintln(out, must(config.AccessTokenPath())); return nil }},
		&cobra.Command{Use: "rotate", Short: "Generate and save a new access token", RunE: func(cmd *cobra.Command, args []string) error { return runAuthRotate(out) }},
	)
	return cmd
}

func runAuthToken(out io.Writer) error {
	token, err := pawauth.EnsureToken()
	if err != nil {
		return err
	}
	fmt.Fprintln(out, token)
	return nil
}

func runAuthRotate(out io.Writer) error {
	token, err := pawauth.RotateToken()
	if err != nil {
		return err
	}
	fmt.Fprintln(out, token)
	return nil
}
