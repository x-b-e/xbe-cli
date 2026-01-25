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

type doRawTransportTractorsCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	ExternalTractorID string
	Importer          string
	Tables            string
	Broker            string
}

func newDoRawTransportTractorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a raw transport tractor",
		Long: `Create a raw transport tractor.

Required flags:
  --external-tractor-id  External tractor identifier (required)
  --broker               Broker ID (required)

Optional flags:
  --importer  Importer name (e.g., quantix_tmw)
  --tables    Raw table payloads as JSON array`,
		Example: `  # Create a raw transport tractor
  xbe do raw-transport-tractors create --external-tractor-id TRC-123 --broker 456 --importer quantix_tmw

  # Create with table payloads
  xbe do raw-transport-tractors create \\
    --external-tractor-id TRC-456 \\
    --broker 456 \\
    --importer quantix_tmw \\
    --tables '[{\"table_name\":\"tractorprofile\",\"rows\":[]}]'

  # Output as JSON
  xbe do raw-transport-tractors create --external-tractor-id TRC-789 --broker 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoRawTransportTractorsCreate,
	}
	initDoRawTransportTractorsCreateFlags(cmd)
	return cmd
}

func init() {
	doRawTransportTractorsCmd.AddCommand(newDoRawTransportTractorsCreateCmd())
}

func initDoRawTransportTractorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-tractor-id", "", "External tractor identifier (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("importer", "", "Importer name (e.g., quantix_tmw)")
	cmd.Flags().String("tables", "", "Raw table payloads as JSON array")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("external-tractor-id")
	_ = cmd.MarkFlagRequired("broker")
}

func runDoRawTransportTractorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRawTransportTractorsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ExternalTractorID) == "" {
		err := fmt.Errorf("--external-tractor-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Broker) == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"external-tractor-id": opts.ExternalTractorID,
	}
	if strings.TrimSpace(opts.Importer) != "" {
		attributes["importer"] = opts.Importer
	}
	if strings.TrimSpace(opts.Tables) != "" {
		var tables any
		if err := json.Unmarshal([]byte(opts.Tables), &tables); err != nil {
			err = fmt.Errorf("invalid tables JSON: %w", err)
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if _, ok := tables.([]any); !ok {
			err := fmt.Errorf("--tables must be a JSON array")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["tables"] = tables
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
			"type":          "raw-transport-tractors",
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

	body, _, err := client.Post(cmd.Context(), "/v1/raw-transport-tractors", jsonBody)
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

	row := rawTransportTractorRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created raw transport tractor %s\n", row.ID)
	return nil
}

func parseDoRawTransportTractorsCreateOptions(cmd *cobra.Command) (doRawTransportTractorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalTractorID, _ := cmd.Flags().GetString("external-tractor-id")
	broker, _ := cmd.Flags().GetString("broker")
	importer, _ := cmd.Flags().GetString("importer")
	tables, _ := cmd.Flags().GetString("tables")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRawTransportTractorsCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		ExternalTractorID: externalTractorID,
		Importer:          importer,
		Tables:            tables,
		Broker:            broker,
	}, nil
}
