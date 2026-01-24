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

type businessUnitMembershipsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type businessUnitMembershipDetails struct {
	ID               string `json:"id"`
	Kind             string `json:"kind,omitempty"`
	BusinessUnitID   string `json:"business_unit_id,omitempty"`
	BusinessUnitName string `json:"business_unit_name,omitempty"`
	MembershipID     string `json:"membership_id,omitempty"`
	MembershipUserID string `json:"membership_user_id,omitempty"`
}

func newBusinessUnitMembershipsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show business unit membership details",
		Long: `Show the full details of a business unit membership.

Output Fields:
  ID
  Kind
  Business Unit
  Membership
  Membership User

Arguments:
  <id>    The business unit membership ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a business unit membership
  xbe view business-unit-memberships show 123

  # Get JSON output
  xbe view business-unit-memberships show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBusinessUnitMembershipsShow,
	}
	initBusinessUnitMembershipsShowFlags(cmd)
	return cmd
}

func init() {
	businessUnitMembershipsCmd.AddCommand(newBusinessUnitMembershipsShowCmd())
}

func initBusinessUnitMembershipsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitMembershipsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBusinessUnitMembershipsShowOptions(cmd)
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
		return fmt.Errorf("business unit membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[business-unit-memberships]", "kind,business-unit,membership")
	query.Set("fields[business-units]", "company-name")
	query.Set("fields[memberships]", "kind,user")
	query.Set("include", "business-unit,membership")

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-memberships/"+id, query)
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

	details := buildBusinessUnitMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBusinessUnitMembershipDetails(cmd, details)
}

func parseBusinessUnitMembershipsShowOptions(cmd *cobra.Command) (businessUnitMembershipsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitMembershipsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBusinessUnitMembershipDetails(resp jsonAPISingleResponse) businessUnitMembershipDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := businessUnitMembershipDetails{
		ID:   resource.ID,
		Kind: stringAttr(attrs, "kind"),
	}

	if rel, ok := resource.Relationships["business-unit"]; ok && rel.Data != nil {
		details.BusinessUnitID = rel.Data.ID
		if businessUnit, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BusinessUnitName = stringAttr(businessUnit.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["membership"]; ok && rel.Data != nil {
		details.MembershipID = rel.Data.ID
	}

	if details.MembershipID != "" {
		if membership, ok := included[resourceKey("memberships", details.MembershipID)]; ok {
			details.MembershipUserID = relationshipIDFromMap(membership.Relationships, "user")
		}
	}

	return details
}

func renderBusinessUnitMembershipDetails(cmd *cobra.Command, details businessUnitMembershipDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.BusinessUnitID != "" {
		label := details.BusinessUnitID
		if details.BusinessUnitName != "" {
			label = fmt.Sprintf("%s (%s)", details.BusinessUnitName, details.BusinessUnitID)
		}
		fmt.Fprintf(out, "Business Unit: %s\n", label)
	}
	if details.MembershipID != "" {
		fmt.Fprintf(out, "Membership ID: %s\n", details.MembershipID)
	}
	if details.MembershipUserID != "" {
		fmt.Fprintf(out, "Membership User ID: %s\n", details.MembershipUserID)
	}

	return nil
}
