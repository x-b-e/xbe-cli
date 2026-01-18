package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication helpers",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store an access token",
	RunE:  runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show auth status",
	RunE:  runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored token",
	RunE:  runAuthLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)

	authLoginCmd.Flags().String("token", "", "Access token")
	authLoginCmd.Flags().Bool("token-stdin", false, "Read token from stdin")
	authLoginCmd.Flags().String("base-url", defaultBaseURL(), "API base URL")

	authStatusCmd.Flags().String("base-url", defaultBaseURL(), "API base URL")

	authLogoutCmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
}

func runAuthLogin(cmd *cobra.Command, _ []string) error {
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return err
	}
	tokenFlag, err := cmd.Flags().GetString("token")
	if err != nil {
		return err
	}
	useStdin, err := cmd.Flags().GetBool("token-stdin")
	if err != nil {
		return err
	}

	if tokenFlag != "" && useStdin {
		return errors.New("use either --token or --token-stdin")
	}

	token, err := readToken(cmd, tokenFlag, useStdin)
	if err != nil {
		return err
	}
	if token == "" {
		return errors.New("token is required")
	}

	store := auth.DefaultStore()
	normalized := auth.NormalizeBaseURL(baseURL)
	if err := store.Set(normalized, token); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Token stored for %s\n", normalized)
	return nil
}

func runAuthStatus(cmd *cobra.Command, _ []string) error {
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return err
	}
	normalized := auth.NormalizeBaseURL(baseURL)

	token, source, err := auth.ResolveToken(normalized, "")
	if err != nil && !errors.Is(err, auth.ErrNotFound) {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Base URL: %s\n", normalized)
	if token == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Token: not set")
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Token: set (source: %s)\n", source)
	return nil
}

func runAuthLogout(cmd *cobra.Command, _ []string) error {
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return err
	}
	normalized := auth.NormalizeBaseURL(baseURL)

	store := auth.DefaultStore()
	if err := store.Delete(normalized); err != nil {
		if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.OutOrStdout(), "No token found")
			return nil
		}
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Token removed")
	return nil
}

func readToken(cmd *cobra.Command, tokenFlag string, useStdin bool) (string, error) {
	if tokenFlag != "" {
		return strings.TrimSpace(tokenFlag), nil
	}

	if useStdin {
		return readTokenFromStdin(cmd.InOrStdin())
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(cmd.ErrOrStderr(), "Access token: ")
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}

	return readTokenFromStdin(cmd.InOrStdin())
}

func readTokenFromStdin(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
