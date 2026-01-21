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

type doDeveloperReferenceTypesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Name         string
	SubjectTypes []string
	Developer    string
}

func newDoDeveloperReferenceTypesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new developer reference type",
		Long: `Create a new developer reference type.

Required flags:
  --name       The reference type name (required)
  --developer  The developer ID (required)

Optional flags:
  --subject-types  Subject types this reference applies to (comma-separated or repeated)`,
		Example: `  # Create a developer reference type
  xbe do developer-reference-types create --name "PO Number" --developer 123

  # Create with subject types
  xbe do developer-reference-types create --name "PO Number" --developer 123 --subject-types "Job,Project"`,
		Args: cobra.NoArgs,
		RunE: runDoDeveloperReferenceTypesCreate,
	}
	initDoDeveloperReferenceTypesCreateFlags(cmd)
	return cmd
}

func init() {
	doDeveloperReferenceTypesCmd.AddCommand(newDoDeveloperReferenceTypesCreateCmd())
}

func initDoDeveloperReferenceTypesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Reference type name (required)")
	cmd.Flags().StringSlice("subject-types", nil, "Subject types (comma-separated or repeated)")
	cmd.Flags().String("developer", "", "Developer ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeveloperReferenceTypesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeveloperReferenceTypesCreateOptions(cmd)
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

	if opts.Developer == "" {
		err := fmt.Errorf("--developer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name": opts.Name,
	}

	if len(opts.SubjectTypes) > 0 {
		attributes["subject-types"] = opts.SubjectTypes
	}

	relationships := map[string]any{
		"developer": map[string]any{
			"data": map[string]any{
				"type": "developers",
				"id":   opts.Developer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "developer-reference-types",
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

	body, _, err := client.Post(cmd.Context(), "/v1/developer-reference-types", jsonBody)
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

	row := buildDeveloperReferenceTypeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created developer reference type %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoDeveloperReferenceTypesCreateOptions(cmd *cobra.Command) (doDeveloperReferenceTypesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	subjectTypes, _ := cmd.Flags().GetStringSlice("subject-types")
	developer, _ := cmd.Flags().GetString("developer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDeveloperReferenceTypesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Name:         name,
		SubjectTypes: subjectTypes,
		Developer:    developer,
	}, nil
}

func buildDeveloperReferenceTypeRowFromSingle(resp jsonAPISingleResponse) developerReferenceTypeRow {
	attrs := resp.Data.Attributes

	row := developerReferenceTypeRow{
		ID:   resp.Data.ID,
		Name: stringAttr(attrs, "name"),
	}

	if st, ok := attrs["subject-types"].([]any); ok {
		for _, s := range st {
			if str, ok := s.(string); ok {
				row.SubjectTypes = append(row.SubjectTypes, str)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["developer"]; ok && rel.Data != nil {
		row.DeveloperID = rel.Data.ID
	}

	return row
}
