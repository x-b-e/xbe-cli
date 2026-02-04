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

type objectiveChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type objectiveChangeDetails struct {
	ID               string `json:"id"`
	ObjectiveID      string `json:"objective_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	ChangedByID      string `json:"changed_by_id,omitempty"`
	StartOnOld       string `json:"start_on_old,omitempty"`
	StartOnNew       string `json:"start_on_new,omitempty"`
	EndOnOld         string `json:"end_on_old,omitempty"`
	EndOnNew         string `json:"end_on_new,omitempty"`
}

func newObjectiveChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show objective change details",
		Long: `Show the full details of an objective change.

Output Fields:
  ID            Objective change identifier
  OBJECTIVE     Objective ID
  ORGANIZATION  Organization (Type/ID)
  BROKER        Broker ID
  START OLD     Previous start date
  START NEW     Updated start date
  END OLD       Previous end date
  END NEW       Updated end date
  CHANGED BY    User who made the change

Arguments:
  <id>  Objective change ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an objective change
  xbe view objective-changes show 123

  # Output as JSON
  xbe view objective-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runObjectiveChangesShow,
	}
	initObjectiveChangesShowFlags(cmd)
	return cmd
}

func init() {
	objectiveChangesCmd.AddCommand(newObjectiveChangesShowCmd())
}

func initObjectiveChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectiveChangesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseObjectiveChangesShowOptions(cmd)
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
		return fmt.Errorf("objective change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[objective-changes]", "objective,start-on-old,start-on-new,end-on-old,end-on-new,organization,broker,changed-by")

	body, _, err := client.Get(cmd.Context(), "/v1/objective-changes/"+id, query)
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

	details := buildObjectiveChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderObjectiveChangeDetails(cmd, details)
}

func parseObjectiveChangesShowOptions(cmd *cobra.Command) (objectiveChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectiveChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildObjectiveChangeDetails(resp jsonAPISingleResponse) objectiveChangeDetails {
	attrs := resp.Data.Attributes
	details := objectiveChangeDetails{
		ID:         resp.Data.ID,
		StartOnOld: formatDate(stringAttr(attrs, "start-on-old")),
		StartOnNew: formatDate(stringAttr(attrs, "start-on-new")),
		EndOnOld:   formatDate(stringAttr(attrs, "end-on-old")),
		EndOnNew:   formatDate(stringAttr(attrs, "end-on-new")),
	}

	if rel, ok := resp.Data.Relationships["objective"]; ok && rel.Data != nil {
		details.ObjectiveID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
	}

	return details
}

func renderObjectiveChangeDetails(cmd *cobra.Command, details objectiveChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ObjectiveID != "" {
		fmt.Fprintf(out, "Objective: %s\n", details.ObjectiveID)
	}
	organization := formatPolymorphic(details.OrganizationType, details.OrganizationID)
	if organization != "" {
		fmt.Fprintf(out, "Organization: %s\n", organization)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByID)
	}
	if details.StartOnOld != "" {
		fmt.Fprintf(out, "Start On Old: %s\n", details.StartOnOld)
	}
	if details.StartOnNew != "" {
		fmt.Fprintf(out, "Start On New: %s\n", details.StartOnNew)
	}
	if details.EndOnOld != "" {
		fmt.Fprintf(out, "End On Old: %s\n", details.EndOnOld)
	}
	if details.EndOnNew != "" {
		fmt.Fprintf(out, "End On New: %s\n", details.EndOnNew)
	}

	return nil
}
