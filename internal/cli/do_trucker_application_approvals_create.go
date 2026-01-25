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

type doTruckerApplicationApprovalsCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	TruckerApplicationID               string
	AddApplicationUserAsTruckerManager bool
}

type truckerApplicationApprovalRow struct {
	ID                                 string `json:"id"`
	TruckerApplicationID               string `json:"trucker_application_id,omitempty"`
	TruckerID                          string `json:"trucker_id,omitempty"`
	AddApplicationUserAsTruckerManager bool   `json:"add_application_user_as_trucker_manager,omitempty"`
}

func newDoTruckerApplicationApprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Approve a trucker application",
		Long: `Approve a trucker application.

Approvals create a trucker from the application data and mark the application
as approved.

Required flags:
  --trucker-application   Trucker application ID

Optional flags:
  --add-application-user-as-trucker-manager   Also add the application user as a trucker manager`,
		Example: `  # Approve a trucker application
  xbe do trucker-application-approvals create --trucker-application 123

  # Also add the application user as a trucker manager
  xbe do trucker-application-approvals create --trucker-application 123 --add-application-user-as-trucker-manager

  # JSON output
  xbe do trucker-application-approvals create --trucker-application 123 --json`,
		RunE: runDoTruckerApplicationApprovalsCreate,
	}
	initDoTruckerApplicationApprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerApplicationApprovalsCmd.AddCommand(newDoTruckerApplicationApprovalsCreateCmd())
}

func initDoTruckerApplicationApprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker-application", "", "Trucker application ID (required)")
	cmd.Flags().Bool("add-application-user-as-trucker-manager", false, "Add the application user as a trucker manager")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("trucker-application")
}

func runDoTruckerApplicationApprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerApplicationApprovalsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{
		"trucker-application-id": opts.TruckerApplicationID,
	}
	if cmd.Flags().Changed("add-application-user-as-trucker-manager") {
		attributes["add-application-user-as-trucker-manager"] = opts.AddApplicationUserAsTruckerManager
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "trucker-application-approvals",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-application-approvals", jsonBody)
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

	row := buildTruckerApplicationApprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker application approval %s\n", row.ID)
	return nil
}

func parseDoTruckerApplicationApprovalsCreateOptions(cmd *cobra.Command) (doTruckerApplicationApprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	truckerApplicationID, _ := cmd.Flags().GetString("trucker-application")
	addApplicationUserAsTruckerManager, _ := cmd.Flags().GetBool("add-application-user-as-trucker-manager")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerApplicationApprovalsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		TruckerApplicationID:               truckerApplicationID,
		AddApplicationUserAsTruckerManager: addApplicationUserAsTruckerManager,
	}, nil
}

func buildTruckerApplicationApprovalRowFromSingle(resp jsonAPISingleResponse) truckerApplicationApprovalRow {
	resource := resp.Data
	row := truckerApplicationApprovalRow{
		ID:                                 resource.ID,
		TruckerApplicationID:               stringAttr(resource.Attributes, "trucker-application-id"),
		TruckerID:                          stringAttr(resource.Attributes, "trucker-id"),
		AddApplicationUserAsTruckerManager: boolAttr(resource.Attributes, "add-application-user-as-trucker-manager"),
	}
	if row.TruckerApplicationID == "" {
		row.TruckerApplicationID = resource.ID
	}
	return row
}
