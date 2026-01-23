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

type doJobProductionPlanMaterialSitesCreateOptions struct {
	BaseURL                                   string
	Token                                     string
	JSON                                      bool
	JobProductionPlanID                       string
	MaterialSiteID                            string
	IsDefault                                 bool
	Miles                                     string
	DefaultTicketMaker                        string
	UserTicketMakerMaterialTypeIDs            []string
	HasUserScaleTickets                       bool
	PlanRequiresSiteSpecificMaterialTypes     bool
	PlanRequiresSupplierSpecificMaterialTypes bool
}

func newDoJobProductionPlanMaterialSitesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan material site",
		Long: `Create a job production plan material site.

Required flags:
  --job-production-plan   Job production plan ID
  --material-site         Material site ID

Optional flags:
  --is-default                                 Set as the default material site
  --miles                                      Planned miles between job site and material site
  --default-ticket-maker                       Default ticket maker (user, material_site)
  --user-ticket-maker-material-type-ids        Material type IDs where user is ticket maker (comma-separated or repeated)
  --has-user-scale-tickets                      Whether users have scale tickets
  --plan-requires-site-specific-material-types  Plan requires site-specific material types
  --plan-requires-supplier-specific-material-types Plan requires supplier-specific material types`,
		Example: `  # Create a job production plan material site
  xbe do job-production-plan-material-sites create \\
    --job-production-plan 123 \\
    --material-site 456

  # Create with planned miles and ticket maker
  xbe do job-production-plan-material-sites create \\
    --job-production-plan 123 \\
    --material-site 456 \\
    --miles 12.5 \\
    --default-ticket-maker material_site`,
		RunE: runDoJobProductionPlanMaterialSitesCreate,
	}
	initDoJobProductionPlanMaterialSitesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialSitesCmd.AddCommand(newDoJobProductionPlanMaterialSitesCreateCmd())
}

func initDoJobProductionPlanMaterialSitesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().Bool("is-default", false, "Set as the default material site")
	cmd.Flags().String("miles", "", "Planned miles between job site and material site")
	cmd.Flags().String("default-ticket-maker", "", "Default ticket maker (user, material_site)")
	cmd.Flags().StringSlice("user-ticket-maker-material-type-ids", nil, "Material type IDs where user is ticket maker (comma-separated or repeated)")
	cmd.Flags().Bool("has-user-scale-tickets", false, "Whether users have scale tickets")
	cmd.Flags().Bool("plan-requires-site-specific-material-types", false, "Plan requires site-specific material types")
	cmd.Flags().Bool("plan-requires-supplier-specific-material-types", false, "Plan requires supplier-specific material types")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("job-production-plan")
	cmd.MarkFlagRequired("material-site")
}

func runDoJobProductionPlanMaterialSitesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanMaterialSitesCreateOptions(cmd)
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
	if cmd.Flags().Changed("is-default") {
		attributes["is-default"] = opts.IsDefault
	}
	if cmd.Flags().Changed("miles") {
		attributes["miles"] = opts.Miles
	}
	if cmd.Flags().Changed("default-ticket-maker") {
		attributes["default-ticket-maker"] = opts.DefaultTicketMaker
	}
	if cmd.Flags().Changed("user-ticket-maker-material-type-ids") {
		values := cleanStringSlice(opts.UserTicketMakerMaterialTypeIDs)
		if values == nil {
			attributes["user-ticket-maker-material-type-ids"] = []string{}
		} else {
			attributes["user-ticket-maker-material-type-ids"] = values
		}
	}
	if cmd.Flags().Changed("has-user-scale-tickets") {
		attributes["has-user-scale-tickets"] = opts.HasUserScaleTickets
	}
	if cmd.Flags().Changed("plan-requires-site-specific-material-types") {
		attributes["plan-requires-site-specific-material-types"] = opts.PlanRequiresSiteSpecificMaterialTypes
	}
	if cmd.Flags().Changed("plan-requires-supplier-specific-material-types") {
		attributes["plan-requires-supplier-specific-material-types"] = opts.PlanRequiresSupplierSpecificMaterialTypes
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSiteID,
			},
		},
	}

	data := map[string]any{
		"type":          "job-production-plan-material-sites",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-material-sites", jsonBody)
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

	row := jobProductionPlanMaterialSiteRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan material site %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanMaterialSitesCreateOptions(cmd *cobra.Command) (doJobProductionPlanMaterialSitesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	materialSiteID, _ := cmd.Flags().GetString("material-site")
	isDefault, _ := cmd.Flags().GetBool("is-default")
	miles, _ := cmd.Flags().GetString("miles")
	defaultTicketMaker, _ := cmd.Flags().GetString("default-ticket-maker")
	userTicketMakerMaterialTypeIDs, _ := cmd.Flags().GetStringSlice("user-ticket-maker-material-type-ids")
	hasUserScaleTickets, _ := cmd.Flags().GetBool("has-user-scale-tickets")
	planRequiresSiteSpecificMaterialTypes, _ := cmd.Flags().GetBool("plan-requires-site-specific-material-types")
	planRequiresSupplierSpecificMaterialTypes, _ := cmd.Flags().GetBool("plan-requires-supplier-specific-material-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanMaterialSitesCreateOptions{
		BaseURL:                               baseURL,
		Token:                                 token,
		JSON:                                  jsonOut,
		JobProductionPlanID:                   jobProductionPlanID,
		MaterialSiteID:                        materialSiteID,
		IsDefault:                             isDefault,
		Miles:                                 miles,
		DefaultTicketMaker:                    defaultTicketMaker,
		UserTicketMakerMaterialTypeIDs:        userTicketMakerMaterialTypeIDs,
		HasUserScaleTickets:                   hasUserScaleTickets,
		PlanRequiresSiteSpecificMaterialTypes: planRequiresSiteSpecificMaterialTypes,
		PlanRequiresSupplierSpecificMaterialTypes: planRequiresSupplierSpecificMaterialTypes,
	}, nil
}
