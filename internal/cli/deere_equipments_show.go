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

type deereEquipmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type deereEquipmentDetails struct {
	ID                            string `json:"id"`
	EquipmentName                 string `json:"equipment_name,omitempty"`
	EquipmentSourceID             string `json:"equipment_source_id,omitempty"`
	EquipmentSerialNumber         string `json:"equipment_serial_number,omitempty"`
	IntegrationIdentifier         string `json:"integration_identifier,omitempty"`
	EquipmentSetAt                string `json:"equipment_set_at,omitempty"`
	BrokerID                      string `json:"broker_id,omitempty"`
	BrokerName                    string `json:"broker_name,omitempty"`
	EquipmentID                   string `json:"equipment_id,omitempty"`
	EquipmentNickname             string `json:"equipment_nickname,omitempty"`
	AssignedEquipmentSerialNumber string `json:"assigned_equipment_serial_number,omitempty"`
}

func newDeereEquipmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show Deere equipment details",
		Long: `Show the full details of a Deere equipment record.

Output Fields:
  ID                 Deere equipment identifier
  Name               Equipment name
  Source ID          Deere equipment source identifier
  Serial Number      Deere equipment serial number
  Integration ID     Integration identifier
  Equipment Set At   Equipment assignment timestamp
  Broker             Broker name or ID
  Equipment          Assigned equipment nickname or ID
  Assigned Serial    Assigned equipment serial number

Arguments:
  <id>  Deere equipment ID (required). Find IDs using the list command.`,
		Example: `  # Show Deere equipment details
  xbe view deere-equipments show 123

  # Output as JSON
  xbe view deere-equipments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDeereEquipmentsShow,
	}
	initDeereEquipmentsShowFlags(cmd)
	return cmd
}

func init() {
	deereEquipmentsCmd.AddCommand(newDeereEquipmentsShowCmd())
}

func initDeereEquipmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeereEquipmentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseDeereEquipmentsShowOptions(cmd)
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
		return fmt.Errorf("deere equipment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[deere-equipments]", "equipment-name,equipment-source-id,equipment-serial-number,integration-identifier,equipment-set-at,broker,equipment")
	query.Set("include", "broker,equipment")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[equipment]", "nickname,serial-number")

	body, _, err := client.Get(cmd.Context(), "/v1/deere-equipments/"+id, query)
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

	details := buildDeereEquipmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDeereEquipmentDetails(cmd, details)
}

func parseDeereEquipmentsShowOptions(cmd *cobra.Command) (deereEquipmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return deereEquipmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDeereEquipmentDetails(resp jsonAPISingleResponse) deereEquipmentDetails {
	row := deereEquipmentRowFromSingle(resp)

	details := deereEquipmentDetails{
		ID:                            row.ID,
		EquipmentName:                 row.EquipmentName,
		EquipmentSourceID:             row.EquipmentSourceID,
		EquipmentSerialNumber:         row.EquipmentSerialNumber,
		IntegrationIdentifier:         row.IntegrationIdentifier,
		EquipmentSetAt:                row.EquipmentSetAt,
		BrokerID:                      row.BrokerID,
		BrokerName:                    row.BrokerName,
		EquipmentID:                   row.EquipmentID,
		EquipmentNickname:             row.EquipmentNickname,
		AssignedEquipmentSerialNumber: row.AssignedEquipmentSerialNumber,
	}

	return details
}

func renderDeereEquipmentDetails(cmd *cobra.Command, details deereEquipmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.EquipmentName != "" {
		fmt.Fprintf(out, "Name: %s\n", details.EquipmentName)
	}
	if details.EquipmentSourceID != "" {
		fmt.Fprintf(out, "Source ID: %s\n", details.EquipmentSourceID)
	}
	if details.EquipmentSerialNumber != "" {
		fmt.Fprintf(out, "Serial Number: %s\n", details.EquipmentSerialNumber)
	}
	if details.IntegrationIdentifier != "" {
		fmt.Fprintf(out, "Integration ID: %s\n", details.IntegrationIdentifier)
	}
	if details.EquipmentSetAt != "" {
		fmt.Fprintf(out, "Equipment Set At: %s\n", details.EquipmentSetAt)
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.EquipmentID != "" || details.EquipmentNickname != "" {
		fmt.Fprintf(out, "Equipment: %s\n", formatRelated(details.EquipmentNickname, details.EquipmentID))
	}
	if details.AssignedEquipmentSerialNumber != "" {
		fmt.Fprintf(out, "Assigned Serial: %s\n", details.AssignedEquipmentSerialNumber)
	}
	return nil
}
