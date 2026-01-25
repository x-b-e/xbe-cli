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

type businessUnitEquipmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type businessUnitEquipmentDetails struct {
	ID                     string `json:"id"`
	BusinessUnitID         string `json:"business_unit_id,omitempty"`
	BusinessUnitName       string `json:"business_unit_name,omitempty"`
	BusinessUnitExternalID string `json:"business_unit_external_id,omitempty"`
	EquipmentID            string `json:"equipment_id,omitempty"`
	EquipmentNickname      string `json:"equipment_nickname,omitempty"`
	EquipmentSerialNumber  string `json:"equipment_serial_number,omitempty"`
}

func newBusinessUnitEquipmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show business unit equipment details",
		Long: `Show the full details of a business unit equipment link.

Output Fields:
  ID                     Resource identifier
  Business Unit           Business unit name or ID
  Business Unit External  Business unit external ID
  Equipment              Equipment nickname or ID
  Equipment Serial        Equipment serial number

Arguments:
  <id>  The business unit equipment ID (required).`,
		Example: `  # Show business unit equipment details
  xbe view business-unit-equipments show 123

  # Output as JSON
  xbe view business-unit-equipments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBusinessUnitEquipmentsShow,
	}
	initBusinessUnitEquipmentsShowFlags(cmd)
	return cmd
}

func init() {
	businessUnitEquipmentsCmd.AddCommand(newBusinessUnitEquipmentsShowCmd())
}

func initBusinessUnitEquipmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBusinessUnitEquipmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBusinessUnitEquipmentsShowOptions(cmd)
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
		return fmt.Errorf("business unit equipment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[business-unit-equipments]", "business-unit,equipment")
	query.Set("include", "business-unit,equipment")
	query.Set("fields[business-units]", "company-name,external-id")
	query.Set("fields[equipment]", "nickname,serial-number")

	body, _, err := client.Get(cmd.Context(), "/v1/business-unit-equipments/"+id, query)
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

	details := buildBusinessUnitEquipmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBusinessUnitEquipmentDetails(cmd, details)
}

func parseBusinessUnitEquipmentsShowOptions(cmd *cobra.Command) (businessUnitEquipmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return businessUnitEquipmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBusinessUnitEquipmentDetails(resp jsonAPISingleResponse) businessUnitEquipmentDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	row := buildBusinessUnitEquipmentRow(resp.Data, included)
	return businessUnitEquipmentDetails{
		ID:                     row.ID,
		BusinessUnitID:         row.BusinessUnitID,
		BusinessUnitName:       row.BusinessUnitName,
		BusinessUnitExternalID: row.BusinessUnitExternalID,
		EquipmentID:            row.EquipmentID,
		EquipmentNickname:      row.EquipmentNickname,
		EquipmentSerialNumber:  row.EquipmentSerialNumber,
	}
}

func renderBusinessUnitEquipmentDetails(cmd *cobra.Command, details businessUnitEquipmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.BusinessUnitID != "" || details.BusinessUnitName != "" {
		fmt.Fprintf(out, "Business Unit: %s\n", formatRelated(details.BusinessUnitName, details.BusinessUnitID))
	}
	if details.BusinessUnitExternalID != "" {
		fmt.Fprintf(out, "Business Unit External ID: %s\n", details.BusinessUnitExternalID)
	}
	if details.EquipmentID != "" || details.EquipmentNickname != "" {
		fmt.Fprintf(out, "Equipment: %s\n", formatRelated(details.EquipmentNickname, details.EquipmentID))
	}
	if details.EquipmentSerialNumber != "" {
		fmt.Fprintf(out, "Equipment Serial: %s\n", details.EquipmentSerialNumber)
	}

	return nil
}
