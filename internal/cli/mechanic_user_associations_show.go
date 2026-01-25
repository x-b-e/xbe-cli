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

type mechanicUserAssociationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type mechanicUserAssociationDetails struct {
	ID                       string `json:"id"`
	UserID                   string `json:"user_id,omitempty"`
	MaintenanceRequirementID string `json:"maintenance_requirement_id,omitempty"`
}

func newMechanicUserAssociationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show mechanic user association details",
		Long: `Show the full details of a mechanic user association.

Output Fields:
  ID
  User ID
  Maintenance Requirement ID

Arguments:
  <id>    The mechanic user association ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a mechanic user association
  xbe view mechanic-user-associations show 123

  # JSON output
  xbe view mechanic-user-associations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMechanicUserAssociationsShow,
	}
	initMechanicUserAssociationsShowFlags(cmd)
	return cmd
}

func init() {
	mechanicUserAssociationsCmd.AddCommand(newMechanicUserAssociationsShowCmd())
}

func initMechanicUserAssociationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMechanicUserAssociationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMechanicUserAssociationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("mechanic user association id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[mechanic-user-associations]", "user,maintenance-requirement")

	body, _, err := client.Get(cmd.Context(), "/v1/mechanic-user-associations/"+id, query)
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

	details := buildMechanicUserAssociationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMechanicUserAssociationDetails(cmd, details)
}

func parseMechanicUserAssociationsShowOptions(cmd *cobra.Command) (mechanicUserAssociationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return mechanicUserAssociationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMechanicUserAssociationDetails(resp jsonAPISingleResponse) mechanicUserAssociationDetails {
	resource := resp.Data
	details := mechanicUserAssociationDetails{ID: resource.ID}

	details.UserID = relationshipIDFromMap(resource.Relationships, "user")
	details.MaintenanceRequirementID = relationshipIDFromMap(resource.Relationships, "maintenance-requirement")

	return details
}

func renderMechanicUserAssociationDetails(cmd *cobra.Command, details mechanicUserAssociationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.MaintenanceRequirementID != "" {
		fmt.Fprintf(out, "Maintenance Requirement ID: %s\n", details.MaintenanceRequirementID)
	}
	return nil
}
