package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doApiTokensUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	RevokedAt string
}

func newDoApiTokensUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an API token",
		Long: `Update an API token.

Only revocation is supported. Setting revoked-at immediately revokes the token.

Optional flags:
  --revoked-at    Revocation timestamp (RFC3339)`,
		Example: `  # Revoke a token now
  xbe do api-tokens update 123 --revoked-at 2026-01-01T00:00:00Z`,
		Args: cobra.ExactArgs(1),
		RunE: runDoApiTokensUpdate,
	}
	initDoApiTokensUpdateFlags(cmd)
	return cmd
}

func init() {
	doApiTokensCmd.AddCommand(newDoApiTokensUpdateCmd())
}

func initDoApiTokensUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("revoked-at", "", "Revocation timestamp (RFC3339)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoApiTokensUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoApiTokensUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("revoked-at") {
		attributes["revoked-at"] = opts.RevokedAt
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "api-tokens",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/api-tokens/"+opts.ID, jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	row := buildApiTokenRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated API token %s\n", row.ID)
	return nil
}

func parseDoApiTokensUpdateOptions(cmd *cobra.Command, args []string) (doApiTokensUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	revokedAt, _ := cmd.Flags().GetString("revoked-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doApiTokensUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		RevokedAt: revokedAt,
	}, nil
}
