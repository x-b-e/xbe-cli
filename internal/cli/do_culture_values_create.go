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

type doCultureValuesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Name         string
	Description  string
	Organization string
}

func newDoCultureValuesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new culture value",
		Long: `Create a new culture value.

Required flags:
  --name          The culture value name (required)
  --organization  Organization in Type|ID format (required, e.g. Broker|123)

Optional flags:
  --description   Description text`,
		Example: `  # Create a culture value for a broker
  xbe do culture-values create --name "Integrity" --organization Broker|123

  # Create with description
  xbe do culture-values create --name "Excellence" --description "Strive for excellence in all we do" --organization Broker|123

  # Get JSON output
  xbe do culture-values create --name "Innovation" --organization Broker|123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCultureValuesCreate,
	}
	initDoCultureValuesCreateFlags(cmd)
	return cmd
}

func init() {
	doCultureValuesCmd.AddCommand(newDoCultureValuesCreateCmd())
}

func initDoCultureValuesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Culture value name (required)")
	cmd.Flags().String("description", "", "Description text")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (required, e.g. Broker|123)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCultureValuesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCultureValuesCreateOptions(cmd)
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

	// Require name
	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require organization
	if opts.Organization == "" {
		err := fmt.Errorf("--organization is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	orgType, orgID, err := parseOrganization(opts.Organization)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"name": opts.Name,
	}
	if opts.Description != "" {
		attributes["description"] = opts.Description
	}

	// Build request data with relationships
	data := map[string]any{
		"type":       "culture-values",
		"attributes": attributes,
		"relationships": map[string]any{
			"organization": map[string]any{
				"data": map[string]string{
					"type": orgType,
					"id":   orgID,
				},
			},
		},
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

	body, _, err := client.Post(cmd.Context(), "/v1/culture-values", jsonBody)
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

	row := buildCultureValueRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created culture value %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoCultureValuesCreateOptions(cmd *cobra.Command) (doCultureValuesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	organization, _ := cmd.Flags().GetString("organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCultureValuesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Name:         name,
		Description:  description,
		Organization: organization,
	}, nil
}

func buildCultureValueRowFromSingle(resp jsonAPISingleResponse) cultureValueRow {
	resource := resp.Data

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	row := cultureValueRow{
		ID:               resource.ID,
		Name:             stringAttr(resource.Attributes, "name"),
		Description:      stringAttr(resource.Attributes, "description"),
		SequencePosition: intAttr(resource.Attributes, "sequence-index"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationID = rel.Data.ID
		row.OrganizationType = rel.Data.Type
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Organization = stringAttr(org.Attributes, "company-name")
			if row.Organization == "" {
				row.Organization = stringAttr(org.Attributes, "name")
			}
		}
	}

	return row
}
