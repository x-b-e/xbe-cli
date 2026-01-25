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

type doProjectTransportOrganizationsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	Name                       string
	ExternalTmsMasterCompanyID string
	Broker                     string
}

func newDoProjectTransportOrganizationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport organization",
		Long: `Create a project transport organization.

Required flags:
  --name    Organization name
  --broker  Broker ID

Optional flags:
  --external-tms-master-company-id  External TMS master company identifier

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project transport organization
  xbe do project-transport-organizations create --name "Acme Transport" --broker 123

  # Create with external TMS master company ID
  xbe do project-transport-organizations create --name "Acme Transport" --broker 123 \
    --external-tms-master-company-id "TMS-001"`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportOrganizationsCreate,
	}
	initDoProjectTransportOrganizationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportOrganizationsCmd.AddCommand(newDoProjectTransportOrganizationsCreateCmd())
}

func initDoProjectTransportOrganizationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Organization name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("external-tms-master-company-id", "", "External TMS master company identifier")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("broker")
}

func runDoProjectTransportOrganizationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportOrganizationsCreateOptions(cmd)
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

	attributes := map[string]any{
		"name": opts.Name,
	}
	if strings.TrimSpace(opts.ExternalTmsMasterCompanyID) != "" {
		attributes["external-tms-master-company-id"] = opts.ExternalTmsMasterCompanyID
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	data := map[string]any{
		"type":          "project-transport-organizations",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-organizations", jsonBody)
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

	if opts.JSON {
		rows := buildProjectTransportOrganizationRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
		if len(rows) > 0 {
			return writeJSON(cmd.OutOrStdout(), rows[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport organization %s (%s)\n", resp.Data.ID, stringAttr(resp.Data.Attributes, "name"))
	return nil
}

func parseDoProjectTransportOrganizationsCreateOptions(cmd *cobra.Command) (doProjectTransportOrganizationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	externalTmsMasterCompanyID, _ := cmd.Flags().GetString("external-tms-master-company-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportOrganizationsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		Name:                       name,
		Broker:                     broker,
		ExternalTmsMasterCompanyID: externalTmsMasterCompanyID,
	}, nil
}
