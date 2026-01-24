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

type doProjectTransportOrganizationsUpdateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ID                         string
	Name                       string
	ExternalTmsMasterCompanyID string
}

func newDoProjectTransportOrganizationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport organization",
		Long: `Update a project transport organization.

Provide the organization ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name                          Organization name
  --external-tms-master-company-id External TMS master company identifier

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a project transport organization name
  xbe do project-transport-organizations update 123 --name "Acme Transport West"

  # Update the external TMS master company ID
  xbe do project-transport-organizations update 123 --external-tms-master-company-id "TMS-002"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportOrganizationsUpdate,
	}
	initDoProjectTransportOrganizationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportOrganizationsCmd.AddCommand(newDoProjectTransportOrganizationsUpdateCmd())
}

func initDoProjectTransportOrganizationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Organization name")
	cmd.Flags().String("external-tms-master-company-id", "", "External TMS master company identifier")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportOrganizationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportOrganizationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("external-tms-master-company-id") {
		attributes["external-tms-master-company-id"] = opts.ExternalTmsMasterCompanyID
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --external-tms-master-company-id")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "project-transport-organizations",
		"id":         opts.ID,
		"attributes": attributes,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-organizations/"+opts.ID, jsonBody)
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

	rows := buildProjectTransportOrganizationRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
	if opts.JSON {
		if len(rows) > 0 {
			return writeJSON(cmd.OutOrStdout(), rows[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	name := stringAttr(resp.Data.Attributes, "name")
	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport organization %s (%s)\n", resp.Data.ID, name)
	return nil
}

func parseDoProjectTransportOrganizationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportOrganizationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	externalTmsMasterCompanyID, _ := cmd.Flags().GetString("external-tms-master-company-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportOrganizationsUpdateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ID:                         args[0],
		Name:                       name,
		ExternalTmsMasterCompanyID: externalTmsMasterCompanyID,
	}, nil
}
