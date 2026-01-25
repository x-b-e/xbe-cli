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

type doJobProductionPlanMaterialSitesUpdateOptions struct {
	BaseURL                                     string
	Token                                       string
	JSON                                        bool
	ID                                          string
	Miles                                       string
	DefaultTicketMaker                          string
	UserTicketMakerMaterialTypeIDs              []string
	IsDefault                                   bool
	NoIsDefault                                 bool
	HasUserScaleTickets                         bool
	NoHasUserScaleTickets                       bool
	PlanRequiresSiteSpecificMaterialTypes       bool
	NoPlanRequiresSiteSpecificMaterialTypes     bool
	PlanRequiresSupplierSpecificMaterialTypes   bool
	NoPlanRequiresSupplierSpecificMaterialTypes bool
}

func newDoJobProductionPlanMaterialSitesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan material site",
		Long: `Update a job production plan material site.

Optional flags:
  --is-default                                  Set as the default material site
  --no-is-default                               Unset default material site
  --miles                                       Planned miles between job site and material site
  --default-ticket-maker                        Default ticket maker (user, material_site)
  --user-ticket-maker-material-type-ids         Material type IDs where user is ticket maker (comma-separated or repeated)
  --has-user-scale-tickets                       Enable user scale tickets
  --no-has-user-scale-tickets                    Disable user scale tickets
  --plan-requires-site-specific-material-types   Enable site-specific material types requirement
  --no-plan-requires-site-specific-material-types Disable site-specific material types requirement
  --plan-requires-supplier-specific-material-types Enable supplier-specific material types requirement
  --no-plan-requires-supplier-specific-material-types Disable supplier-specific material types requirement`,
		Example: `  # Update planned miles
  xbe do job-production-plan-material-sites update 123 --miles 15

  # Set as default material site
  xbe do job-production-plan-material-sites update 123 --is-default

  # Update ticket maker material types
  xbe do job-production-plan-material-sites update 123 --user-ticket-maker-material-type-ids 45,67`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanMaterialSitesUpdate,
	}
	initDoJobProductionPlanMaterialSitesUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanMaterialSitesCmd.AddCommand(newDoJobProductionPlanMaterialSitesUpdateCmd())
}

func initDoJobProductionPlanMaterialSitesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-default", false, "Set as the default material site")
	cmd.Flags().Bool("no-is-default", false, "Unset default material site")
	cmd.Flags().String("miles", "", "Planned miles between job site and material site")
	cmd.Flags().String("default-ticket-maker", "", "Default ticket maker (user, material_site)")
	cmd.Flags().StringSlice("user-ticket-maker-material-type-ids", nil, "Material type IDs where user is ticket maker (comma-separated or repeated)")
	cmd.Flags().Bool("has-user-scale-tickets", false, "Enable user scale tickets")
	cmd.Flags().Bool("no-has-user-scale-tickets", false, "Disable user scale tickets")
	cmd.Flags().Bool("plan-requires-site-specific-material-types", false, "Enable site-specific material types requirement")
	cmd.Flags().Bool("no-plan-requires-site-specific-material-types", false, "Disable site-specific material types requirement")
	cmd.Flags().Bool("plan-requires-supplier-specific-material-types", false, "Enable supplier-specific material types requirement")
	cmd.Flags().Bool("no-plan-requires-supplier-specific-material-types", false, "Disable supplier-specific material types requirement")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanMaterialSitesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanMaterialSitesUpdateOptions(cmd, args)
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
		attributes["is-default"] = true
	}
	if cmd.Flags().Changed("no-is-default") {
		attributes["is-default"] = false
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
		attributes["has-user-scale-tickets"] = true
	}
	if cmd.Flags().Changed("no-has-user-scale-tickets") {
		attributes["has-user-scale-tickets"] = false
	}
	if cmd.Flags().Changed("plan-requires-site-specific-material-types") {
		attributes["plan-requires-site-specific-material-types"] = true
	}
	if cmd.Flags().Changed("no-plan-requires-site-specific-material-types") {
		attributes["plan-requires-site-specific-material-types"] = false
	}
	if cmd.Flags().Changed("plan-requires-supplier-specific-material-types") {
		attributes["plan-requires-supplier-specific-material-types"] = true
	}
	if cmd.Flags().Changed("no-plan-requires-supplier-specific-material-types") {
		attributes["plan-requires-supplier-specific-material-types"] = false
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "job-production-plan-material-sites",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-material-sites/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan material site %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanMaterialSitesUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanMaterialSitesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isDefault, _ := cmd.Flags().GetBool("is-default")
	noIsDefault, _ := cmd.Flags().GetBool("no-is-default")
	miles, _ := cmd.Flags().GetString("miles")
	defaultTicketMaker, _ := cmd.Flags().GetString("default-ticket-maker")
	userTicketMakerMaterialTypeIDs, _ := cmd.Flags().GetStringSlice("user-ticket-maker-material-type-ids")
	hasUserScaleTickets, _ := cmd.Flags().GetBool("has-user-scale-tickets")
	noHasUserScaleTickets, _ := cmd.Flags().GetBool("no-has-user-scale-tickets")
	planRequiresSiteSpecificMaterialTypes, _ := cmd.Flags().GetBool("plan-requires-site-specific-material-types")
	noPlanRequiresSiteSpecificMaterialTypes, _ := cmd.Flags().GetBool("no-plan-requires-site-specific-material-types")
	planRequiresSupplierSpecificMaterialTypes, _ := cmd.Flags().GetBool("plan-requires-supplier-specific-material-types")
	noPlanRequiresSupplierSpecificMaterialTypes, _ := cmd.Flags().GetBool("no-plan-requires-supplier-specific-material-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanMaterialSitesUpdateOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		ID:                                      strings.TrimSpace(args[0]),
		Miles:                                   miles,
		DefaultTicketMaker:                      defaultTicketMaker,
		UserTicketMakerMaterialTypeIDs:          userTicketMakerMaterialTypeIDs,
		IsDefault:                               isDefault,
		NoIsDefault:                             noIsDefault,
		HasUserScaleTickets:                     hasUserScaleTickets,
		NoHasUserScaleTickets:                   noHasUserScaleTickets,
		PlanRequiresSiteSpecificMaterialTypes:   planRequiresSiteSpecificMaterialTypes,
		NoPlanRequiresSiteSpecificMaterialTypes: noPlanRequiresSiteSpecificMaterialTypes,
		PlanRequiresSupplierSpecificMaterialTypes:   planRequiresSupplierSpecificMaterialTypes,
		NoPlanRequiresSupplierSpecificMaterialTypes: noPlanRequiresSupplierSpecificMaterialTypes,
	}, nil
}
