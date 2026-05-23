package web

type Lang string

const (
	LangZH Lang = "zh-CN"
	LangEN Lang = "en"
)

var labels = map[Lang]map[string]string{
	LangZH: {
		"html_lang":        "zh-CN",
		"title":            "本地 Agent 用量统计",
		"subtitle":         "离线优先的本地 token 用量监控",
		"grain_label":      "统计粒度",
		"day":              "按天",
		"week":             "按星期",
		"month":            "按月",
		"language":         "语言",
		"export":           "导出",
		"total_tokens":     "总 Tokens",
		"input_tokens":     "输入 Tokens",
		"output_tokens":    "输出 Tokens",
		"cache_read":       "缓存读取",
		"cache_write":      "缓存写入",
		"reasoning_tokens": "推理 Tokens",
		"request_count":    "请求次数",
		"model_count":      "模型数",
		"token_trend":      "Token 用量趋势",
		"unit_generated":   "单位：tokens · 生成时间",
		"daily_requests":   "每日 API 请求次数",
		"by_day":           "按自然日统计",
		"tool_ranking":     "工具排行",
		"project_ranking":  "项目路径排行",
		"model_details":    "模型用量明细",
		"model_subtitle":   "按每日 + 工具 + 模型聚合",
		"period":           "日期",
		"tool":             "工具",
		"model":            "模型",
		"unresolved":       "未解析",
		"tool_tokens":      "工具 Tokens",
		"est_usd":          "预估费用(USD)",
		"empty_models":     "暂无模型用量数据",
	},
	LangEN: {
		"html_lang":        "en",
		"title":            "Local Agent Usage",
		"subtitle":         "Offline-first token usage monitor",
		"grain_label":      "Time grain",
		"day":              "Day",
		"week":             "Week",
		"month":            "Month",
		"language":         "Language",
		"export":           "Export",
		"total_tokens":     "Total Tokens",
		"input_tokens":     "Input Tokens",
		"output_tokens":    "Output Tokens",
		"cache_read":       "Cache Read",
		"cache_write":      "Cache Write",
		"reasoning_tokens": "Reasoning Tokens",
		"request_count":    "Requests",
		"model_count":      "Models",
		"token_trend":      "Token Usage Trend",
		"unit_generated":   "Unit: tokens · Generated",
		"daily_requests":   "Daily API Requests",
		"by_day":           "Grouped by day",
		"tool_ranking":     "Tool Ranking",
		"project_ranking":  "Project Ranking",
		"model_details":    "Model Usage Details",
		"model_subtitle":   "Grouped by day + tool + model",
		"period":           "Period",
		"tool":             "Tool",
		"model":            "Model",
		"unresolved":       "Unresolved",
		"tool_tokens":      "Tool Tokens",
		"est_usd":          "Est. USD",
		"empty_models":     "No model usage data",
	},
}

func parseLang(raw string) Lang {
	if raw == string(LangZH) {
		return LangZH
	}
	return LangEN
}

func translate(lang Lang, key string) string {
	if value := labels[lang][key]; value != "" {
		return value
	}
	if value := labels[LangEN][key]; value != "" {
		return value
	}
	return key
}
