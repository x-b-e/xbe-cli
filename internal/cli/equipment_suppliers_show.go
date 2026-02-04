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

type equipmentSuppliersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentSupplierDetails struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	ContractNumber     string   `json:"contract_number,omitempty"`
	BrokerID           string   `json:"broker_id,omitempty"`
	BrokerName         string   `json:"broker_name,omitempty"`
	EquipmentRentalIDs []string `json:"equipment_rental_ids,omitempty"`
	FileAttachmentIDs  []string `json:"file_attachment_ids,omitempty"`
}

func newEquipmentSuppliersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment supplier details",
		Long: `Show the full details of a specific equipment supplier.

Output Fields:
  ID                Equipment supplier identifier
  Name              Supplier name
  Contract Number   Contract number
  Broker            Broker name and ID
  Equipment Rentals Equipment rental IDs
  File Attachments  File attachment IDs

Arguments:
  <id>    The equipment supplier ID (required). You can find IDs using the list command.`,
		Example: `  # Show equipment supplier details
  xbe view equipment-suppliers show 123

  # Get JSON output
  xbe view equipment-suppliers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentSuppliersShow,
	}
	initEquipmentSuppliersShowFlags(cmd)
	return cmd
}

func init() {
	equipmentSuppliersCmd.AddCommand(newEquipmentSuppliersShowCmd())
}

func initEquipmentSuppliersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentSuppliersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseEquipmentSuppliersShowOptions(cmd)
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
		return fmt.Errorf("equipment supplier id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[equipment-suppliers]", "name,contract-number,broker,equipment-rentals,file-attachments")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-suppliers/"+id, query)
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

	details := buildEquipmentSupplierDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentSupplierDetails(cmd, details)
}

func parseEquipmentSuppliersShowOptions(cmd *cobra.Command) (equipmentSuppliersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentSuppliersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentSupplierDetails(resp jsonAPISingleResponse) equipmentSupplierDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := equipmentSupplierDetails{
		ID:             resource.ID,
		Name:           strings.TrimSpace(stringAttr(attrs, "name")),
		ContractNumber: strings.TrimSpace(stringAttr(attrs, "contract-number")),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}
	if rel, ok := resource.Relationships["equipment-rentals"]; ok && rel.raw != nil {
		details.EquipmentRentalIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok && rel.raw != nil {
		details.FileAttachmentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderEquipmentSupplierDetails(cmd *cobra.Command, details equipmentSupplierDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.ContractNumber != "" {
		fmt.Fprintf(out, "Contract Number: %s\n", details.ContractNumber)
	}
	if details.BrokerName != "" && details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s (%s)\n", details.BrokerName, details.BrokerID)
	} else if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if len(details.EquipmentRentalIDs) > 0 {
		fmt.Fprintf(out, "Equipment Rentals: %s\n", strings.Join(details.EquipmentRentalIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachments: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
