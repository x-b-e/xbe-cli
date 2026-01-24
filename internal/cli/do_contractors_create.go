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

type doContractorsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Name    string
	Broker  string
}

func newDoContractorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new contractor",
		Long: `Create a new contractor.

Required flags:
  --name      Contractor name
  --broker    Broker ID`,
		Example: `  # Create a contractor
  xbe do contractors create --name "Acme Subcontracting" --broker 123

  # JSON output
  xbe do contractors create --name "Acme Subcontracting" --broker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoContractorsCreate,
	}
	initDoContractorsCreateFlags(cmd)
	return cmd
}

func init() {
	doContractorsCmd.AddCommand(newDoContractorsCreateCmd())
}

func initDoContractorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Contractor name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("broker")
}

func runDoContractorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoContractorsCreateOptions(cmd)
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

	attributes := map[string]any{
		"name": opts.Name,
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "contractors",
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

	body, _, err := client.Post(cmd.Context(), "/v1/contractors", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created contractor %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoContractorsCreateOptions(cmd *cobra.Command) (doContractorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doContractorsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Name:    name,
		Broker:  broker,
	}, nil
}

func contractorRowFromSingle(resp jsonAPISingleResponse) contractorRow {
	row := contractorRow{
		ID:   resp.Data.ID,
		Name: strings.TrimSpace(stringAttr(resp.Data.Attributes, "name")),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}
