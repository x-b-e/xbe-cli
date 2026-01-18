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

type doGlossaryTermsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	Term       string
	Definition string
	Source     string
}

func newDoGlossaryTermsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a glossary term",
		Long: `Update an existing glossary term.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The glossary term ID (required)

Flags:
  --term         Update the term name
  --definition   Update the definition
  --source       Update the source`,
		Example: `  # Update just the definition
  xbe do glossary-terms update 123 --definition "New definition"

  # Update multiple fields
  xbe do glossary-terms update 123 --term "New Term" --definition "New definition"

  # Get JSON output
  xbe do glossary-terms update 123 --definition "New definition" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoGlossaryTermsUpdate,
	}
	initDoGlossaryTermsUpdateFlags(cmd)
	return cmd
}

func init() {
	doGlossaryTermsCmd.AddCommand(newDoGlossaryTermsUpdateCmd())
}

func initDoGlossaryTermsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("term", "", "New term name")
	cmd.Flags().String("definition", "", "New definition")
	cmd.Flags().String("source", "", "New source")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGlossaryTermsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGlossaryTermsUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("glossary term id is required")
	}

	// Require at least one field to update
	if opts.Term == "" && opts.Definition == "" && opts.Source == "" {
		err := fmt.Errorf("at least one of --term, --definition, or --source is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build JSON:API request body
	attributes := map[string]string{}
	if opts.Term != "" {
		attributes["term"] = opts.Term
	}
	if opts.Definition != "" {
		attributes["definition"] = opts.Definition
	}
	if opts.Source != "" {
		attributes["source"] = opts.Source
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "glossary-terms",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/glossary-terms/"+id, jsonBody)
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

	details := buildGlossaryTermDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderGlossaryTermDetails(cmd, details)
}

func parseDoGlossaryTermsUpdateOptions(cmd *cobra.Command) (doGlossaryTermsUpdateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doGlossaryTermsUpdateOptions{}, err
	}
	term, err := cmd.Flags().GetString("term")
	if err != nil {
		return doGlossaryTermsUpdateOptions{}, err
	}
	definition, err := cmd.Flags().GetString("definition")
	if err != nil {
		return doGlossaryTermsUpdateOptions{}, err
	}
	source, err := cmd.Flags().GetString("source")
	if err != nil {
		return doGlossaryTermsUpdateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doGlossaryTermsUpdateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doGlossaryTermsUpdateOptions{}, err
	}

	return doGlossaryTermsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		Term:       term,
		Definition: definition,
		Source:     source,
	}, nil
}
