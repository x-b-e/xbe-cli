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

type doTrailerCredentialsCreateOptions struct {
	BaseURL                                  string
	Token                                    string
	JSON                                     bool
	TrailerID                                string
	TractorTrailerCredentialClassificationID string
	IssuedOn                                 string
	ExpiresOn                                string
}

func newDoTrailerCredentialsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new trailer credential",
		Long: `Create a new trailer credential.

Required flags:
  --trailer                                    Trailer ID (required)
  --tractor-trailer-credential-classification  Classification ID (required)

Optional flags:
  --issued-on     Issue date (YYYY-MM-DD)
  --expires-on    Expiration date (YYYY-MM-DD)`,
		Example: `  # Create a trailer credential
  xbe do trailer-credentials create --trailer 123 --tractor-trailer-credential-classification 456

  # Create with dates
  xbe do trailer-credentials create --trailer 123 --tractor-trailer-credential-classification 456 --issued-on 2024-01-01 --expires-on 2025-01-01`,
		Args: cobra.NoArgs,
		RunE: runDoTrailerCredentialsCreate,
	}
	initDoTrailerCredentialsCreateFlags(cmd)
	return cmd
}

func init() {
	doTrailerCredentialsCmd.AddCommand(newDoTrailerCredentialsCreateCmd())
}

func initDoTrailerCredentialsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trailer", "", "Trailer ID (required)")
	cmd.Flags().String("tractor-trailer-credential-classification", "", "Classification ID (required)")
	cmd.Flags().String("issued-on", "", "Issue date (YYYY-MM-DD)")
	cmd.Flags().String("expires-on", "", "Expiration date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTrailerCredentialsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTrailerCredentialsCreateOptions(cmd)
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

	if opts.TrailerID == "" {
		err := fmt.Errorf("--trailer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TractorTrailerCredentialClassificationID == "" {
		err := fmt.Errorf("--tractor-trailer-credential-classification is required")
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
		"trailer": map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.TrailerID,
			},
		},
		"tractor-trailer-credential-classification": map[string]any{
			"data": map[string]any{
				"type": "tractor-trailer-credential-classifications",
				"id":   opts.TractorTrailerCredentialClassificationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trailer-credentials",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trailer-credentials", jsonBody)
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

	row := buildTrailerCredentialRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created trailer credential %s\n", row.ID)
	return nil
}

func parseDoTrailerCredentialsCreateOptions(cmd *cobra.Command) (doTrailerCredentialsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trailerID, _ := cmd.Flags().GetString("trailer")
	classificationID, _ := cmd.Flags().GetString("tractor-trailer-credential-classification")
	issuedOn, _ := cmd.Flags().GetString("issued-on")
	expiresOn, _ := cmd.Flags().GetString("expires-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTrailerCredentialsCreateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		TrailerID:                                trailerID,
		TractorTrailerCredentialClassificationID: classificationID,
		IssuedOn:                                 issuedOn,
		ExpiresOn:                                expiresOn,
	}, nil
}

func buildTrailerCredentialRowFromSingle(resp jsonAPISingleResponse) trailerCredentialRow {
	attrs := resp.Data.Attributes

	row := trailerCredentialRow{
		ID:        resp.Data.ID,
		IssuedOn:  stringAttr(attrs, "issued-on"),
		ExpiresOn: stringAttr(attrs, "expires-on"),
	}

	if rel, ok := resp.Data.Relationships["trailer"]; ok && rel.Data != nil {
		row.TrailerID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["tractor-trailer-credential-classification"]; ok && rel.Data != nil {
		row.TractorTrailerCredentialClassificationID = rel.Data.ID
	}

	return row
}
