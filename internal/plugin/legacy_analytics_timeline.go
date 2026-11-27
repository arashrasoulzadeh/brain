package plugin

import (
	"encoding/json"
	"time"
)

type analyticsWindow struct {
	from, until time.Time
	bounded     bool
}

func readAnalyticsWindow(input map[string]any, required bool) (analyticsWindow, error) {
	from, hasFrom := input["from"]
	until, hasUntil := input["until"]
	if !hasFrom && !hasUntil && !required {
		return analyticsWindow{}, nil
	}
	startText, startOK := from.(string)
	endText, endOK := until.(string)
	if !startOK || !endOK || len(startText) > 40 || len(endText) > 40 {
		return analyticsWindow{}, ErrInvalidWorkerRequest
	}
	start, err := time.Parse(time.RFC3339Nano, startText)
	if err != nil {
		return analyticsWindow{}, ErrInvalidWorkerRequest
	}
	end, err := time.Parse(time.RFC3339Nano, endText)
	if err != nil || !end.After(start) || end.Sub(start) > 366*24*time.Hour {
		return analyticsWindow{}, ErrInvalidWorkerRequest
	}
	return analyticsWindow{from: start.UTC(), until: end.UTC(), bounded: true}, nil
}

func eventTime(record map[string]any) (time.Time, bool) {
	value, ok := record["occurred_at"].(string)
	if !ok {
		return time.Time{}, false
	}
	stamp, err := time.Parse(time.RFC3339Nano, value)
	return stamp.UTC(), err == nil
}

func (w analyticsWindow) contains(stamp time.Time, known bool) bool {
	return !w.bounded || (known && !stamp.Before(w.from) && stamp.Before(w.until))
}

func prepareAnalyticsTimeline(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
	if storage == nil {
		return request, nil, ErrStoredResourceMissing
	}
	var input map[string]any
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	if project, exists := input["projectId"]; exists && project != request.ProjectID {
		return request, nil, ErrInvalidWorkerRequest
	}
	window, err := readAnalyticsWindow(input, true)
	if err != nil {
		return request, nil, err
	}
	bucket, _ := input["bucket"].(string)
	var step time.Duration
	switch bucket {
	case "hour":
		step = time.Hour
	case "day":
		step = 24 * time.Hour
	default:
		return request, nil, ErrInvalidWorkerRequest
	}
	name, supplied := input["name"]
	if supplied {
		value, ok := name.(string)
		if !ok || !analyticsEventName.MatchString(value) {
			return request, nil, ErrInvalidWorkerRequest
		}
	}
	first := window.from.Truncate(step)
	// The last UTC bucket may be partial; the range remains half-open.
	if window.until.Sub(first) > 168*step {
		return request, nil, ErrInvalidWorkerRequest
	}
	type bucketRow struct {
		Start string `json:"start"`
		Count int    `json:"count"`
	}
	rows := make([]bucketRow, 0)
	for cursor := first; cursor.Before(window.until); cursor = cursor.Add(step) {
		rows = append(rows, bucketRow{Start: cursor.Format(time.RFC3339)})
	}
	state, err := storage.State(request.ProjectID, "analytics", "events")
	if err != nil {
		return request, nil, err
	}
	events, _ := state["events"].([]any)
	total, undated := 0, 0
	for _, raw := range events {
		record, ok := raw.(map[string]any)
		eventName, _ := record["name"].(string)
		if !ok || record["projectId"] != request.ProjectID || !analyticsEventName.MatchString(eventName) || (supplied && name != eventName) {
			continue
		}
		stamp, known := eventTime(record)
		if !known {
			undated++
			continue
		}
		if !window.contains(stamp, known) {
			continue
		}
		rows[int(stamp.Sub(first)/step)].Count++
		total++
	}
	result := map[string]any{"from": window.from.Format(time.RFC3339Nano), "until": window.until.Format(time.RFC3339Nano), "bucket": bucket, "rows": rows, "total_events": total, "undated_events": undated}
	if supplied {
		result["name"] = name
	}
	payload, err := json.Marshal(map[string]any{"projectId": request.ProjectID, "result": result})
	request.Payload = payload
	return request, nil, err
}
