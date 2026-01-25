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

type timeCardTimeChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardTimeChangeDetails struct {
	ID                    string `json:"id"`
	TimeCardID            string `json:"time_card_id,omitempty"`
	CreatedByID           string `json:"created_by_id,omitempty"`
	Comment               string `json:"comment,omitempty"`
	IsProcessed           bool   `json:"is_processed"`
	TimeChangesAttributes any    `json:"time_changes_attributes,omitempty"`
	TimeChangesDetails    any    `json:"time_changes_details,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
}

func newTimeCardTimeChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card time change details",
		Long: `Show the full details of a time card time change.

Output Fields:
  ID
  Time Card ID
  Created By ID
  Comment
  Is Processed
  Time Changes Attributes
  Time Changes Details
  Created At
  Updated At

Arguments:
  <id>    The time card time change ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show time card time change details
  xbe view time-card-time-changes show 123

  # Get JSON output
  xbe view time-card-time-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardTimeChangesShow,
	}
	initTimeCardTimeChangesShowFlags(cmd)
	return cmd
}

func init() {
	timeCardTimeChangesCmd.AddCommand(newTimeCardTimeChangesShowCmd())
}

func initTimeCardTimeChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardTimeChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardTimeChangesShowOptions(cmd)
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
		return fmt.Errorf("time card time change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-card-time-changes]", "comment,is-processed,time-changes-attributes,time-changes-details,created-at,updated-at,time-card,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-time-changes/"+id, query)
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

	details := buildTimeCardTimeChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardTimeChangeDetails(cmd, details)
}

func parseTimeCardTimeChangesShowOptions(cmd *cobra.Command) (timeCardTimeChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardTimeChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardTimeChangeDetails(resp jsonAPISingleResponse) timeCardTimeChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := timeCardTimeChangeDetails{
		ID:                    resource.ID,
		TimeCardID:            relationshipIDFromMap(resource.Relationships, "time-card"),
		CreatedByID:           relationshipIDFromMap(resource.Relationships, "created-by"),
		Comment:               stringAttr(attrs, "comment"),
		IsProcessed:           boolAttr(attrs, "is-processed"),
		TimeChangesAttributes: attrs["time-changes-attributes"],
		TimeChangesDetails:    attrs["time-changes-details"],
		CreatedAt:             formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:             formatDateTime(stringAttr(attrs, "updated-at")),
	}

	return details
}

func renderTimeCardTimeChangeDetails(cmd *cobra.Command, details timeCardTimeChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card ID: %s\n", details.TimeCardID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}
	fmt.Fprintf(out, "Processed: %t\n", details.IsProcessed)
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	if details.TimeChangesAttributes != nil {
		fmt.Fprintln(out, "\nTime Changes Attributes:")
		fmt.Fprintln(out, formatDetailsValue(details.TimeChangesAttributes))
	}
	if details.TimeChangesDetails != nil {
		fmt.Fprintln(out, "\nTime Changes Details:")
		fmt.Fprintln(out, formatDetailsValue(details.TimeChangesDetails))
	}

	return nil
}

func formatDetailsValue(value any) string {
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return ""
		}
		return typed
	default:
		return formatJSONBlock(value, "  ")
	}
}
