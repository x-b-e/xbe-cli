package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type brokerRetainerPaymentForecastScheduleEntry struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type brokerRetainerPaymentForecastRow struct {
	ID           string                                       `json:"id"`
	BrokerID     string                                       `json:"broker_id,omitempty"`
	Broker       string                                       `json:"broker,omitempty"`
	Date         string                                       `json:"date,omitempty"`
	PaymentCount int                                          `json:"payment_count,omitempty"`
	TotalAmount  float64                                      `json:"total_amount,omitempty"`
	Schedule     []brokerRetainerPaymentForecastScheduleEntry `json:"schedule,omitempty"`
}

func buildBrokerRetainerPaymentForecastRowFromSingle(resp jsonAPISingleResponse) brokerRetainerPaymentForecastRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	return buildBrokerRetainerPaymentForecastRow(resp.Data, included)
}

func buildBrokerRetainerPaymentForecastRow(resource jsonAPIResource, included map[string]jsonAPIResource) brokerRetainerPaymentForecastRow {
	schedule := parseBrokerRetainerPaymentForecastSchedule(resource.Attributes)
	total := 0.0
	for _, entry := range schedule {
		total += entry.Amount
	}

	row := brokerRetainerPaymentForecastRow{
		ID:           resource.ID,
		Date:         formatDate(stringAttr(resource.Attributes, "date")),
		PaymentCount: len(schedule),
		TotalAmount:  total,
		Schedule:     schedule,
		BrokerID:     relationshipIDFromMap(resource.Relationships, "broker"),
	}

	if row.BrokerID != "" {
		if broker, ok := included[resourceKey("brokers", row.BrokerID)]; ok {
			row.Broker = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	return row
}

func parseBrokerRetainerPaymentForecastSchedule(attrs map[string]any) []brokerRetainerPaymentForecastScheduleEntry {
	if attrs == nil {
		return nil
	}
	value, ok := attrs["schedule"]
	if !ok || value == nil {
		return nil
	}

	var entries []brokerRetainerPaymentForecastScheduleEntry
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			itemMap, ok := item.(map[string]any)
			if !ok || itemMap == nil {
				continue
			}
			entry := brokerRetainerPaymentForecastScheduleEntry{
				Date:   formatDate(stringAttr(itemMap, "date")),
				Amount: floatAttr(itemMap, "amount"),
			}
			entries = append(entries, entry)
		}
	case []map[string]any:
		for _, itemMap := range typed {
			entry := brokerRetainerPaymentForecastScheduleEntry{
				Date:   formatDate(stringAttr(itemMap, "date")),
				Amount: floatAttr(itemMap, "amount"),
			}
			entries = append(entries, entry)
		}
	}

	return entries
}

func renderBrokerRetainerPaymentForecastDetails(cmd *cobra.Command, row brokerRetainerPaymentForecastRow) error {
	out := cmd.OutOrStdout()

	if row.PaymentCount == 0 && len(row.Schedule) > 0 {
		row.PaymentCount = len(row.Schedule)
	}
	if row.TotalAmount == 0 && len(row.Schedule) > 0 {
		for _, entry := range row.Schedule {
			row.TotalAmount += entry.Amount
		}
	}

	fmt.Fprintf(out, "Broker retainer payment forecast %s\n", row.ID)
	if row.Broker != "" {
		fmt.Fprintf(out, "Broker: %s\n", row.Broker)
	} else if row.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", row.BrokerID)
	}
	if row.Date != "" {
		fmt.Fprintf(out, "Date: %s\n", row.Date)
	}
	if row.PaymentCount > 0 {
		fmt.Fprintf(out, "Payments: %d\n", row.PaymentCount)
	}
	if row.TotalAmount > 0 {
		fmt.Fprintf(out, "Total: %s\n", formatForecastAmount(row.TotalAmount))
	}

	if len(row.Schedule) == 0 {
		return nil
	}

	fmt.Fprintln(out, "Schedule:")
	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "DATE\tAMOUNT")
	for _, entry := range row.Schedule {
		fmt.Fprintf(writer, "%s\t%s\n", entry.Date, formatForecastAmount(entry.Amount))
	}
	return writer.Flush()
}

func formatForecastAmount(value float64) string {
	if value == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", value)
}
