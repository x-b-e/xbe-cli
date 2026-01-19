package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long: `Manage authentication credentials for the XBE API.

The auth commands help you securely store and manage API tokens. Tokens are
stored in your system's secure credential storage:
  - macOS: Keychain
  - Linux: Secret Service (GNOME Keyring, KWallet)
  - Windows: Credential Manager

If secure storage is unavailable, tokens are stored in ~/.config/xbe/config.json

Token Resolution Order:
  1. --token flag (highest priority)
  2. XBE_TOKEN or XBE_API_TOKEN environment variable
  3. System keychain
  4. Config file (lowest priority)`,
	Annotations: map[string]string{"group": GroupAuth},
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store an access token",
	Long: `Store an access token for API authentication.

In interactive mode, opens your browser to the API tokens page where you can
create or copy a token. The token is validated before being stored.

You can provide the token via:
  - Interactive prompt (opens browser, hides input)
  - --token flag
  - --token-stdin flag (for piping from password managers)

Tokens are stored per base URL, allowing you to have different tokens
for different XBE environments (e.g., staging vs production).`,
	Example: `  # Interactive login (opens browser, prompts for token)
  xbe auth login

  # Interactive login without opening browser
  xbe auth login --no-web

  # Provide token via flag
  xbe auth login --token YOUR_TOKEN

  # Pipe token from a password manager
  op read "op://Vault/XBE/token" | xbe auth login --token-stdin

  # Store token for a different environment
  xbe auth login --base-url https://staging.x-b-e.com`,
	RunE: runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long: `Show the current authentication status.

Displays whether a token is configured for the specified base URL and
where the token is being loaded from (flag, environment, keychain, or file).

This is useful for debugging authentication issues or verifying your
configuration before running other commands.`,
	Example: `  # Check auth status for default URL
  xbe auth status

  # Check auth status for a specific environment
  xbe auth status --base-url https://staging.x-b-e.com`,
	RunE: runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored token",
	Long: `Remove the stored authentication token.

Deletes the token from secure storage for the specified base URL.
This does not affect tokens stored in environment variables.`,
	Example: `  # Remove token for default URL
  xbe auth logout

  # Remove token for a specific environment
  xbe auth logout --base-url https://staging.x-b-e.com`,
	RunE: runAuthLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)

	authLoginCmd.Flags().String("token", "", "Access token")
	authLoginCmd.Flags().Bool("token-stdin", false, "Read token from stdin")
	authLoginCmd.Flags().Bool("no-web", false, "Skip opening browser (for scripted/headless use)")
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
	noWeb, err := cmd.Flags().GetBool("no-web")
	if err != nil {
		return err
	}

	if tokenFlag != "" && useStdin {
		return errors.New("use either --token or --token-stdin")
	}

	normalized := auth.NormalizeBaseURL(baseURL)

	// Open browser in interactive mode (unless --no-web or using --token/--token-stdin)
	if tokenFlag == "" && !useStdin && !noWeb && term.IsTerminal(int(os.Stdin.Fd())) {
		url := tokenURL(baseURL)
		fmt.Fprintf(cmd.OutOrStdout(), "Opening %s ...\n", url)
		if err := openBrowser(url); err != nil {
			// Non-fatal: continue with prompt even if browser fails
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not open browser: %v\n", err)
		}
	}

	token, err := readToken(cmd, tokenFlag, useStdin)
	if err != nil {
		return err
	}
	if token == "" {
		return errors.New("token is required")
	}

	// Validate token before storing
	result, err := validateToken(cmd, normalized, token)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	store := auth.DefaultStore()
	if err := store.Set(normalized, token); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s (%s)\n", result.Name, result.Email)
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
		fmt.Fprint(cmd.ErrOrStderr(), "Paste your API token: ")
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

func tokenURL(baseURL string) string {
	if strings.Contains(baseURL, "staging") {
		return "https://staging-client.x-b-e.com/#/browse/users/me/api-tokens"
	}
	return "https://client.x-b-e.com/#/browse/users/me/api-tokens"
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

type tokenValidationResult struct {
	Name  string
	Email string
}

func validateToken(cmd *cobra.Command, baseURL, token string) (*tokenValidationResult, error) {
	client := api.NewClient(baseURL, token)

	query := url.Values{}
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/users/me", query)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tokenValidationResult{
		Name:  strings.TrimSpace(stringAttr(resp.Data.Attributes, "name")),
		Email: strings.TrimSpace(stringAttr(resp.Data.Attributes, "email-address")),
	}, nil
}
