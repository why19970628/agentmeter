package usage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ScanOptions struct {
	DefaultTool string
	MaxFileSize int64
}

func ScanDir(root string, opts ScanOptions) ([]Event, error) {
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = 16 << 20
	}

	var events []Event
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !looksLikeUsageFile(path) {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() > opts.MaxFileSize {
			return nil
		}

		fileEvents, err := scanFile(path, opts)
		if err != nil {
			return nil
		}
		events = append(events, fileEvents...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

func scanFile(path string, opts ScanOptions) ([]Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, nil
	}
	if trimmed[0] == '[' || trimmed[0] == '{' {
		if events, err := parseJSONDocument(data, path, opts); err == nil && len(events) > 0 {
			return events, nil
		}
	}
	return parseJSONLines(bytes.NewReader(data), path, opts)
}

func parseJSONDocument(data []byte, path string, opts ScanOptions) ([]Event, error) {
	var arr []map[string]any
	if err := json.Unmarshal(data, &arr); err == nil {
		return parseRecords(arr, path, opts), nil
	}

	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	if items, ok := asObjectSlice(obj["events"]); ok {
		return parseRecords(items, path, opts), nil
	}
	if event, ok := recordToEvent(obj, path, opts); ok {
		return []Event{event}, nil
	}
	return nil, errors.New("no usage records")
}

func parseJSONLines(r io.Reader, path string, opts ScanOptions) ([]Event, error) {
	var records []map[string]any
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return parseRecords(records, path, opts), nil
}

type scanContext struct {
	latestModel      string
	latestProject    string
	modelBySession   map[string]string
	projectBySession map[string]string
}

func (c *scanContext) remember(record map[string]any) {
	sessionID := stringAt(record, "session_id", "sessionId", "conversation_id")

	model := modelNameFrom(record)
	if model != "" {
		c.latestModel = model
		if sessionID != "" {
			c.modelBySession[sessionID] = model
		}
	}

	project := projectPathFrom(record)
	if project != "" {
		c.latestProject = project
		if sessionID != "" {
			c.projectBySession[sessionID] = project
		}
	}
}

func (c scanContext) lookupModel(sessionID string) string {
	if sessionID != "" {
		if model := c.modelBySession[sessionID]; model != "" {
			return model
		}
	}
	return c.latestModel
}

func (c scanContext) lookupProject(sessionID string) string {
	if sessionID != "" {
		if project := c.projectBySession[sessionID]; project != "" {
			return project
		}
	}
	return c.latestProject
}

func parseRecords(records []map[string]any, path string, opts ScanOptions) []Event {
	events := make([]Event, 0, len(records))
	context := scanContext{modelBySession: map[string]string{}, projectBySession: map[string]string{}}
	for _, record := range records {
		context.remember(record)
	}
	for _, record := range records {
		if event, ok := recordToEvent(record, path, opts); ok {
			if event.ModelName == "" {
				event.ModelName = context.lookupModel(event.SessionID)
			}
			if event.ProjectPath == "" {
				event.ProjectPath = context.lookupProject(event.SessionID)
			}
			events = append(events, event)
		}
	}
	return events
}

func recordToEvent(record map[string]any, path string, opts ScanOptions) (Event, bool) {
	usage := objectAt(record, "usage")
	input := intAt(record, "input_tokens", "prompt_tokens")
	output := intAt(record, "output_tokens", "completion_tokens")
	cacheRead := intAt(record, "cache_read_tokens", "cache_read_input_tokens", "cached_input_tokens", "cached_content_token_count")
	cacheWrite := intAt(record, "cache_write_tokens", "cache_creation_tokens", "cache_creation_input_tokens")
	reasoning := intAt(record, "reasoning_tokens", "thoughts_token_count")
	toolTokens := intAt(record, "tool_tokens", "tool_token_count")
	total := intAt(record, "total_tokens")
	if input == 0 {
		input = nestedIntAt(record, "input_tokens", "prompt_tokens")
	}
	if output == 0 {
		output = nestedIntAt(record, "output_tokens", "completion_tokens")
	}
	if cacheRead == 0 {
		cacheRead = nestedIntAt(record, "cache_read_tokens", "cache_read_input_tokens", "cached_input_tokens", "cached_content_token_count")
	}
	if cacheWrite == 0 {
		cacheWrite = nestedIntAt(record, "cache_write_tokens", "cache_creation_tokens", "cache_creation_input_tokens")
	}
	if reasoning == 0 {
		reasoning = nestedIntAt(record, "reasoning_tokens", "thoughts_token_count")
	}
	if toolTokens == 0 {
		toolTokens = nestedIntAt(record, "tool_tokens", "tool_token_count")
	}
	if total == 0 {
		total = nestedIntAt(record, "total_tokens")
	}

	if usage != nil {
		input = firstNonZero(input, intAt(usage, "input_tokens", "prompt_tokens"))
		output = firstNonZero(output, intAt(usage, "output_tokens", "completion_tokens"))
		cacheRead = firstNonZero(cacheRead, intAt(usage, "cache_read_tokens", "cache_read_input_tokens", "cached_input_tokens", "cached_content_token_count"))
		cacheWrite = firstNonZero(cacheWrite, intAt(usage, "cache_write_tokens", "cache_creation_tokens", "cache_creation_input_tokens"))
		reasoning = firstNonZero(reasoning, intAt(usage, "reasoning_tokens", "thoughts_token_count"))
		toolTokens = firstNonZero(toolTokens, intAt(usage, "tool_tokens", "tool_token_count"))
		total = firstNonZero(total, intAt(usage, "total_tokens"))
	}
	cacheRead = firstNonZero(cacheRead, tokenDetail(record, "input_tokens_details", "cached_tokens"))
	cacheWrite = firstNonZero(cacheWrite, cacheCreationTokens(record))
	reasoning = firstNonZero(reasoning, tokenDetail(record, "output_tokens_details", "reasoning_tokens"))
	if total == 0 {
		total = input + output + cacheWrite + reasoning + toolTokens
	}
	if total == 0 && cacheRead == 0 {
		return Event{}, false
	}

	occurredAt := timeAt(record, "occurred_at", "timestamp", "created_at", "time")
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}

	tool := stringAt(record, "tool_name", "tool", "source")
	if tool == "" {
		tool = opts.DefaultTool
	}
	if tool == "" {
		tool = inferToolFromPath(path)
	}

	return Event{
		ToolName:         tool,
		ModelName:        modelNameFrom(record),
		SessionID:        stringAt(record, "session_id", "sessionId", "conversation_id"),
		ProjectPath:      projectPathFrom(record),
		InputTokens:      input,
		OutputTokens:     output,
		CacheReadTokens:  cacheRead,
		CacheWriteTokens: cacheWrite,
		ReasoningTokens:  reasoning,
		ToolTokens:       toolTokens,
		TotalTokens:      total,
		CostAmount:       floatAt(record, "cost", "cost_amount"),
		OccurredAt:       occurredAt,
		SourceFile:       path,
	}, true
}

