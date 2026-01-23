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

type doDeveloperReferencesCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	DeveloperReferenceType string
	SubjectType            string
	SubjectID              string
	Value                  string
}

func newDoDeveloperReferencesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a developer reference",
		Long: `Create a developer reference.

Required:
  --developer-reference-type  Developer reference type ID
  --subject-type              Subject type (e.g., projects, customers)
  --subject-id                Subject ID

Optional:
  --value                     Reference value`,
		Example: `  # Create a developer reference
  xbe do developer-references create --developer-reference-type 123 --subject-type projects --subject-id 456

  # Create with value
  xbe do developer-references create --developer-reference-type 123 --subject-type projects --subject-id 456 \
    --value "EXT-12345"`,
		RunE: runDoDeveloperReferencesCreate,
	}
	initDoDeveloperReferencesCreateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperReferencesCmd.AddCommand(newDoDeveloperReferencesCreateCmd())
}

func initDoDeveloperReferencesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("developer-reference-type", "", "Developer reference type ID")
	cmd.Flags().String("subject-type", "", "Subject type (e.g., projects, customers)")
	cmd.Flags().String("subject-id", "", "Subject ID")
	cmd.Flags().String("value", "", "Reference value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("developer-reference-type")
	_ = cmd.MarkFlagRequired("subject-type")
	_ = cmd.MarkFlagRequired("subject-id")
}

func runDoDeveloperReferencesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDeveloperReferencesCreateOptions(cmd)
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

	if opts.Value != "" {
		attributes["value"] = opts.Value
	}

	relationships := map[string]any{
		"developer-reference-type": map[string]any{
			"data": map[string]any{
				"type": "developer-reference-types",
				"id":   opts.DeveloperReferenceType,
			},
		},
		"subject": map[string]any{
			"data": map[string]any{
				"type": opts.SubjectType,
				"id":   opts.SubjectID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "developer-references",
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

	body, _, err := client.Post(cmd.Context(), "/v1/developer-references", jsonBody)
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
		row := developerReferenceRow{
			ID:    resp.Data.ID,
			Value: stringAttr(resp.Data.Attributes, "value"),
		}
		if rel, ok := resp.Data.Relationships["developer-reference-type"]; ok && rel.Data != nil {
			row.DeveloperReferenceTypeID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["subject"]; ok && rel.Data != nil {
			row.SubjectType = rel.Data.Type
			row.SubjectID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created developer reference %s\n", resp.Data.ID)
	return nil
}

func parseDoDeveloperReferencesCreateOptions(cmd *cobra.Command) (doDeveloperReferencesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	developerReferenceType, _ := cmd.Flags().GetString("developer-reference-type")
	subjectType, _ := cmd.Flags().GetString("subject-type")
	subjectID, _ := cmd.Flags().GetString("subject-id")
	value, _ := cmd.Flags().GetString("value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperReferencesCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		DeveloperReferenceType: developerReferenceType,
		SubjectType:            subjectType,
		SubjectID:              subjectID,
		Value:                  value,
	}, nil
}
