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

type doEquipmentUtilizationReadingsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	Equipment     string
	BusinessUnit  string
	User          string
	ReportedAt    string
	Odometer      string
	Hourmeter     string
	OtherReadings string
}

func newDoEquipmentUtilizationReadingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment utilization reading",
		Long: `Create an equipment utilization reading.

Required flags:
  --equipment        Equipment ID
  --reported-at      Reported timestamp (ISO 8601)
  --odometer OR --hourmeter

Optional flags:
  --business-unit    Business unit ID
  --user             User ID
  --other-readings   Other readings payload (JSON string)`,
		Example: `  # Create a reading with an odometer value
  xbe do equipment-utilization-readings create --equipment 123 --reported-at 2025-01-01T08:00:00Z --odometer 100

  # Create a reading with hourmeter and source
  xbe do equipment-utilization-readings create --equipment 123 --reported-at 2025-01-02T08:00:00Z \
    --hourmeter 12 --other-readings '{"source":"telematics"}'

  # Get JSON output
  xbe do equipment-utilization-readings create --equipment 123 --reported-at 2025-01-01T08:00:00Z --odometer 100 --json`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentUtilizationReadingsCreate,
	}
	initDoEquipmentUtilizationReadingsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentUtilizationReadingsCmd.AddCommand(newDoEquipmentUtilizationReadingsCreateCmd())
}

func initDoEquipmentUtilizationReadingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("business-unit", "", "Business unit ID")
	cmd.Flags().String("user", "", "User ID")
	cmd.Flags().String("reported-at", "", "Reported timestamp (ISO 8601)")
	cmd.Flags().String("odometer", "", "Odometer reading")
	cmd.Flags().String("hourmeter", "", "Hourmeter reading")
	cmd.Flags().String("other-readings", "", "Other readings payload (JSON string)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentUtilizationReadingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentUtilizationReadingsCreateOptions(cmd)
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

	if opts.Equipment == "" {
		err := fmt.Errorf("--equipment is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.ReportedAt == "" {
		err := fmt.Errorf("--reported-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Odometer == "" && opts.Hourmeter == "" {
		err := fmt.Errorf("at least one of --odometer or --hourmeter is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"reported-at": opts.ReportedAt,
	}
	if opts.Odometer != "" {
		attributes["odometer"] = opts.Odometer
	}
	if opts.Hourmeter != "" {
		attributes["hourmeter"] = opts.Hourmeter
	}
	if opts.OtherReadings != "" {
		var other any
		if err := json.Unmarshal([]byte(opts.OtherReadings), &other); err != nil {
			err := fmt.Errorf("invalid other-readings JSON: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["other-readings"] = other
	}

	relationships := map[string]any{
		"equipment": map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		},
	}

	if opts.BusinessUnit != "" {
		relationships["business-unit"] = map[string]any{
			"data": map[string]any{
				"type": "business-units",
				"id":   opts.BusinessUnit,
			},
		}
	}
	if opts.User != "" {
		relationships["user"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-utilization-readings",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-utilization-readings", jsonBody)
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

	row := buildEquipmentUtilizationReadingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment utilization reading %s\n", row.ID)
	return nil
}

func parseDoEquipmentUtilizationReadingsCreateOptions(cmd *cobra.Command) (doEquipmentUtilizationReadingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipment, _ := cmd.Flags().GetString("equipment")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	user, _ := cmd.Flags().GetString("user")
	reportedAt, _ := cmd.Flags().GetString("reported-at")
	odometer, _ := cmd.Flags().GetString("odometer")
	hourmeter, _ := cmd.Flags().GetString("hourmeter")
	otherReadings, _ := cmd.Flags().GetString("other-readings")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentUtilizationReadingsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		Equipment:     equipment,
		BusinessUnit:  businessUnit,
		User:          user,
		ReportedAt:    reportedAt,
		Odometer:      odometer,
		Hourmeter:     hourmeter,
		OtherReadings: otherReadings,
	}, nil
}