func modelNameFrom(record map[string]any) string {
	return firstNonEmpty(
		stringAt(record, "model_name", "model", "model_id", "modelId", "model_slug", "modelSlug"),
		nestedStringAt(record, "model_name", "model", "model_id", "modelId", "model_slug", "modelSlug"),
	)
}

func projectPathFrom(record map[string]any) string {
	return firstNonEmpty(
		stringAt(record, projectPathKeys()...),
		nestedStringAt(record, projectPathKeys()...),
	)
}

func projectPathKeys() []string {
	return []string{
		"project_path",
		"project",
		"cwd",
		"workspace_path",
		"workspacePath",
		"current_working_directory",
		"currentWorkingDirectory",
		"root_path",
		"rootPath",
		"repo_path",
		"repoPath",
	}
}

func tokenDetail(record map[string]any, objectKey string, valueKey string) int64 {
	if value := tokenDetailIn(record, objectKey, valueKey, 0); value != 0 {
		return value
	}
	return 0
}

func tokenDetailIn(record map[string]any, objectKey string, valueKey string, depth int) int64 {
	if depth > 4 {
		return 0
	}
	if obj := objectAt(record, objectKey); obj != nil {
		if value := intAt(obj, valueKey); value != 0 {
			return value
		}
	}
	for _, raw := range record {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if value := tokenDetailIn(child, objectKey, valueKey, depth+1); value != 0 {
			return value
		}
	}
	return 0
}

