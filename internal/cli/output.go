package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type outputFormat string

const (
	outputTable outputFormat = "table"
	outputJSON  outputFormat = "json"
	outputYAML  outputFormat = "yaml"
)

type outputSettings struct {
	Format      outputFormat
	JQ          string
	Buffer      *bytes.Buffer
	OriginalOut io.Writer
	OutputSet   bool
}

type outputContextKey string

const outputSettingsKey outputContextKey = "output_settings"

var lastOutputCmd *cobra.Command

func initOutputFlags(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	flags := cmd.PersistentFlags()
	if flags.Lookup("output") == nil {
		flags.String("output", string(outputTable), "Output format: table, json, yaml")
	}
	if flags.Lookup("jq") == nil {
		flags.String("jq", "", "Apply jq-style filter to JSON output")
	}
}

func prepareOutput(cmd *cobra.Command) error {
	settings, err := resolveOutputSettings(cmd)
	if err != nil {
		return err
	}
	if settings.OutputSet && settings.Format == outputTable {
		if err := setBoolFlag(cmd, "json", false); err != nil {
			return err
		}
	} else if settings.Format != outputTable {
		if err := setBoolFlag(cmd, "json", true); err != nil {
			return err
		}
	}
	if settings.Buffer != nil {
		lastOutputCmd = cmd
		cmd.SetOut(settings.Buffer)
		cmd.SetContext(context.WithValue(cmd.Context(), outputSettingsKey, settings))
	}
	return nil
}

func finalizeOutput(cmdErr error) error {
	if lastOutputCmd == nil {
		return cmdErr
	}
	settings, ok := lastOutputCmd.Context().Value(outputSettingsKey).(outputSettings)
	if !ok || settings.Buffer == nil {
		lastOutputCmd = nil
		return cmdErr
	}
	errOut := lastOutputCmd.ErrOrStderr()
	lastOutputCmd = nil

	if cmdErr != nil {
		return cmdErr
	}

	payload := settings.Buffer.Bytes()
	value, err := decodeJSON(payload)
	if err != nil {
		fmt.Fprintln(settings.OriginalOut, string(payload))
		return err
	}

	if settings.JQ != "" {
		filtered, err := applyJQ(value, settings.JQ)
		if err != nil {
			fmt.Fprintln(errOut, err.Error())
			return err
		}
		value = filtered
	}

	switch settings.Format {
	case outputYAML:
		return writeYAMLOutput(settings.OriginalOut, value)
	case outputJSON:
		return writeJSONOutput(settings.OriginalOut, value)
	default:
		return writeJSONOutput(settings.OriginalOut, value)
	}
}

func resolveOutputSettings(cmd *cobra.Command) (outputSettings, error) {
	outputRaw := strings.ToLower(strings.TrimSpace(getStringFlag(cmd, "output")))
	if outputRaw == "" {
		outputRaw = string(outputTable)
	}
	outputChanged := flagChanged(cmd, "output")
	jqExpr := strings.TrimSpace(getStringFlag(cmd, "jq"))
	jsonFlag := getBoolFlag(cmd, "json")

	if !outputChanged && jsonFlag {
		outputRaw = string(outputJSON)
	}
	if jqExpr != "" && !outputChanged {
		outputRaw = string(outputJSON)
	}

	format := outputFormat(outputRaw)
	if format != outputTable && format != outputJSON && format != outputYAML {
		return outputSettings{}, fmt.Errorf("invalid --output value %q (use table, json, yaml)", outputRaw)
	}

	if jqExpr != "" && outputChanged && format == outputTable {
		return outputSettings{}, fmt.Errorf("--jq requires JSON or YAML output")
	}

	if (format != outputTable || jqExpr != "") && !commandSupportsJSON(cmd) {
		return outputSettings{}, fmt.Errorf("--output %s requires a command that supports --json", format)
	}

	settings := outputSettings{
		Format:    format,
		JQ:        jqExpr,
		OutputSet: outputChanged,
	}

	if jqExpr != "" || format == outputYAML {
		settings.Buffer = &bytes.Buffer{}
		settings.OriginalOut = cmd.OutOrStdout()
	}

	return settings, nil
}

func commandSupportsJSON(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if cmd.Flags().Lookup("json") != nil {
		return true
	}
	if cmd.InheritedFlags().Lookup("json") != nil {
		return true
	}
	return false
}

func setBoolFlag(cmd *cobra.Command, name string, value bool) error {
	if cmd.Flags().Lookup(name) != nil {
		return cmd.Flags().Set(name, fmt.Sprintf("%t", value))
	}
	if cmd.InheritedFlags().Lookup(name) != nil {
		return cmd.InheritedFlags().Set(name, fmt.Sprintf("%t", value))
	}
	return nil
}

func flagChanged(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Changed(name) {
		return true
	}
	if cmd.InheritedFlags().Changed(name) {
		return true
	}
	return false
}

func decodeJSON(payload []byte) (any, error) {
	if len(bytes.TrimSpace(payload)) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	// Disallow trailing non-whitespace content.
	if decoder.More() {
		return nil, errors.New("multiple JSON values in output")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, errors.New("multiple JSON values in output")
		}
		return nil, err
	}
	return value, nil
}

func applyJQ(value any, expr string) (any, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, err
	}
	code, err := gojq.Compile(query)
	if err != nil {
		return nil, err
	}

	iter := code.Run(value)
	results := []any{}
	for {
		item, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := item.(error); ok {
			return nil, err
		}
		results = append(results, item)
	}

	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0], nil
	default:
		return results, nil
	}
}

func writeJSONOutput(out io.Writer, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, string(payload)); err != nil {
		return err
	}
	return nil
}

func writeYAMLOutput(out io.Writer, value any) error {
	payload, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	if len(payload) == 0 || payload[len(payload)-1] != '\n' {
		payload = append(payload, '\n')
	}
	if _, err := out.Write(payload); err != nil {
		return err
	}
	return nil
}
