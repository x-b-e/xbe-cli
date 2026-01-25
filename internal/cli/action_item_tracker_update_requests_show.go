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

type actionItemTrackerUpdateRequestsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type actionItemTrackerUpdateRequestDetails struct {
	ID                  string `json:"id"`
	ActionItemTrackerID string `json:"action_item_tracker_id,omitempty"`
	RequestedByID       string `json:"requested_by_id,omitempty"`
	RequestedByName     string `json:"requested_by_name,omitempty"`
	RequestedByEmail    string `json:"requested_by_email,omitempty"`
	RequestedFromID     string `json:"requested_from_id,omitempty"`
	RequestedFromName   string `json:"requested_from_name,omitempty"`
	RequestedFromEmail  string `json:"requested_from_email,omitempty"`
	RequestNote         string `json:"request_note,omitempty"`
	DueOn               string `json:"due_on,omitempty"`
	UpdateNote          string `json:"update_note,omitempty"`
	Status              string `json:"status,omitempty"`
}

func newActionItemTrackerUpdateRequestsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show action item tracker update request details",
		Long: `Show the full details of an action item tracker update request.

Output Fields:
  ID              Update request identifier
  TRACKER         Action item tracker ID
  REQUESTED BY    User who requested the update
  REQUESTED FROM  User who should provide the update
  DUE ON          Requested due date
  STATUS          pending or fulfilled
  REQUEST NOTE    Requested update note
  UPDATE NOTE     Fulfillment update note

Arguments:
  <id>  The update request ID (required).`,
		Example: `  # Show an update request
  xbe view action-item-tracker-update-requests show 123

  # Output as JSON
  xbe view action-item-tracker-update-requests show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runActionItemTrackerUpdateRequestsShow,
	}
	initActionItemTrackerUpdateRequestsShowFlags(cmd)
	return cmd
}

func init() {
	actionItemTrackerUpdateRequestsCmd.AddCommand(newActionItemTrackerUpdateRequestsShowCmd())
}

func initActionItemTrackerUpdateRequestsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemTrackerUpdateRequestsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseActionItemTrackerUpdateRequestsShowOptions(cmd)
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
		return fmt.Errorf("action item tracker update request id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[action-item-tracker-update-requests]", "request-note,due-on,update-note,action-item-tracker,requested-by,requested-from")
	query.Set("include", "requested-by,requested-from")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-tracker-update-requests/"+id, query)
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

	details := buildActionItemTrackerUpdateRequestDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemTrackerUpdateRequestDetails(cmd, details)
}

func parseActionItemTrackerUpdateRequestsShowOptions(cmd *cobra.Command) (actionItemTrackerUpdateRequestsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemTrackerUpdateRequestsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildActionItemTrackerUpdateRequestDetails(resp jsonAPISingleResponse) actionItemTrackerUpdateRequestDetails {
	attrs := resp.Data.Attributes

	details := actionItemTrackerUpdateRequestDetails{
		ID:          resp.Data.ID,
		RequestNote: strings.TrimSpace(stringAttr(attrs, "request-note")),
		DueOn:       formatDate(stringAttr(attrs, "due-on")),
		UpdateNote:  strings.TrimSpace(stringAttr(attrs, "update-note")),
	}

	if strings.TrimSpace(details.UpdateNote) != "" {
		details.Status = "fulfilled"
	} else {
		details.Status = "pending"
	}

	requestedByType := ""
	requestedFromType := ""
	if rel, ok := resp.Data.Relationships["action-item-tracker"]; ok && rel.Data != nil {
		details.ActionItemTrackerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["requested-by"]; ok && rel.Data != nil {
		details.RequestedByID = rel.Data.ID
		requestedByType = rel.Data.Type
	}
	if rel, ok := resp.Data.Relationships["requested-from"]; ok && rel.Data != nil {
		details.RequestedFromID = rel.Data.ID
		requestedFromType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return details
	}

	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.RequestedByID != "" && requestedByType != "" {
		if user, ok := included[resourceKey(requestedByType, details.RequestedByID)]; ok {
			details.RequestedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			details.RequestedByEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
		}
	}
	if details.RequestedFromID != "" && requestedFromType != "" {
		if user, ok := included[resourceKey(requestedFromType, details.RequestedFromID)]; ok {
			details.RequestedFromName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			details.RequestedFromEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
		}
	}

	return details
}

func renderActionItemTrackerUpdateRequestDetails(cmd *cobra.Command, details actionItemTrackerUpdateRequestDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ActionItemTrackerID != "" {
		fmt.Fprintf(out, "Tracker: %s\n", details.ActionItemTrackerID)
	}
	if details.RequestedByID != "" {
		fmt.Fprintf(out, "Requested By ID: %s\n", details.RequestedByID)
	}
	if details.RequestedByName != "" {
		fmt.Fprintf(out, "Requested By Name: %s\n", details.RequestedByName)
	}
	if details.RequestedByEmail != "" {
		fmt.Fprintf(out, "Requested By Email: %s\n", details.RequestedByEmail)
	}
	if details.RequestedFromID != "" {
		fmt.Fprintf(out, "Requested From ID: %s\n", details.RequestedFromID)
	}
	if details.RequestedFromName != "" {
		fmt.Fprintf(out, "Requested From Name: %s\n", details.RequestedFromName)
	}
	if details.RequestedFromEmail != "" {
		fmt.Fprintf(out, "Requested From Email: %s\n", details.RequestedFromEmail)
	}
	if details.DueOn != "" {
		fmt.Fprintf(out, "Due On: %s\n", details.DueOn)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.RequestNote != "" {
		fmt.Fprintf(out, "Request Note: %s\n", details.RequestNote)
	}
	if details.UpdateNote != "" {
		fmt.Fprintf(out, "Update Note: %s\n", details.UpdateNote)
	}

	return nil
}
