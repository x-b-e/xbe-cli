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

type doActionItemLineItemsUpdateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ID                string
	Title             string
	Status            string
	DueOn             string
	ResponsiblePerson string
}

func newDoActionItemLineItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an action item line item",
		Long: `Update an action item line item.

Optional flags:
  --title               Line item title
  --status              Status (open/closed)
  --due-on              Due date (YYYY-MM-DD)
  --responsible-person  Responsible person user ID (set empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update status
  xbe do action-item-line-items update 456 --status closed

  # Update title and due date
  xbe do action-item-line-items update 456 --title "Follow up" --due-on 2025-03-01

  # Clear responsible person
  xbe do action-item-line-items update 456 --responsible-person ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoActionItemLineItemsUpdate,
	}
	initDoActionItemLineItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doActionItemLineItemsCmd.AddCommand(newDoActionItemLineItemsUpdateCmd())
}

func initDoActionItemLineItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("title", "", "Line item title")
	cmd.Flags().String("status", "", "Status (open/closed)")
	cmd.Flags().String("due-on", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().String("responsible-person", "", "Responsible person user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemLineItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemLineItemsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("title") {
		attributes["title"] = opts.Title
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("due-on") {
		attributes["due-on"] = opts.DueOn
	}
	if cmd.Flags().Changed("responsible-person") {
		if strings.TrimSpace(opts.ResponsiblePerson) == "" {
			relationships["responsible-person"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["responsible-person"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.ResponsiblePerson,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "action-item-line-items",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Patch(cmd.Context(), "/v1/action-item-line-items/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated action item line item %s\n", row.ID)
	return nil
}

func parseDoActionItemLineItemsUpdateOptions(cmd *cobra.Command, args []string) (doActionItemLineItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	title, _ := cmd.Flags().GetString("title")
	status, _ := cmd.Flags().GetString("status")
	dueOn, _ := cmd.Flags().GetString("due-on")
	responsiblePerson, _ := cmd.Flags().GetString("responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemLineItemsUpdateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ID:                args[0],
		Title:             title,
		Status:            status,
		DueOn:             dueOn,
		ResponsiblePerson: responsiblePerson,
	}, nil
}
