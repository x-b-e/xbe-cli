package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type rootCausesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rootCauseDetails struct {
	ID           string   `json:"id"`
	Title        string   `json:"title,omitempty"`
	Description  string   `json:"description,omitempty"`
	IsTriaged    bool     `json:"is_triaged"`
	IncidentType string   `json:"incident_type,omitempty"`
	IncidentID   string   `json:"incident_id,omitempty"`
	RootCauseID  string   `json:"root_cause_id,omitempty"`
	ActionItems  []string `json:"action_items,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
}

func newRootCausesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show root cause details",
		Long: `Show full details of a root cause.

Output Fields:
  ID              Root cause identifier
  Title           Title
  Description     Description
  Triaged         Whether the root cause is triaged
  Incident        Incident type and ID
  Parent Root Cause  Parent root cause ID
  Action Items    Linked action item IDs
  Created At      Creation timestamp
  Updated At      Last update timestamp

Arguments:
  <id>    Root cause ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a root cause
  xbe view root-causes show 123

  # JSON output
  xbe view root-causes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRootCausesShow,
	}
	initRootCausesShowFlags(cmd)
	return cmd
}

func init() {
	rootCausesCmd.AddCommand(newRootCausesShowCmd())
}

func initRootCausesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRootCausesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRootCausesShowOptions(cmd)
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
		return fmt.Errorf("root cause id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[root-causes]", "title,description,is-triaged,created-at,updated-at,incident,root-cause,action-items")

	body, status, err := client.Get(cmd.Context(), "/v1/root-causes/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderRootCausesShowUnavailable(cmd, opts.JSON)
		}
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

	details := buildRootCauseDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRootCauseDetails(cmd, details)
}

func parseRootCausesShowOptions(cmd *cobra.Command) (rootCausesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rootCausesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRootCauseDetails(resp jsonAPISingleResponse) rootCauseDetails {
	attrs := resp.Data.Attributes
	details := rootCauseDetails{
		ID:          resp.Data.ID,
		Title:       strings.TrimSpace(stringAttr(attrs, "title")),
		Description: strings.TrimSpace(stringAttr(attrs, "description")),
		IsTriaged:   boolAttr(attrs, "is-triaged"),
		CreatedAt:   formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["incident"]; ok && rel.Data != nil {
		details.IncidentType = rel.Data.Type
		details.IncidentID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["root-cause"]; ok && rel.Data != nil {
		details.RootCauseID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["action-items"]; ok && rel.raw != nil {
		ids := relationshipIDs(rel)
		if len(ids) > 0 {
			values := make([]string, 0, len(ids))
			for _, item := range ids {
				values = append(values, formatIncidentReference(item.Type, item.ID))
			}
			details.ActionItems = values
		}
	}

	return details
}

func renderRootCausesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), rootCauseDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Root cause not found.")
	return nil
}

func renderRootCauseDetails(cmd *cobra.Command, details rootCauseDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Title: %s\n", formatOptional(details.Title))
	fmt.Fprintf(out, "Description: %s\n", formatOptional(details.Description))
	fmt.Fprintf(out, "Triaged: %s\n", yesNo(details.IsTriaged))
	fmt.Fprintf(out, "Incident: %s\n", formatOptional(formatIncidentReference(details.IncidentType, details.IncidentID)))
	fmt.Fprintf(out, "Parent Root Cause: %s\n", formatOptional(details.RootCauseID))
	fmt.Fprintf(out, "Action Items: %s\n", formatOptional(strings.Join(details.ActionItems, ", ")))
	fmt.Fprintf(out, "Created At: %s\n", formatOptional(details.CreatedAt))
	fmt.Fprintf(out, "Updated At: %s\n", formatOptional(details.UpdatedAt))

	return nil
}
