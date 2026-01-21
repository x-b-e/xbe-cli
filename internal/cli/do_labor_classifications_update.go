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

type doLaborClassificationsUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	Name               string
	Abbreviation       string
	MobilizationMethod string
	IsTimeCardApprover bool
	AllowColors        bool
	IsManager          bool
	CanManageProjects  bool
}

func newDoLaborClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a labor classification",
		Long: `Update an existing labor classification.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The labor classification ID (required)

Flags:
  --name                   Update the name
  --abbreviation           Update the abbreviation
  --mobilization-method    Update the mobilization method
  --is-time-card-approver  Update time card approver status
  --allow-colors           Update color assignment permission
  --is-manager             Update manager status
  --can-manage-projects    Update project management permission`,
		Example: `  # Update just the name
  xbe do labor-classifications update 123 --name "Senior Raker"

  # Update multiple fields
  xbe do labor-classifications update 123 --name "Foreman" --is-manager

  # Get JSON output
  xbe do labor-classifications update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLaborClassificationsUpdate,
	}
	initDoLaborClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doLaborClassificationsCmd.AddCommand(newDoLaborClassificationsUpdateCmd())
}

func initDoLaborClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("abbreviation", "", "New abbreviation")
	cmd.Flags().String("mobilization-method", "", "New mobilization method")
	cmd.Flags().Bool("is-time-card-approver", false, "Can approve time cards")
	cmd.Flags().Bool("allow-colors", false, "Allow color assignments")
	cmd.Flags().Bool("is-manager", false, "Is a manager role")
	cmd.Flags().Bool("can-manage-projects", false, "Can manage projects")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLaborClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLaborClassificationsUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("labor classification id is required")
	}

	// Check if at least one field is being updated
	hasUpdate := opts.Name != "" || opts.Abbreviation != "" || opts.MobilizationMethod != "" ||
		cmd.Flags().Changed("is-time-card-approver") ||
		cmd.Flags().Changed("allow-colors") ||
		cmd.Flags().Changed("is-manager") ||
		cmd.Flags().Changed("can-manage-projects")

	if !hasUpdate {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{}
	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Abbreviation != "" {
		attributes["abbreviation"] = opts.Abbreviation
	}
	if opts.MobilizationMethod != "" {
		attributes["mobilization-method"] = opts.MobilizationMethod
	}
	if cmd.Flags().Changed("is-time-card-approver") {
		attributes["is-time-card-approver"] = opts.IsTimeCardApprover
	}
	if cmd.Flags().Changed("allow-colors") {
		attributes["allow-colors"] = opts.AllowColors
	}
	if cmd.Flags().Changed("is-manager") {
		attributes["is-manager"] = opts.IsManager
	}
	if cmd.Flags().Changed("can-manage-projects") {
		attributes["can-manage-projects"] = opts.CanManageProjects
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "labor-classifications",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/labor-classifications/"+id, jsonBody)
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

	row := buildLaborClassificationRow(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated labor classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoLaborClassificationsUpdateOptions(cmd *cobra.Command) (doLaborClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	abbreviation, _ := cmd.Flags().GetString("abbreviation")
	mobilizationMethod, _ := cmd.Flags().GetString("mobilization-method")
	isTimeCardApprover, _ := cmd.Flags().GetBool("is-time-card-approver")
	allowColors, _ := cmd.Flags().GetBool("allow-colors")
	isManager, _ := cmd.Flags().GetBool("is-manager")
	canManageProjects, _ := cmd.Flags().GetBool("can-manage-projects")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLaborClassificationsUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		Name:               name,
		Abbreviation:       abbreviation,
		MobilizationMethod: mobilizationMethod,
		IsTimeCardApprover: isTimeCardApprover,
		AllowColors:        allowColors,
		IsManager:          isManager,
		CanManageProjects:  canManageProjects,
	}, nil
}
