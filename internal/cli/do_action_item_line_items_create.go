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

type doActionItemLineItemsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	Title             string
	Status            string
	DueOn             string
	ResponsiblePerson string
	ActionItem        string
}

func newDoActionItemLineItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an action item line item",
		Long: `Create an action item line item.

Required flags:
  --action-item  Action item ID (required)
  --title        Line item title (required)

Optional flags:
  --status              Status (open/closed)
  --due-on              Due date (YYYY-MM-DD)
  --responsible-person  Responsible person user ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a line item
  xbe do action-item-line-items create --action-item 123 --title "Review plan"

  # Create with status and due date
  xbe do action-item-line-items create --action-item 123 --title "Approve" --status open --due-on 2025-02-01

  # Assign a responsible person
  xbe do action-item-line-items create --action-item 123 --title "Follow up" --responsible-person 456`,
		Args: cobra.NoArgs,
		RunE: runDoActionItemLineItemsCreate,
	}
	initDoActionItemLineItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doActionItemLineItemsCmd.AddCommand(newDoActionItemLineItemsCreateCmd())
}

func initDoActionItemLineItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Line item title (required)")
	cmd.Flags().String("status", "", "Status (open/closed)")
	cmd.Flags().String("due-on", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().String("responsible-person", "", "Responsible person user ID")
	cmd.Flags().String("action-item", "", "Action item ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemLineItemsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoActionItemLineItemsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ActionItem) == "" {
		err := fmt.Errorf("--action-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Title) == "" {
		err := fmt.Errorf("--title is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"title": opts.Title,
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.DueOn != "" {
		attributes["due-on"] = opts.DueOn
	}

	relationships := map[string]any{
		"action-item": map[string]any{
			"data": map[string]any{
				"type": "action-items",
				"id":   opts.ActionItem,
			},
		},
	}
	if opts.ResponsiblePerson != "" {
		relationships["responsible-person"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.ResponsiblePerson,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "action-item-line-items",
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

	body, _, err := client.Post(cmd.Context(), "/v1/action-item-line-items", jsonBody)
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

	row := buildActionItemLineItemRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created action item line item %s\n", row.ID)
	return nil
}

func parseDoActionItemLineItemsCreateOptions(cmd *cobra.Command) (doActionItemLineItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	status, _ := cmd.Flags().GetString("status")
	dueOn, _ := cmd.Flags().GetString("due-on")
	responsiblePerson, _ := cmd.Flags().GetString("responsible-person")
	actionItem, _ := cmd.Flags().GetString("action-item")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemLineItemsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		Title:             title,
		Status:            status,
		DueOn:             dueOn,
		ResponsiblePerson: responsiblePerson,
		ActionItem:        actionItem,
	}, nil
}
