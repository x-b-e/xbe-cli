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

type doTractorCredentialsUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	IssuedOn  string
	ExpiresOn string
}

func newDoTractorCredentialsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tractor credential",
		Long: `Update a tractor credential.

Optional flags:
  --issued-on     Issue date (YYYY-MM-DD)
  --expires-on    Expiration date (YYYY-MM-DD)`,
		Example: `  # Update issue date
  xbe do tractor-credentials update 123 --issued-on 2024-01-15

  # Update expiration date
  xbe do tractor-credentials update 123 --expires-on 2025-06-30`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTractorCredentialsUpdate,
	}
	initDoTractorCredentialsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTractorCredentialsCmd.AddCommand(newDoTractorCredentialsUpdateCmd())
}

func initDoTractorCredentialsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("issued-on", "", "Issue date (YYYY-MM-DD)")
	cmd.Flags().String("expires-on", "", "Expiration date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorCredentialsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorCredentialsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}

	if cmd.Flags().Changed("issued-on") {
		attributes["issued-on"] = opts.IssuedOn
	}
	if cmd.Flags().Changed("expires-on") {
		attributes["expires-on"] = opts.ExpiresOn
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "tractor-credentials",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/tractor-credentials/"+opts.ID, jsonBody)
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

	row := buildTractorCredentialRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tractor credential %s\n", row.ID)
	return nil
}

func parseDoTractorCredentialsUpdateOptions(cmd *cobra.Command, args []string) (doTractorCredentialsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	issuedOn, _ := cmd.Flags().GetString("issued-on")
	expiresOn, _ := cmd.Flags().GetString("expires-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorCredentialsUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		IssuedOn:  issuedOn,
		ExpiresOn: expiresOn,
	}, nil
}
