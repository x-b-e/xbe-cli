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

type doObjectiveStakeholderClassificationsUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	InterestDegree float64
}

func newDoObjectiveStakeholderClassificationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an objective stakeholder classification",
		Long: `Update an objective stakeholder classification.

Provide the classification ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --interest-degree  Interest degree between 0 and 1

Note: Objective and stakeholder classification cannot be changed after creation.
Only admin users can update objective stakeholder classifications.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update interest degree
  xbe do objective-stakeholder-classifications update 123 --interest-degree 0.85

  # Get JSON output
  xbe do objective-stakeholder-classifications update 123 --interest-degree 0.65 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoObjectiveStakeholderClassificationsUpdate,
	}
	initDoObjectiveStakeholderClassificationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doObjectiveStakeholderClassificationsCmd.AddCommand(newDoObjectiveStakeholderClassificationsUpdateCmd())
}

func initDoObjectiveStakeholderClassificationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Float64("interest-degree", 0, "Interest degree between 0 and 1")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoObjectiveStakeholderClassificationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoObjectiveStakeholderClassificationsUpdateOptions(cmd, args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("interest-degree") {
		if opts.InterestDegree <= 0 || opts.InterestDegree > 1 {
			err := fmt.Errorf("--interest-degree must be between 0 and 1")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["interest-degree"] = opts.InterestDegree
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --interest-degree")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "objective-stakeholder-classifications",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Patch(cmd.Context(), "/v1/objective-stakeholder-classifications/"+opts.ID, jsonBody)
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

	row := buildObjectiveStakeholderClassificationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated objective stakeholder classification %s\n", row.ID)
	return nil
}

func parseDoObjectiveStakeholderClassificationsUpdateOptions(cmd *cobra.Command, args []string) (doObjectiveStakeholderClassificationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	interestDegree, _ := cmd.Flags().GetFloat64("interest-degree")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doObjectiveStakeholderClassificationsUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		InterestDegree: interestDegree,
	}, nil
}
