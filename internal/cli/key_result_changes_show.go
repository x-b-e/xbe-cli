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

type keyResultChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type keyResultChangeDetails struct {
	ID               string `json:"id"`
	KeyResultID      string `json:"key_result_id,omitempty"`
	ObjectiveID      string `json:"objective_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	ChangedByID      string `json:"changed_by_id,omitempty"`
	StartOnOld       string `json:"start_on_old,omitempty"`
	StartOnNew       string `json:"start_on_new,omitempty"`
	EndOnOld         string `json:"end_on_old,omitempty"`
	EndOnNew         string `json:"end_on_new,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

func newKeyResultChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show key result change details",
		Long: `Show the full details of a key result change.

Output Fields:
  ID
  Key Result ID
  Objective ID
  Broker ID
  Organization Type
  Organization ID
  Changed By ID
  Start On Old
  Start On New
  End On Old
  End On New
  Created At
  Updated At

Arguments:
  <id>    The key result change ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show key result change details
  xbe view key-result-changes show 123

  # Get JSON output
  xbe view key-result-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runKeyResultChangesShow,
	}
	initKeyResultChangesShowFlags(cmd)
	return cmd
}

func init() {
	keyResultChangesCmd.AddCommand(newKeyResultChangesShowCmd())
}

func initKeyResultChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runKeyResultChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseKeyResultChangesShowOptions(cmd)
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
		return fmt.Errorf("key result change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[key-result-changes]", "start-on-old,start-on-new,end-on-old,end-on-new,created-at,updated-at,key-result,objective,broker,organization,changed-by")

	body, _, err := client.Get(cmd.Context(), "/v1/key-result-changes/"+id, query)
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

	details := buildKeyResultChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderKeyResultChangeDetails(cmd, details)
}

func parseKeyResultChangesShowOptions(cmd *cobra.Command) (keyResultChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return keyResultChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildKeyResultChangeDetails(resp jsonAPISingleResponse) keyResultChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := keyResultChangeDetails{
		ID:         resource.ID,
		StartOnOld: formatDate(stringAttr(attrs, "start-on-old")),
		StartOnNew: formatDate(stringAttr(attrs, "start-on-new")),
		EndOnOld:   formatDate(stringAttr(attrs, "end-on-old")),
		EndOnNew:   formatDate(stringAttr(attrs, "end-on-new")),
		CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	details.KeyResultID = relationshipIDFromMap(resource.Relationships, "key-result")
	details.ObjectiveID = relationshipIDFromMap(resource.Relationships, "objective")
	details.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
	details.ChangedByID = relationshipIDFromMap(resource.Relationships, "changed-by")

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationID = rel.Data.ID
		details.OrganizationType = rel.Data.Type
	}

	return details
}

func renderKeyResultChangeDetails(cmd *cobra.Command, details keyResultChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.KeyResultID != "" {
		fmt.Fprintf(out, "Key Result ID: %s\n", details.KeyResultID)
	}
	if details.ObjectiveID != "" {
		fmt.Fprintf(out, "Objective ID: %s\n", details.ObjectiveID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.OrganizationType != "" {
		fmt.Fprintf(out, "Organization Type: %s\n", details.OrganizationType)
	}
	if details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization ID: %s\n", details.OrganizationID)
	}
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By ID: %s\n", details.ChangedByID)
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
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
