package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type brokerEquipmentClassificationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerEquipmentClassificationDetails struct {
	ID                        string `json:"id"`
	BrokerID                  string `json:"broker_id,omitempty"`
	EquipmentClassificationID string `json:"equipment_classification_id,omitempty"`
}

func newBrokerEquipmentClassificationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker equipment classification details",
		Long: `Show the full details of a broker equipment classification.

Output Fields:
  ID                 Broker equipment classification identifier
  BROKER ID          Broker ID
  EQUIP CLASS ID     Equipment classification ID

Arguments:
  <id>  The broker equipment classification ID (required). Use the list command to find IDs.`,
		Example: `  # Show a broker equipment classification
  xbe view broker-equipment-classifications show 123

  # Show as JSON
  xbe view broker-equipment-classifications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerEquipmentClassificationsShow,
	}
	initBrokerEquipmentClassificationsShowFlags(cmd)
	return cmd
}

func init() {
	brokerEquipmentClassificationsCmd.AddCommand(newBrokerEquipmentClassificationsShowCmd())
}

func initBrokerEquipmentClassificationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerEquipmentClassificationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerEquipmentClassificationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("broker equipment classification id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-equipment-classifications]", "broker,equipment-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-equipment-classifications/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildBrokerEquipmentClassificationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerEquipmentClassificationDetails(cmd, details)
}

func parseBrokerEquipmentClassificationsShowOptions(cmd *cobra.Command) (brokerEquipmentClassificationsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return brokerEquipmentClassificationsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return brokerEquipmentClassificationsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return brokerEquipmentClassificationsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return brokerEquipmentClassificationsShowOptions{}, err
	}

	return brokerEquipmentClassificationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerEquipmentClassificationDetails(resp jsonAPISingleResponse) brokerEquipmentClassificationDetails {
	details := brokerEquipmentClassificationDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-classification"]; ok && rel.Data != nil {
		details.EquipmentClassificationID = rel.Data.ID
	}

	return details
}

func renderBrokerEquipmentClassificationDetails(cmd *cobra.Command, details brokerEquipmentClassificationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.EquipmentClassificationID != "" {
		fmt.Fprintf(out, "Equipment Classification ID: %s\n", details.EquipmentClassificationID)
	}

	return nil
}
