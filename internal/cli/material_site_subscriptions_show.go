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

type materialSiteSubscriptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialSiteSubscriptionDetails struct {
	ID                      string `json:"id"`
	ContactMethod           string `json:"contact_method,omitempty"`
	CalculatedContactMethod string `json:"calculated_contact_method,omitempty"`
	UserID                  string `json:"user_id,omitempty"`
	User                    string `json:"user,omitempty"`
	UserEmail               string `json:"user_email,omitempty"`
	UserMobile              string `json:"user_mobile,omitempty"`
	MaterialSiteID          string `json:"material_site_id,omitempty"`
	MaterialSite            string `json:"material_site,omitempty"`
}

func newMaterialSiteSubscriptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material site subscription details",
		Long: `Show the full details of a material site subscription.

Arguments:
  <id>  Material site subscription ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a subscription
  xbe view material-site-subscriptions show 123

  # JSON output
  xbe view material-site-subscriptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialSiteSubscriptionsShow,
	}
	initMaterialSiteSubscriptionsShowFlags(cmd)
	return cmd
}

func init() {
	materialSiteSubscriptionsCmd.AddCommand(newMaterialSiteSubscriptionsShowCmd())
}

func initMaterialSiteSubscriptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteSubscriptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialSiteSubscriptionsShowOptions(cmd)
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
		return fmt.Errorf("material site subscription id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-site-subscriptions]", "contact-method,calculated-contact-method,user,material-site")
	query.Set("fields[users]", "name,email-address,mobile-number")
	query.Set("fields[material-sites]", "name")
	query.Set("include", "user,material-site")

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-subscriptions/"+id, query)
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

	details := buildMaterialSiteSubscriptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialSiteSubscriptionDetails(cmd, details)
}

func parseMaterialSiteSubscriptionsShowOptions(cmd *cobra.Command) (materialSiteSubscriptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteSubscriptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialSiteSubscriptionDetails(resp jsonAPISingleResponse) materialSiteSubscriptionDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialSiteSubscriptionDetails{
		ID:                      resource.ID,
		ContactMethod:           stringAttr(attrs, "contact-method"),
		CalculatedContactMethod: stringAttr(attrs, "calculated-contact-method"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.User = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
			details.UserMobile = stringAttr(user.Attributes, "mobile-number")
		}
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSite = stringAttr(site.Attributes, "name")
		}
	}

	return details
}

func renderMaterialSiteSubscriptionDetails(cmd *cobra.Command, details materialSiteSubscriptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ContactMethod != "" {
		fmt.Fprintf(out, "Contact Method: %s\n", details.ContactMethod)
	}
	if details.CalculatedContactMethod != "" {
		fmt.Fprintf(out, "Calculated Contact Method: %s\n", details.CalculatedContactMethod)
	}
	if details.User != "" {
		fmt.Fprintf(out, "User: %s\n", details.User)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.UserMobile != "" {
		fmt.Fprintf(out, "User Mobile: %s\n", details.UserMobile)
	}
	if details.MaterialSite != "" {
		fmt.Fprintf(out, "Material Site: %s\n", details.MaterialSite)
	}
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site ID: %s\n", details.MaterialSiteID)
	}

	return nil
}
