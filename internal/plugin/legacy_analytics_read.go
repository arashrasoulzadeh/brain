package plugin

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
)

var analyticsEventName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]{0,63}$`)

func analyticsPageInt(input map[string]any, key string, fallback, min, max int) (int, error) {
	value, supplied := input[key]
	if !supplied {
		return fallback, nil
	}
	var n int
	switch raw := value.(type) {
	case float64:
		if raw < float64(min) || raw > float64(max) || raw != float64(int(raw)) {
			return 0, ErrInvalidWorkerRequest
		}
		n = int(raw)
	case string:
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, ErrInvalidWorkerRequest
		}
		n = parsed
	default:
		return 0, ErrInvalidWorkerRequest
	}
	if n < min || n > max {
		return 0, ErrInvalidWorkerRequest
	}
	return n, nil
}

// Reads are derived solely from the host-owned project event store. History
// returns an exact event name admitted by the resource-bound route; the
// project-wide aggregate route has its own separate capability.
func prepareAnalyticsRead(storage *SQLiteStorage, request WorkerRequest) (WorkerRequest, func(sqliteExecutor) error, error) {
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
	offset, err := analyticsPageInt(input, "offset", 0, 0, 1000000000)
	if err != nil {
		return request, nil, err
	}
	summary := request.Operation == "analytics.summary"
	maxLimit := 20
	if summary {
		maxLimit = 100
	}
	limit, err := analyticsPageInt(input, "limit", 20, 1, maxLimit)
	if err != nil {
		return request, nil, err
	}
	name, _ := input["name"].(string)
	if !summary && !analyticsEventName.MatchString(name) {
		return request, nil, ErrInvalidWorkerRequest
	}
	state, err := storage.State(request.ProjectID, "analytics", "events")
	if err != nil {
		return request, nil, err
	}
	events, _ := state["events"].([]any)
	rows := make([]map[string]any, 0, limit)
	total := 0
	if summary {
		counts := make(map[string]int)
		for _, raw := range events {
			record, ok := raw.(map[string]any)
			eventName, _ := record["name"].(string)
			if ok && record["projectId"] == request.ProjectID && analyticsEventName.MatchString(eventName) {
				counts[eventName]++
				total++
			}
		}
		names := make([]string, 0, len(counts))
		for eventName := range counts {
			names = append(names, eventName)
		}
		sort.Strings(names)
		for i := offset; i < len(names) && len(rows) < limit; i++ {
			rows = append(rows, map[string]any{"name": names[i], "count": counts[names[i]]})
		}
		result := map[string]any{"total_events": total, "total_names": len(names), "rows": rows, "offset": offset}
		if offset+len(rows) < len(names) {
			result["next_offset"] = offset + len(rows)
		}
		encoded, err := json.Marshal(map[string]any{"projectId": request.ProjectID, "result": result})
		request.Payload = encoded
		return request, nil, err
	}
	// Bound both the number of records and the serialized worker payload.
	// Older records may not have timestamps; never invent them.
	size := 0
	full := false
	for i := len(events) - 1; i >= 0; i-- {
		record, ok := events[i].(map[string]any)
		if !ok || record["projectId"] != request.ProjectID || record["name"] != name {
			continue
		}
		total++
		if total <= offset || len(rows) == limit || full {
			continue
		}
		row := map[string]any{"name": name, "fields": record["fields"]}
		if timestamp, ok := record["occurred_at"].(string); ok {
			row["occurred_at"] = timestamp
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			return request, nil, err
		}
		if size+len(encoded) > 200*1024 {
			if len(rows) == 0 {
				return request, nil, ErrInvalidWorkerRequest
			}
			full = true
			continue
		}
		size += len(encoded)
		rows = append(rows, row)
	}
	result := map[string]any{"name": name, "total_events": total, "rows": rows, "offset": offset}
	if offset+len(rows) < total {
		result["next_offset"] = offset + len(rows)
	}
	encoded, err := json.Marshal(map[string]any{"projectId": request.ProjectID, "name": name, "result": result})
	request.Payload = encoded
	return request, nil, err
}
