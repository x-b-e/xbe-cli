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

type doContractorsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	ID      string
	Name    string
}

func newDoContractorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a contractor",
		Long: `Update a contractor.

Optional flags:
  --name    Contractor name

Note: Broker relationships cannot be changed after creation.`,
		Example: `  # Update contractor name
  xbe do contractors update 123 --name "Acme Contractors"

  # JSON output
  xbe do contractors update 123 --name "Acme Contractors" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoContractorsUpdate,
	}
	initDoContractorsUpdateFlags(cmd)
	return cmd
}

func init() {
	doContractorsCmd.AddCommand(newDoContractorsUpdateCmd())
}

func initDoContractorsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Contractor name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoContractorsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoContractorsUpdateOptions(cmd, args)
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

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "contractors",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/contractors/"+opts.ID, jsonBody)
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

	row := contractorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated contractor %s\n", row.ID)
	return nil
}

func parseDoContractorsUpdateOptions(cmd *cobra.Command, args []string) (doContractorsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doContractorsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		ID:      args[0],
		Name:    name,
	}, nil
}
