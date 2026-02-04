package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type actionItemLineItemsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type actionItemLineItemDetails struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	Status                string `json:"status"`
	DueOn                 string `json:"due_on,omitempty"`
	ActionItemID          string `json:"action_item_id,omitempty"`
	ActionItemTitle       string `json:"action_item_title,omitempty"`
	ResponsiblePersonID   string `json:"responsible_person_id,omitempty"`
	ResponsiblePersonName string `json:"responsible_person_name,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newActionItemLineItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show action item line item details",
		Long: `Show the full details of an action item line item.

Output Fields:
  ID
  Title
  Status
  Due On
  Action Item (ID + title)
  Responsible Person (ID + name)
  Created At
  Updated At

Arguments:
  <id>    The action item line item ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an action item line item
  xbe view action-item-line-items show 123

  # Output as JSON
  xbe view action-item-line-items show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runActionItemLineItemsShow,
	}
	initActionItemLineItemsShowFlags(cmd)
	return cmd
}

func init() {
	actionItemLineItemsCmd.AddCommand(newActionItemLineItemsShowCmd())
}

func initActionItemLineItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemLineItemsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseActionItemLineItemsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("action item line item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[action-item-line-items]", "title,status,due-on,responsible-person,action-item,created-at,updated-at")
	query.Set("include", "responsible-person,action-item")
	query.Set("fields[users]", "name")
	query.Set("fields[action-items]", "title")

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-line-items/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildActionItemLineItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemLineItemDetails(cmd, details)
}

func parseActionItemLineItemsShowOptions(cmd *cobra.Command) (actionItemLineItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemLineItemsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildActionItemLineItemDetails(resp jsonAPISingleResponse) actionItemLineItemDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	attrs := resp.Data.Attributes
	details := actionItemLineItemDetails{
		ID:        resp.Data.ID,
		Title:     strings.TrimSpace(stringAttr(attrs, "title")),
		Status:    stringAttr(attrs, "status"),
		DueOn:     formatDate(stringAttr(attrs, "due-on")),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["responsible-person"]; ok && rel.Data != nil {
		details.ResponsiblePersonID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ResponsiblePersonName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["action-item"]; ok && rel.Data != nil {
		details.ActionItemID = rel.Data.ID
		if inc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ActionItemTitle = strings.TrimSpace(stringAttr(inc.Attributes, "title"))
		}
	}

	return details
}

func renderActionItemLineItemDetails(cmd *cobra.Command, details actionItemLineItemDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Title != "" {
		fmt.Fprintf(out, "Title: %s\n", details.Title)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.DueOn != "" {
		fmt.Fprintf(out, "Due On: %s\n", details.DueOn)
	}
	if details.ActionItemID != "" {
		label := details.ActionItemID
		if details.ActionItemTitle != "" {
			label = fmt.Sprintf("%s (%s)", details.ActionItemID, details.ActionItemTitle)
		}
		fmt.Fprintf(out, "Action Item: %s\n", label)
	}
	if details.ResponsiblePersonID != "" {
		label := details.ResponsiblePersonID
		if details.ResponsiblePersonName != "" {
			label = fmt.Sprintf("%s (%s)", details.ResponsiblePersonID, details.ResponsiblePersonName)
		}
		fmt.Fprintf(out, "Responsible Person: %s\n", label)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
