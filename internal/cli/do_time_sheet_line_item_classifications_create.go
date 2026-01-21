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

type doTimeSheetLineItemClassificationsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Name         string
	Description  string
	SubjectTypes []string
}

func newDoTimeSheetLineItemClassificationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new time sheet line item classification",
		Long: `Create a new time sheet line item classification.

Required flags:
  --name  The classification name (required)

Optional flags:
  --description    Classification description
  --subject-types  Subject types (comma-separated or repeated flag)`,
		Example: `  # Create a basic classification
  xbe do time-sheet-line-item-classifications create --name "Overtime"

  # Create with description
  xbe do time-sheet-line-item-classifications create --name "Overtime" --description "Hours worked beyond standard"

  # Get JSON output
  xbe do time-sheet-line-item-classifications create --name "Holiday" --json`,
		Args: cobra.NoArgs,
		RunE: runDoTimeSheetLineItemClassificationsCreate,
	}
	initDoTimeSheetLineItemClassificationsCreateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetLineItemClassificationsCmd.AddCommand(newDoTimeSheetLineItemClassificationsCreateCmd())
}

func initDoTimeSheetLineItemClassificationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name (required)")
	cmd.Flags().String("description", "", "Classification description")
	cmd.Flags().StringSlice("subject-types", nil, "Subject types (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetLineItemClassificationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetLineItemClassificationsCreateOptions(cmd)
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

	attributes := map[string]any{
		"name": opts.Name,
	}

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if len(opts.SubjectTypes) > 0 {
		attributes["subject-types"] = opts.SubjectTypes
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "time-sheet-line-item-classifications",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/time-sheet-line-item-classifications", jsonBody)
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

	row := buildTimeSheetLineItemClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created time sheet line item classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTimeSheetLineItemClassificationsCreateOptions(cmd *cobra.Command) (doTimeSheetLineItemClassificationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	subjectTypes, _ := cmd.Flags().GetStringSlice("subject-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetLineItemClassificationsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Name:         name,
		Description:  description,
		SubjectTypes: subjectTypes,
	}, nil
}

func buildTimeSheetLineItemClassificationRowFromSingle(resp jsonAPISingleResponse) timeSheetLineItemClassificationRow {
	attrs := resp.Data.Attributes

	row := timeSheetLineItemClassificationRow{
		ID:          resp.Data.ID,
		Name:        stringAttr(attrs, "name"),
		Description: stringAttr(attrs, "description"),
	}

	if st, ok := attrs["subject-types"].([]any); ok {
		for _, s := range st {
			if str, ok := s.(string); ok {
				row.SubjectTypes = append(row.SubjectTypes, str)
			}
		}
	}

	return row
}
