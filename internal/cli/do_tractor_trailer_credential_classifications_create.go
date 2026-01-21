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

type doTractorTrailerCredentialClassificationsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	Name             string
	Description      string
	IssuerName       string
	ExternalID       string
	OrganizationType string
	OrganizationID   string
}

func newDoTractorTrailerCredentialClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tractor/trailer credential classification",
		Long: `Create a new tractor/trailer credential classification.

Required flags:
  --name               The classification name (required)
  --organization-type  Organization type (e.g., brokers, truckers) (required)
  --organization-id    Organization ID (required)

Optional flags:
  --description  Classification description
  --issuer-name  Name of the issuing authority
  --external-id  External identifier`,
		Example: `  # Create a tractor/trailer credential classification
  xbe do tractor-trailer-credential-classifications create --name "Insurance" --organization-type brokers --organization-id 123

  # Create with all options
  xbe do tractor-trailer-credential-classifications create --name "Insurance" --description "Liability insurance" --issuer-name "State DMV" --organization-type brokers --organization-id 123`,
		Args: cobra.NoArgs,
		RunE: runDoTractorTrailerCredentialClassificationsCreate,
	}
	initDoTractorTrailerCredentialClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTractorTrailerCredentialClassificationsCmd.AddCommand(newDoTractorTrailerCredentialClassificationsCreateCmd())
}

func initDoTractorTrailerCredentialClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name (required)")
	cmd.Flags().String("description", "", "Classification description")
	cmd.Flags().String("issuer-name", "", "Issuing authority name")
	cmd.Flags().String("external-id", "", "External identifier")
	cmd.Flags().String("organization-type", "", "Organization type (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorTrailerCredentialClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorTrailerCredentialClassificationsCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.IssuerName != "" {
		attributes["issuer-name"] = opts.IssuerName
	}
	if opts.ExternalID != "" {
		attributes["external-id"] = opts.ExternalID
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tractor-trailer-credential-classifications",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tractor-trailer-credential-classifications", jsonBody)
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

	row := buildTractorTrailerCredentialClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tractor/trailer credential classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTractorTrailerCredentialClassificationsCreateOptions(cmd *cobra.Command) (doTractorTrailerCredentialClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	issuerName, _ := cmd.Flags().GetString("issuer-name")
	externalID, _ := cmd.Flags().GetString("external-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorTrailerCredentialClassificationsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		Name:             name,
		Description:      description,
		IssuerName:       issuerName,
		ExternalID:       externalID,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
	}, nil
}

func buildTractorTrailerCredentialClassificationRowFromSingle(resp jsonAPISingleResponse) tractorTrailerCredentialClassificationRow {
	attrs := resp.Data.Attributes

	row := tractorTrailerCredentialClassificationRow{
		ID:          resp.Data.ID,
		Name:        stringAttr(attrs, "name"),
		Description: stringAttr(attrs, "description"),
		IssuerName:  stringAttr(attrs, "issuer-name"),
		ExternalID:  stringAttr(attrs, "external-id"),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}
