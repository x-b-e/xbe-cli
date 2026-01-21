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

type doUserCredentialsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	UserID                         string
	UserCredentialClassificationID string
	IssuedOn                       string
	ExpiresOn                      string
}

func newDoUserCredentialsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user credential",
		Long: `Create a new user credential.

Required flags:
  --user                              User ID (required)
  --user-credential-classification    Classification ID (required)

Optional flags:
  --issued-on     Issue date (YYYY-MM-DD)
  --expires-on    Expiration date (YYYY-MM-DD)`,
		Example: `  # Create a user credential
  xbe do user-credentials create --user 123 --user-credential-classification 456

  # Create with dates
  xbe do user-credentials create --user 123 --user-credential-classification 456 --issued-on 2024-01-01 --expires-on 2025-01-01`,
		Args: cobra.NoArgs,
		RunE: runDoUserCredentialsCreate,
	}
	initDoUserCredentialsCreateFlags(cmd)
	return cmd
}

func init() {
	doUserCredentialsCmd.AddCommand(newDoUserCredentialsCreateCmd())
}

func initDoUserCredentialsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("user-credential-classification", "", "Classification ID (required)")
	cmd.Flags().String("issued-on", "", "Issue date (YYYY-MM-DD)")
	cmd.Flags().String("expires-on", "", "Expiration date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserCredentialsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUserCredentialsCreateOptions(cmd)
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

	if opts.UserID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.UserCredentialClassificationID == "" {
		err := fmt.Errorf("--user-credential-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.IssuedOn != "" {
		attributes["issued-on"] = opts.IssuedOn
	}
	if opts.ExpiresOn != "" {
		attributes["expires-on"] = opts.ExpiresOn
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
		"user-credential-classification": map[string]any{
			"data": map[string]any{
				"type": "user-credential-classifications",
				"id":   opts.UserCredentialClassificationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "user-credentials",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/user-credentials", jsonBody)
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

	row := buildUserCredentialRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created user credential %s\n", row.ID)
	return nil
}

func parseDoUserCredentialsCreateOptions(cmd *cobra.Command) (doUserCredentialsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	classificationID, _ := cmd.Flags().GetString("user-credential-classification")
	issuedOn, _ := cmd.Flags().GetString("issued-on")
	expiresOn, _ := cmd.Flags().GetString("expires-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserCredentialsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		UserID:                         userID,
		UserCredentialClassificationID: classificationID,
		IssuedOn:                       issuedOn,
		ExpiresOn:                      expiresOn,
	}, nil
}

func buildUserCredentialRowFromSingle(resp jsonAPISingleResponse) userCredentialRow {
	attrs := resp.Data.Attributes

	row := userCredentialRow{
		ID:        resp.Data.ID,
		IssuedOn:  stringAttr(attrs, "issued-on"),
		ExpiresOn: stringAttr(attrs, "expires-on"),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["user-credential-classification"]; ok && rel.Data != nil {
		row.UserCredentialClassificationID = rel.Data.ID
	}

	return row
}
