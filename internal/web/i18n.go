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
		"page_controls":    "页面控制",
		"language":         "语言",
		"theme":            "主题",
		"token_display":    "Token 数量显示",
		"token_compact":    "友好数量",
		"token_raw":        "原始数量",
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
		"all_tools":        "全部工具",
		"all_models":       "全部模型",
		"period":           "日期",
		"tool":             "工具",
		"model":            "模型",
		"unresolved":       "未解析",
		"tool_tokens":      "工具 Tokens",
		"est_usd":          "预估费用(USD)",
		"pricing_title":    "费用估算",
		"pricing_api":      "优先使用日志中的官方 billed cost；缺失时按模型 token 费率做本地估算。",
		"pricing_codex":    "Codex：Input Tokens 展示为 input_tokens - cache_read_tokens；Cache Read 单独展示，并按模型缓存读取费率估算。",
		"pricing_note":     "本地估算可能与最终账单存在差异。",
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
		"page_controls":    "Page controls",
		"language":         "Language",
		"theme":            "Theme",
		"token_display":    "Token display",
		"token_compact":    "Readable",
		"token_raw":        "Raw",
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
		"all_tools":        "All Tools",
		"all_models":       "All Models",
		"period":           "Period",
		"tool":             "Tool",
		"model":            "Model",
		"unresolved":       "Unresolved",
		"tool_tokens":      "Tool Tokens",
		"est_usd":          "Est. USD",
		"pricing_title":    "Cost estimate",
		"pricing_api":      "Official billed cost is used when present; otherwise cost is estimated from model token rates.",
		"pricing_codex":    "Codex: Input Tokens display input_tokens - cache_read_tokens; Cache Read is shown separately and estimated with the model cached-read rate.",
		"pricing_note":     "Local estimates may differ from final billing.",
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
