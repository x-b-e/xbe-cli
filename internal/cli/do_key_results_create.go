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

type doKeyResultsCreateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	Title                            string
	TitleSummaryExplicit             string
	StartOn                          string
	EndOn                            string
	CompletionPercentage             string
	Objective                        string
	Owner                            string
	CustomerSuccessResponsiblePerson string
}

func newDoKeyResultsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a key result",
		Long: `Create a key result.

Required flags:
  --title        The key result title (required)
  --objective    Objective ID (required)

Optional flags:
  --title-summary-explicit             Explicit title summary
  --start-on                           Start date (YYYY-MM-DD)
  --end-on                             End date (YYYY-MM-DD)
  --completion-percentage              Completion percentage (0-1)
  --owner                              Owner user ID
  --customer-success-responsible-person  Customer success responsible person user ID`,
		Example: `  # Create a key result
  xbe do key-results create --title "Launch beta" --objective 123

  # Create with dates and completion
  xbe do key-results create --title "Ship v1" --objective 123 --start-on 2025-01-01 --end-on 2025-06-30 --completion-percentage 0.2

  # Create with owner
  xbe do key-results create --title "Improve uptime" --objective 123 --owner 456

  # Output as JSON
  xbe do key-results create --title "New KR" --objective 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoKeyResultsCreate,
	}
	initDoKeyResultsCreateFlags(cmd)
	return cmd
}

func init() {
	doKeyResultsCmd.AddCommand(newDoKeyResultsCreateCmd())
}

func initDoKeyResultsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Key result title (required)")
	cmd.Flags().String("objective", "", "Objective ID (required)")
	cmd.Flags().String("title-summary-explicit", "", "Explicit title summary")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("completion-percentage", "", "Completion percentage (0-1)")
	cmd.Flags().String("owner", "", "Owner user ID")
	cmd.Flags().String("customer-success-responsible-person", "", "Customer success responsible person user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoKeyResultsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoKeyResultsCreateOptions(cmd)
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

	if opts.Title == "" {
		err := fmt.Errorf("--title is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Objective == "" {
		err := fmt.Errorf("--objective is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"title": opts.Title,
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

	relationships := map[string]any{
		"objective": map[string]any{
			"data": map[string]string{
				"type": "objectives",
				"id":   opts.Objective,
			},
		},
	}
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "key-results",
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

	body, _, err := client.Post(cmd.Context(), "/v1/key-results", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created key result %s (%s)\n", result["id"], result["title"])
	return nil
}

func parseDoKeyResultsCreateOptions(cmd *cobra.Command) (doKeyResultsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	objective, _ := cmd.Flags().GetString("objective")
	titleSummaryExplicit, _ := cmd.Flags().GetString("title-summary-explicit")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	completionPercentage, _ := cmd.Flags().GetString("completion-percentage")
	owner, _ := cmd.Flags().GetString("owner")
	customerSuccessResponsiblePerson, _ := cmd.Flags().GetString("customer-success-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doKeyResultsCreateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		Title:                            title,
		Objective:                        objective,
		TitleSummaryExplicit:             titleSummaryExplicit,
		StartOn:                          startOn,
		EndOn:                            endOn,
		CompletionPercentage:             completionPercentage,
		Owner:                            owner,
		CustomerSuccessResponsiblePerson: customerSuccessResponsiblePerson,
	}, nil
}
