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

type doPredictionSubjectGapPortionsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Name        string
	Amount      string
	Status      string
	Description string
}

func newDoPredictionSubjectGapPortionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prediction subject gap portion",
		Long: `Update a prediction subject gap portion.

Optional flags:
  --name         Portion name
  --amount       Portion amount
  --status       Portion status (draft/approved)
  --description  Portion description

Notes:
  Status updates may require gap manager permissions.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update name and amount
  xbe do prediction-subject-gap-portions update 123 --name "Labor" --amount 45

  # Update description
  xbe do prediction-subject-gap-portions update 123 --description "Updated description"

  # Update status
  xbe do prediction-subject-gap-portions update 123 --status approved`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPredictionSubjectGapPortionsUpdate,
	}
	initDoPredictionSubjectGapPortionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPredictionSubjectGapPortionsCmd.AddCommand(newDoPredictionSubjectGapPortionsUpdateCmd())
}

func initDoPredictionSubjectGapPortionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Portion name")
	cmd.Flags().String("amount", "", "Portion amount")
	cmd.Flags().String("status", "", "Portion status (draft/approved)")
	cmd.Flags().String("description", "", "Portion description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPredictionSubjectGapPortionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPredictionSubjectGapPortionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("amount") {
		attributes["amount"] = opts.Amount
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "prediction-subject-gap-portions",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/prediction-subject-gap-portions/"+opts.ID, jsonBody)
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

	row := predictionSubjectGapPortionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated prediction subject gap portion %s\n", row.ID)
	return nil
}

func parseDoPredictionSubjectGapPortionsUpdateOptions(cmd *cobra.Command, args []string) (doPredictionSubjectGapPortionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	amount, _ := cmd.Flags().GetString("amount")
	status, _ := cmd.Flags().GetString("status")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPredictionSubjectGapPortionsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Name:        name,
		Amount:      amount,
		Status:      status,
		Description: description,
	}, nil
}
