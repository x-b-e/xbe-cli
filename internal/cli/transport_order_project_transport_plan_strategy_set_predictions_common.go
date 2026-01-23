package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type strategySetPrediction struct {
	StrategySetID string  `json:"strategy_set_id"`
	Probability   float64 `json:"probability"`
}

func parseStrategySetPredictions(value any) []strategySetPrediction {
	if value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []strategySetPrediction:
		return typed
	case []map[string]any:
		preds := make([]strategySetPrediction, 0, len(typed))
		for _, item := range typed {
			if pred, ok := buildStrategySetPrediction(item); ok {
				preds = append(preds, pred)
			}
		}
		return preds
	case []any:
		preds := make([]strategySetPrediction, 0, len(typed))
		for _, item := range typed {
			if item == nil {
				continue
			}
			switch entry := item.(type) {
			case map[string]any:
				if pred, ok := buildStrategySetPrediction(entry); ok {
					preds = append(preds, pred)
				}
			}
		}
		return preds
	default:
		return nil
	}
}

func buildStrategySetPrediction(entry map[string]any) (strategySetPrediction, bool) {
	if entry == nil {
		return strategySetPrediction{}, false
	}

	strategySetID := predictionString(entry["strategy_set_id"])
	if strategySetID == "" {
		strategySetID = predictionString(entry["strategy-set-id"])
	}

	probability := predictionFloat(entry["probability"])

	return strategySetPrediction{
		StrategySetID: strategySetID,
		Probability:   probability,
	}, true
}

func topStrategySetPrediction(predictions []strategySetPrediction) (strategySetPrediction, bool) {
	if len(predictions) == 0 {
		return strategySetPrediction{}, false
	}

	top := predictions[0]
	for _, prediction := range predictions[1:] {
		if prediction.Probability > top.Probability {
			top = prediction
		}
	}

	return top, true
}

func predictionString(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func predictionFloat(value any) float64 {
	if value == nil {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		if f, err := typed.Float64(); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(typed, 64); err == nil {
			return f
		}
	}
	return 0
}
