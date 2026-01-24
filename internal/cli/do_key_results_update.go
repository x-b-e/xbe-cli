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

type doKeyResultsUpdateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	Title                            string
	TitleSummaryExplicit             string
	StartOn                          string
	EndOn                            string
	CompletionPercentage             string
	Owner                            string
	CustomerSuccessResponsiblePerson string
}

func newDoKeyResultsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a key result",
		Long: `Update an existing key result.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The key result ID (required)

Flags:
  --title                               Update the title
  --title-summary-explicit              Update the explicit title summary
  --start-on                            Update the start date (YYYY-MM-DD)
  --end-on                              Update the end date (YYYY-MM-DD)
  --completion-percentage               Update the completion percentage (0-1)
  --owner                               Update the owner user ID
  --customer-success-responsible-person Update the customer success responsible person user ID`,
		Example: `  # Update the title
  xbe do key-results update 123 --title "Updated KR"

  # Update dates
  xbe do key-results update 123 --start-on 2025-01-01 --end-on 2025-06-30

  # Update completion percentage
  xbe do key-results update 123 --completion-percentage 0.75

  # Output as JSON
  xbe do key-results update 123 --title "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoKeyResultsUpdate,
	}
	initDoKeyResultsUpdateFlags(cmd)
	return cmd
}

func init() {
	doKeyResultsCmd.AddCommand(newDoKeyResultsUpdateCmd())
}

func initDoKeyResultsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("title-summary-explicit", "", "New explicit title summary")
	cmd.Flags().String("start-on", "", "New start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "New end date (YYYY-MM-DD)")
	cmd.Flags().String("completion-percentage", "", "New completion percentage (0-1)")
	cmd.Flags().String("owner", "", "New owner user ID")
	cmd.Flags().String("customer-success-responsible-person", "", "New customer success responsible person user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoKeyResultsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoKeyResultsUpdateOptions(cmd)
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
		return fmt.Errorf("key result id is required")
	}

	if opts.Title == "" && opts.TitleSummaryExplicit == "" && opts.StartOn == "" && opts.EndOn == "" &&
		opts.CompletionPercentage == "" && opts.Owner == "" && opts.CustomerSuccessResponsiblePerson == "" {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Title != "" {
		attributes["title"] = opts.Title
	}
	if opts.TitleSummaryExplicit != "" {
		attributes["title-summary-explicit"] = opts.TitleSummaryExplicit
	}
	if opts.StartOn != "" {
		attributes["start-on"] = opts.StartOn
	}
	if opts.EndOn != "" {
		attributes["end-on"] = opts.EndOn
	}
	if opts.CompletionPercentage != "" {
		attributes["completion-percentage"] = opts.CompletionPercentage
	}

	data := map[string]any{
		"id":         id,
		"type":       "key-results",
		"attributes": attributes,
	}

	relationships := map[string]any{}
	if opts.Owner != "" {
		relationships["owner"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.Owner,
			},
		}
	}
	if opts.CustomerSuccessResponsiblePerson != "" {
		relationships["customer-success-responsible-person"] = map[string]any{
			"data": map[string]string{
				"type": "users",
				"id":   opts.CustomerSuccessResponsiblePerson,
			},
		}
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/key-results/"+id, jsonBody)
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

	result := map[string]any{
		"id":     resp.Data.ID,
		"title":  stringAttr(resp.Data.Attributes, "title"),
		"status": stringAttr(resp.Data.Attributes, "status"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated key result %s (%s)\n", result["id"], result["title"])
	return nil
}

func parseDoKeyResultsUpdateOptions(cmd *cobra.Command) (doKeyResultsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	titleSummaryExplicit, _ := cmd.Flags().GetString("title-summary-explicit")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	completionPercentage, _ := cmd.Flags().GetString("completion-percentage")
	owner, _ := cmd.Flags().GetString("owner")
	customerSuccessResponsiblePerson, _ := cmd.Flags().GetString("customer-success-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doKeyResultsUpdateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		Title:                            title,
		TitleSummaryExplicit:             titleSummaryExplicit,
		StartOn:                          startOn,
		EndOn:                            endOn,
		CompletionPercentage:             completionPercentage,
		Owner:                            owner,
		CustomerSuccessResponsiblePerson: customerSuccessResponsiblePerson,
	}, nil
}
