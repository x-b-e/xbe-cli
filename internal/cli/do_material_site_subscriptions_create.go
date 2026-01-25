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

type doMaterialSiteSubscriptionsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	UserID        string
	MaterialSite  string
	ContactMethod string
}

func newDoMaterialSiteSubscriptionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material site subscription",
		Long: `Create a material site subscription.

Required flags:
  --user           User ID
  --material-site  Material site ID

Optional flags:
  --contact-method  Contact method (email_address, mobile_number)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a subscription
  xbe do material-site-subscriptions create --user 123 --material-site 456 --contact-method email_address

  # Create using default contact method
  xbe do material-site-subscriptions create --user 123 --material-site 456`,
		RunE: runDoMaterialSiteSubscriptionsCreate,
	}
	initDoMaterialSiteSubscriptionsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteSubscriptionsCmd.AddCommand(newDoMaterialSiteSubscriptionsCreateCmd())
}

func initDoMaterialSiteSubscriptionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("contact-method", "", "Contact method (email_address, mobile_number)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("material-site")
}

func runDoMaterialSiteSubscriptionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialSiteSubscriptionsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	attributes := map[string]any{}
	if opts.ContactMethod != "" {
		attributes["contact-method"] = opts.ContactMethod
	}

	relationships := map[string]any{
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.UserID,
			},
		},
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		},
	}

	data := map[string]any{
		"type":          "material-site-subscriptions",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-site-subscriptions", jsonBody)
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

	row := materialSiteSubscriptionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created material site subscription %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteSubscriptionsCreateOptions(cmd *cobra.Command) (doMaterialSiteSubscriptionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	userID, _ := cmd.Flags().GetString("user")
	materialSite, _ := cmd.Flags().GetString("material-site")
	contactMethod, _ := cmd.Flags().GetString("contact-method")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteSubscriptionsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		UserID:        userID,
		MaterialSite:  materialSite,
		ContactMethod: contactMethod,
	}, nil
}

func materialSiteSubscriptionRowFromSingle(resp jsonAPISingleResponse) materialSiteSubscriptionRow {
	resource := resp.Data
	attrs := resource.Attributes

	row := materialSiteSubscriptionRow{
		ID:                      resource.ID,
		ContactMethod:           stringAttr(attrs, "contact-method"),
		CalculatedContactMethod: stringAttr(attrs, "calculated-contact-method"),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
	}

	return row
}
