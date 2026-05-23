package usage

import "time"

type Event struct {
	ToolName         string
	ModelName        string
	SessionID        string
	ProjectPath      string
	InputTokens      int64
	OutputTokens     int64
	CacheReadTokens  int64
	CacheWriteTokens int64
	ReasoningTokens  int64
	ToolTokens       int64
	TotalTokens      int64
	CostAmount       float64
	OccurredAt       time.Time
	SourceFile       string
}

type Grain string

const (
	GrainDay   Grain = "day"
	GrainWeek  Grain = "week"
	GrainMonth Grain = "month"
)

type Bucket struct {
	Key              string  `json:"key"`
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	CacheReadTokens  int64   `json:"cache_read_tokens"`
	CacheWriteTokens int64   `json:"cache_write_tokens"`
	ReasoningTokens  int64   `json:"reasoning_tokens"`
	ToolTokens       int64   `json:"tool_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	RequestCount     int64   `json:"request_count"`
	CostAmount       float64 `json:"cost_amount"`
}

type RankingItem struct {
	Name             string  `json:"name"`
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	CacheReadTokens  int64   `json:"cache_read_tokens"`
	CacheWriteTokens int64   `json:"cache_write_tokens"`
	ReasoningTokens  int64   `json:"reasoning_tokens"`
	ToolTokens       int64   `json:"tool_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	RequestCount     int64   `json:"request_count"`
	CostAmount       float64 `json:"cost_amount"`
}

type Summary struct {
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	CacheReadTokens  int64   `json:"cache_read_tokens"`
	CacheWriteTokens int64   `json:"cache_write_tokens"`
	ReasoningTokens  int64   `json:"reasoning_tokens"`
	ToolTokens       int64   `json:"tool_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	RequestCount     int64   `json:"request_count"`
	CostAmount       float64 `json:"cost_amount"`
	ModelCount       int     `json:"model_count"`
	ToolCount        int     `json:"tool_count"`
}

type Dashboard struct {
	Summary        Summary       `json:"summary"`
	Trend          []Bucket      `json:"trend"`
	ToolRanking    []RankingItem `json:"tool_ranking"`
	ModelRanking   []RankingItem `json:"model_ranking"`
	ProjectRanking []RankingItem `json:"project_ranking"`
}
