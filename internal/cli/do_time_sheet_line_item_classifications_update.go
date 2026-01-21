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

type doTimeSheetLineItemClassificationsUpdateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	ID           string
	Name         string
	Description  string
	SubjectTypes []string
}

func newDoTimeSheetLineItemClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing time sheet line item classification",
		Long: `Update an existing time sheet line item classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name           The classification name
  --description    Classification description
  --subject-types  Subject types (comma-separated or repeated)`,
		Example: `  # Update name
  xbe do time-sheet-line-item-classifications update 123 --name "Updated Name"

  # Update multiple fields
  xbe do time-sheet-line-item-classifications update 123 --name "New Name" --description "New desc"

  # Get JSON output
  xbe do time-sheet-line-item-classifications update 123 --name "New Name" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTimeSheetLineItemClassificationsUpdate,
	}
	initDoTimeSheetLineItemClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTimeSheetLineItemClassificationsCmd.AddCommand(newDoTimeSheetLineItemClassificationsUpdateCmd())
}

func initDoTimeSheetLineItemClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Classification name")
	cmd.Flags().String("description", "", "Classification description")
	cmd.Flags().StringSlice("subject-types", nil, "Subject types (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTimeSheetLineItemClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTimeSheetLineItemClassificationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("subject-types") {
		attributes["subject-types"] = opts.SubjectTypes
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one of --name, --description, --subject-types")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "time-sheet-line-item-classifications",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/time-sheet-line-item-classifications/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated time sheet line item classification %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoTimeSheetLineItemClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doTimeSheetLineItemClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	subjectTypes, _ := cmd.Flags().GetStringSlice("subject-types")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTimeSheetLineItemClassificationsUpdateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		ID:           args[0],
		Name:         name,
		Description:  description,
		SubjectTypes: subjectTypes,
	}, nil
}
