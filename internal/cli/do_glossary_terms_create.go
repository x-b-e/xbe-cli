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

type doGlossaryTermsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	Term       string
	Definition string
	Source     string
}

func newDoGlossaryTermsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new glossary term",
		Long: `Create a new glossary term.

Required flags:
  --term         The term name (required)
  --definition   The definition (required)

Optional flags:
  --source       The source of the definition`,
		Example: `  # Create a glossary term
  xbe do glossary-terms create --term "Paving" --definition "The process of laying asphalt"

  # Create with source
  xbe do glossary-terms create --term "Paving" --definition "The process of laying asphalt" --source "expert"

  # Get JSON output
  xbe do glossary-terms create --term "Paving" --definition "The process of laying asphalt" --json`,
		Args: cobra.NoArgs,
		RunE: runDoGlossaryTermsCreate,
	}
	initDoGlossaryTermsCreateFlags(cmd)
	return cmd
}

func init() {
	doGlossaryTermsCmd.AddCommand(newDoGlossaryTermsCreateCmd())
}

func initDoGlossaryTermsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("term", "", "Term name (required)")
	cmd.Flags().String("definition", "", "Definition (required)")
	cmd.Flags().String("source", "", "Source of the definition")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoGlossaryTermsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoGlossaryTermsCreateOptions(cmd)
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

	// Require term and definition
	if opts.Term == "" {
		err := fmt.Errorf("--term is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Definition == "" {
		err := fmt.Errorf("--definition is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build JSON:API request body
	attributes := map[string]string{
		"term":       opts.Term,
		"definition": opts.Definition,
	}
	if opts.Source != "" {
		attributes["source"] = opts.Source
	}

	requestBody := map[string]any{
		"data": map[string]any{
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

	body, _, err := client.Post(cmd.Context(), "/v1/glossary-terms", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created glossary term %s (%s)\n", details.ID, details.Term)
	return renderGlossaryTermDetails(cmd, details)
}

func parseDoGlossaryTermsCreateOptions(cmd *cobra.Command) (doGlossaryTermsCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doGlossaryTermsCreateOptions{}, err
	}
	term, err := cmd.Flags().GetString("term")
	if err != nil {
		return doGlossaryTermsCreateOptions{}, err
	}
	definition, err := cmd.Flags().GetString("definition")
	if err != nil {
		return doGlossaryTermsCreateOptions{}, err
	}
	source, err := cmd.Flags().GetString("source")
	if err != nil {
		return doGlossaryTermsCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doGlossaryTermsCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doGlossaryTermsCreateOptions{}, err
	}

	return doGlossaryTermsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		Term:       term,
		Definition: definition,
		Source:     source,
	}, nil
}