func cacheCreationTokens(record map[string]any) int64 {
	if value := cacheCreationTokensIn(record, 0); value != 0 {
		return value
	}
	return 0
}

func cacheCreationTokensIn(record map[string]any, depth int) int64 {
	if depth > 4 {
		return 0
	}
	if obj := objectAt(record, "cache_creation"); obj != nil {
		return intAt(obj, "ephemeral_5m_input_tokens") + intAt(obj, "ephemeral_1h_input_tokens")
	}
	for _, raw := range record {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if value := cacheCreationTokensIn(child, depth+1); value != 0 {
			return value
		}
	}
	return 0
}

func looksLikeUsageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json" || ext == ".jsonl" || ext == ".log"
}

func inferToolFromPath(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "codex"):
		return "codex"
	case strings.Contains(lower, "claude"):
		return "claude-code"
	case strings.Contains(lower, "cursor"):
		return "cursor"
	case strings.Contains(lower, "gemini"):
		return "gemini"
	default:
		return "local-agent"
	}
}

func objectAt(record map[string]any, key string) map[string]any {
	value, ok := record[key].(map[string]any)
	if !ok {
		return nil
	}
	return value
}

func asObjectSlice(value any) ([]map[string]any, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if ok {
			out = append(out, obj)
		}
	}
	return out, len(out) > 0
}

func stringAt(record map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := record[key].(string); ok {
			return value
		}
	}
	return ""
}

func nestedStringAt(record map[string]any, keys ...string) string {
	return nestedStringAtDepth(record, 0, keys...)
}

func nestedStringAtDepth(record map[string]any, depth int, keys ...string) string {
	if depth > 4 {
		return ""
	}
	if value := stringAt(record, keys...); value != "" {
		return value
	}
	for _, raw := range record {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if value := nestedStringAtDepth(child, depth+1, keys...); value != "" {
			return value
		}
	}
	return ""
}

func intAt(record map[string]any, keys ...string) int64 {
	for _, key := range keys {
		switch value := record[key].(type) {
		case float64:
			return int64(value)
		case int64:
			return value
		case json.Number:
			n, _ := value.Int64()
			return n
		}
	}
	return 0
}

func nestedIntAt(record map[string]any, keys ...string) int64 {
	return nestedIntAtDepth(record, 0, keys...)
}

func nestedIntAtDepth(record map[string]any, depth int, keys ...string) int64 {
	if depth > 4 {
		return 0
	}
	if value := intAt(record, keys...); value != 0 {
		return value
	}
	for _, raw := range record {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if value := nestedIntAtDepth(child, depth+1, keys...); value != 0 {
			return value
		}
	}
	return 0
}

func floatAt(record map[string]any, keys ...string) float64 {
	for _, key := range keys {
		switch value := record[key].(type) {
		case float64:
			return value
		case int64:
			return float64(value)
		case json.Number:
			n, _ := value.Float64()
			return n
		}
	}
	return 0
}

func timeAt(record map[string]any, keys ...string) time.Time {
	for _, key := range keys {
		value, ok := record[key].(string)
		if !ok || value == "" {
			continue
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			t, err := time.Parse(layout, value)
			if err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func firstNonZero(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
