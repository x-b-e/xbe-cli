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

type versionEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type versionEventDetails struct {
	ID               string `json:"id"`
	EventAt          string `json:"event_at,omitempty"`
	EventType        string `json:"event_type,omitempty"`
	EventItemType    string `json:"event_item_type,omitempty"`
	EventItemID      string `json:"event_item_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
	EventItemChanges any    `json:"event_item_changes,omitempty"`
}

func newVersionEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show version event details",
		Long: `Show the full details of a version event.

Output Fields:
  ID
  Event At
  Event Type
  Event Item Type
  Event Item ID
  Broker ID
  Created At
  Updated At
  Event Item Changes

Arguments:
  <id>    The version event ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show version event details
  xbe view version-events show 123

  # Get JSON output
  xbe view version-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runVersionEventsShow,
	}
	initVersionEventsShowFlags(cmd)
	return cmd
}

func init() {
	versionEventsCmd.AddCommand(newVersionEventsShowCmd())
}

func initVersionEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runVersionEventsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseVersionEventsShowOptions(cmd)
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
		return fmt.Errorf("version event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[version-events]", "event-at,event-type,event-item-changes,created-at,updated-at,broker,event-item")

	body, _, err := client.Get(cmd.Context(), "/v1/version-events/"+id, query)
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

	details := buildVersionEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderVersionEventDetails(cmd, details)
}

func parseVersionEventsShowOptions(cmd *cobra.Command) (versionEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return versionEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildVersionEventDetails(resp jsonAPISingleResponse) versionEventDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := versionEventDetails{
		ID:               resource.ID,
		EventAt:          formatDateTime(stringAttr(attrs, "event-at")),
		EventType:        stringAttr(attrs, "event-type"),
		CreatedAt:        formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:        formatDateTime(stringAttr(attrs, "updated-at")),
		EventItemChanges: anyAttr(attrs, "event-item-changes"),
	}

	details.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")

	if rel, ok := resource.Relationships["event-item"]; ok && rel.Data != nil {
		details.EventItemID = rel.Data.ID
		details.EventItemType = rel.Data.Type
	}

	return details
}

func renderVersionEventDetails(cmd *cobra.Command, details versionEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EventAt != "" {
		fmt.Fprintf(out, "Event At: %s\n", details.EventAt)
	}
	if details.EventType != "" {
		fmt.Fprintf(out, "Event Type: %s\n", details.EventType)
	}
	if details.EventItemType != "" {
		fmt.Fprintf(out, "Event Item Type: %s\n", details.EventItemType)
	}
	if details.EventItemID != "" {
		fmt.Fprintf(out, "Event Item ID: %s\n", details.EventItemID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.EventItemChanges != nil {
		fmt.Fprintln(out, "Event Item Changes:")
		if err := writeJSON(out, details.EventItemChanges); err != nil {
			return err
		}
	}

	return nil
}
