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

type driverManagersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverManagerDetails struct {
	ID                  string `json:"id"`
	TruckerID           string `json:"trucker_id,omitempty"`
	ManagerMembershipID string `json:"manager_membership_id,omitempty"`
	ManagedMembershipID string `json:"managed_membership_id,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

func newDriverManagersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver manager details",
		Long: `Show the full details of a driver manager.

Output Fields:
  ID               Driver manager identifier
  Trucker          Trucker ID
  Manager Membership  Manager membership ID
  Managed Membership  Managed membership ID
  Created At       Created timestamp
  Updated At       Updated timestamp

Arguments:
  <id>    The driver manager ID (required). You can find IDs using the list command.`,
		Example: `  # Show a driver manager
  xbe view driver-managers show 123

  # Get JSON output
  xbe view driver-managers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverManagersShow,
	}
	initDriverManagersShowFlags(cmd)
	return cmd
}

func init() {
	driverManagersCmd.AddCommand(newDriverManagersShowCmd())
}

func initDriverManagersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverManagersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseDriverManagersShowOptions(cmd)
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
		return fmt.Errorf("driver manager id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-managers/"+id, nil)
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

	details := buildDriverManagerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverManagerDetails(cmd, details)
}

func parseDriverManagersShowOptions(cmd *cobra.Command) (driverManagersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverManagersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverManagerDetails(resp jsonAPISingleResponse) driverManagerDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := driverManagerDetails{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["manager-membership"]; ok && rel.Data != nil {
		details.ManagerMembershipID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["managed-membership"]; ok && rel.Data != nil {
		details.ManagedMembershipID = rel.Data.ID
	}

	return details
}

func renderDriverManagerDetails(cmd *cobra.Command, details driverManagerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if details.ManagerMembershipID != "" {
		fmt.Fprintf(out, "Manager Membership: %s\n", details.ManagerMembershipID)
	}
	if details.ManagedMembershipID != "" {
		fmt.Fprintf(out, "Managed Membership: %s\n", details.ManagedMembershipID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
