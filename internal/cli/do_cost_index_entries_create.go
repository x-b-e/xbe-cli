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

type doCostIndexEntriesCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	CostIndex string
	StartOn   string
	EndOn     string
	Value     string
}

func newDoCostIndexEntriesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cost index entry",
		Long: `Create a new cost index entry.

Required flags:
  --cost-index  The parent cost index ID (required)
  --start-on    Entry start date (required, format: YYYY-MM-DD)
  --end-on      Entry end date (required, format: YYYY-MM-DD)
  --value       Entry value (required)`,
		Example: `  # Create a cost index entry
  xbe do cost-index-entries create --cost-index 123 --start-on "2024-01-01" --end-on "2024-03-31" --value 1.05`,
		Args: cobra.NoArgs,
		RunE: runDoCostIndexEntriesCreate,
	}
	initDoCostIndexEntriesCreateFlags(cmd)
	return cmd
}

func init() {
	doCostIndexEntriesCmd.AddCommand(newDoCostIndexEntriesCreateCmd())
}

func initDoCostIndexEntriesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("cost-index", "", "Parent cost index ID (required)")
	cmd.Flags().String("start-on", "", "Entry start date (required)")
	cmd.Flags().String("end-on", "", "Entry end date (required)")
	cmd.Flags().String("value", "", "Entry value (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCostIndexEntriesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCostIndexEntriesCreateOptions(cmd)
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

	if opts.CostIndex == "" {
		err := fmt.Errorf("--cost-index is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.StartOn == "" {
		err := fmt.Errorf("--start-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.EndOn == "" {
		err := fmt.Errorf("--end-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Value == "" {
		err := fmt.Errorf("--value is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"start-on": opts.StartOn,
		"end-on":   opts.EndOn,
		"value":    opts.Value,
	}

	relationships := map[string]any{
		"cost-index": map[string]any{
			"data": map[string]any{
				"type": "cost-indexes",
				"id":   opts.CostIndex,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "cost-index-entries",
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

	body, _, err := client.Post(cmd.Context(), "/v1/cost-index-entries", jsonBody)
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

	row := buildCostIndexEntryRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created cost index entry %s\n", row.ID)
	return nil
}

func parseDoCostIndexEntriesCreateOptions(cmd *cobra.Command) (doCostIndexEntriesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	costIndex, _ := cmd.Flags().GetString("cost-index")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	value, _ := cmd.Flags().GetString("value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCostIndexEntriesCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		CostIndex: costIndex,
		StartOn:   startOn,
		EndOn:     endOn,
		Value:     value,
	}, nil
}

func buildCostIndexEntryRowFromSingle(resp jsonAPISingleResponse) costIndexEntryRow {
	attrs := resp.Data.Attributes

	row := costIndexEntryRow{
		ID:      resp.Data.ID,
		StartOn: stringAttr(attrs, "start-on"),
		EndOn:   stringAttr(attrs, "end-on"),
		Value:   attrs["value"],
	}

	if rel, ok := resp.Data.Relationships["cost-index"]; ok && rel.Data != nil {
		row.CostIndexID = rel.Data.ID
	}

	return row
}
