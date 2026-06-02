<script setup>
import { computed, nextTick, onMounted, ref, watch } from "vue";
import * as echarts from "echarts";

const text = {
  en: {
    title: "Local Agent Usage",
    subtitle: "Offline-first token usage monitor",
    localOnly: "LOCAL ONLY",
    totalTokens: "Total Tokens",
    inputTokens: "Input Tokens",
    outputTokens: "Output Tokens",
    cacheRead: "Cache Read",
    cacheWrite: "Cache Write",
    requests: "Requests",
    cost: "Est. USD",
    dailyRequests: "Daily API Requests",
    toolRanking: "Tool Ranking",
    projectRanking: "Project Ranking",
    modelDetails: "Model Usage Details",
    allTools: "All Tools",
    allModels: "All Models",
    period: "Period",
    tool: "Tool",
    model: "Model",
    tokens: "Tokens",
    columnCost: "Column cost",
    rowCost: "Row cost",
    item: "Item",
    total: "Total",
    export: "Export",
    readable: "Readable",
    raw: "Raw",
    formulaTitle: "Cost estimate",
    formulaApi: "API: non-cached input uses input rates, cache read uses cached rates, output/reasoning/tool tokens use output rates.",
    formulaCodex: "Codex: keeps the largest same-day cumulative snapshot; cost is roughly (input×5 + output×30) / 1M × 0.10.",
    formulaNote: "Local estimate, not an official bill.",
  },
  "zh-CN": {
    title: "本地 Agent 用量统计",
    subtitle: "离线优先的本地 token 用量监控",
    localOnly: "LOCAL ONLY",
    totalTokens: "总 Tokens",
    inputTokens: "输入 Tokens",
    outputTokens: "输出 Tokens",
    cacheRead: "缓存读取",
    cacheWrite: "缓存写入",
    requests: "请求次数",
    cost: "预估费用(USD)",
    dailyRequests: "每日 API 请求次数",
    toolRanking: "工具排行",
    projectRanking: "项目排行",
    modelDetails: "模型用量明细",
    allTools: "全部工具",
    allModels: "全部模型",
    period: "日期",
    tool: "工具",
    model: "模型",
    tokens: "Tokens",
    columnCost: "该列预估费用",
    rowCost: "该行预估费用",
    item: "项目",
    total: "总量",
    export: "导出",
    readable: "友好数量",
    raw: "原始数量",
    formulaTitle: "费用估算",
    formulaApi: "API：非缓存输入按输入价，缓存读取按缓存价，输出/推理/工具 tokens 按输出价。",
    formulaCodex: "Codex：按当天同模型累计快照取最大值；费用约为 (输入×5 + 输出×30) / 1M × 0.10。",
    formulaNote: "本地估算，非官方账单。",
  },
};

const storage = {
  theme: "agentmeter-theme",
  token: "agentmeter-token-display",
  lang: "agentmeter-lang",
};

const lang = ref(localStorage.getItem(storage.lang) || "en");
const theme = ref(localStorage.getItem(storage.theme) || "dark");
const tokenMode = ref(localStorage.getItem(storage.token) || "compact");
const dashboard = ref(null);
const loading = ref(true);
const selectedTool = ref("");
const selectedModel = ref("");
const requestChart = ref(null);
const activeTip = ref(null);
let requestInstance;

const t = computed(() => text[lang.value]);
const compact = computed(() => tokenMode.value !== "raw");
const toolRows = computed(() => dashboard.value?.tool_ranking || []);
const projectRows = computed(() => dashboard.value?.project_ranking || []);
const modelRows = computed(() => (dashboard.value?.model_rows || []).map((row) => ({
  period: row.period,
  tool: row.tool,
  model: row.model,
  input: row.input_tokens || 0,
  output: row.output_tokens || 0,
  cacheRead: row.cache_read_tokens || 0,
  cacheWrite: row.cache_write_tokens || 0,
  reasoning: row.reasoning_tokens || 0,
  toolTokens: row.tool_tokens || 0,
  total: row.total_tokens || 0,
  cost: row.est_usd || 0,
  breakdown: row.cost_breakdown || {},
})).filter((row) => row.period !== "Total"));
const filteredRows = computed(() => modelRows.value.filter((row) =>
  (!selectedTool.value || row.tool === selectedTool.value) &&
  (!selectedModel.value || row.model === selectedModel.value)
));
const tools = computed(() => [...new Set(modelRows.value.map((row) => row.tool))].sort());
const models = computed(() => [...new Set(modelRows.value.filter((row) => !selectedTool.value || row.tool === selectedTool.value).map((row) => row.model))].sort());

function fmt(value, locale = lang.value, useCompact = compact.value) {
  const numberLocale = locale === "zh-CN" ? "zh-CN" : "en-US";
  const options = useCompact ? { notation: "compact", maximumFractionDigits: 1 } : {};
  return new Intl.NumberFormat(numberLocale, options).format(value || 0);
}

function money(value) {
  return `$${Number(value || 0).toFixed(3)}`;
}

function shortDate(value) {
  const parts = String(value || "").split("-");
  return parts.length >= 3 ? `${Number(parts[1])}-${Number(parts[2])}` : value;
}

function fullDate(value) {
  return String(value || "");
}

function projectName(value) {
  const parts = String(value || "").replace(/\\/g, "/").replace(/\/+$/, "").split("/").filter(Boolean);
  return parts[parts.length - 1] || value || "unknown";
}

function costInfo(row, key, label, tokens) {
  const amount = row?.breakdown?.[key] || 0;
  const note = row?.breakdown?.note;
  const items = [
    { key: "input", label: t.value.inputTokens, tokens: row?.input || 0, amount: row?.breakdown?.input || 0 },
    { key: "output", label: t.value.outputTokens, tokens: row?.output || 0, amount: row?.breakdown?.output || 0 },
    { key: "cache_read", label: t.value.cacheRead, tokens: row?.cacheRead || 0, amount: row?.breakdown?.cache_read || 0 },
    { key: "cache_write", label: t.value.cacheWrite, tokens: row?.cacheWrite || 0, amount: row?.breakdown?.cache_write || 0 },
    { key: "reasoning", label: "Reasoning", tokens: row?.reasoning || 0, amount: row?.breakdown?.reasoning || 0 },
  ];
  return {
    key,
    label,
    tokens,
    amount,
    items,
    rowCost: row?.breakdown?.total || row?.cost || 0,
    note: note || (amount <= 0 ? "No cost contribution estimated for this column." : ""),
  };
}

function showTip(event, row, key, label, tokens) {
  activeTip.value = {
    ...costInfo(row, key, label, tokens),
    x: event.clientX,
    y: event.clientY,
  };
}

function moveTip(event) {
  if (!activeTip.value) return;
  activeTip.value = {
    ...activeTip.value,
    x: event.clientX,
    y: event.clientY,
  };
}

function hideTip() {
  activeTip.value = null;
}

async function load() {
  loading.value = true;
  const res = await fetch("/api/dashboard?grain=day");
  dashboard.value = await res.json();
  loading.value = false;
  await nextTick();
  renderCharts();
}

function renderCharts() {
  renderRequestChart();
}

function renderRequestChart() {
  if (!requestChart.value || !dashboard.value) return;
  requestInstance ||= echarts.init(requestChart.value);
  const data = dashboard.value.trend || [];
  requestInstance.setOption({
    animationDuration: 700,
    animationDurationUpdate: 480,
    animationEasing: "cubicOut",
    animationEasingUpdate: "cubicOut",
    backgroundColor: "transparent",
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "line", lineStyle: { color: getCss("--line"), width: 1.5 } },
      formatter: (items) => {
        const item = Array.isArray(items) ? items[0] : items;
        const row = data[item?.dataIndex] || {};
        return `${fullDate(row.key)}<br/>${t.value.requests}: ${item?.value || 0}`;
      },
    },
    grid: { left: 54, right: 30, top: 34, bottom: 42 },
    xAxis: { type: "category", data: data.map((d) => shortDate(d.key)), axisLabel: { color: getCss("--muted") }, axisLine: { lineStyle: { color: getCss("--line") } } },
    yAxis: { type: "value", minInterval: 1, axisLabel: { color: getCss("--muted") }, splitLine: { lineStyle: { color: getCss("--table-border") } } },
    series: [{
      name: t.value.requests,
      type: "line",
      smooth: 0.18,
      smoothMonotone: "x",
      showSymbol: true,
      symbol: "circle",
      symbolSize: 5,
      data: data.map((d) => d.request_count),
      label: { show: true, position: "top", distance: 8, color: getCss("--muted"), fontSize: 11, formatter: ({ value }) => value },
      labelLayout: { hideOverlap: true },
      areaStyle: { color: "rgba(120,168,255,.10)" },
      lineStyle: { color: "#78a8ff", width: 2.6 },
      itemStyle: { color: "#78a8ff", borderColor: getCss("--panel"), borderWidth: 2 },
      emphasis: { focus: "series", scale: 1.35 },
    }],
  });
}

function getCss(name) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

function toggleTheme() {
  theme.value = theme.value === "light" ? "dark" : "light";
}

function toggleLang() {
  lang.value = lang.value === "en" ? "zh-CN" : "en";
}

function toggleTokenMode() {
  tokenMode.value = tokenMode.value === "raw" ? "compact" : "raw";
}

watch(theme, (value) => {
  document.documentElement.dataset.theme = value;
  document.body.dataset.theme = value;
  localStorage.setItem(storage.theme, value);
  nextTick(renderCharts);
}, { immediate: true });

watch(lang, (value) => {
  document.documentElement.lang = value;
  localStorage.setItem(storage.lang, value);
  nextTick(renderCharts);
});

watch(tokenMode, (value) => localStorage.setItem(storage.token, value));

onMounted(() => {
  load();
  window.addEventListener("resize", () => {
    requestInstance?.resize();
  });
});
</script>

<template>
  <main v-if="dashboard && !loading" class="page">
    <header class="topbar">
      <div>
        <p class="eyebrow">AgentMeter</p>
        <h1>{{ t.title }}</h1>
        <p class="subtitle">{{ t.subtitle }}</p>
      </div>
      <div class="actions">
        <span class="status-pill">{{ t.localOnly }}</span>
        <button class="theme-switch" @click="toggleTheme">
          <span class="theme-icon theme-icon-dark">◐</span>
          <span class="theme-icon theme-icon-light">☼</span>
        </button>
        <button class="token-display-toggle" @click="toggleTokenMode">
          {{ tokenMode === "raw" ? t.raw : t.readable }}
        </button>
        <button class="language-switch" :class="lang === 'en' ? 'is-en' : 'is-zh'" @click="toggleLang">
          <span>EN</span>
          <strong>中文</strong>
        </button>
        <a class="export" href="/export.csv">{{ t.export }}</a>
      </div>
    </header>

    <section class="stats">
      <article><span>{{ t.totalTokens }}</span><strong>{{ fmt(dashboard.summary.total_tokens) }}</strong></article>
      <article><span>{{ t.cost }}</span><strong>{{ money(dashboard.summary.cost_amount) }}</strong></article>
      <article><span>{{ t.inputTokens }}</span><strong>{{ fmt(dashboard.summary.input_tokens) }}</strong></article>
      <article><span>{{ t.cacheRead }}</span><strong>{{ fmt(dashboard.summary.cache_read_tokens) }}</strong></article>
    </section>

    <section class="chart-panel">
      <div class="panel-title">
        <h2>{{ t.dailyRequests }}</h2>
        <span>{{ t.requests }}</span>
      </div>
      <div ref="requestChart" class="vue-chart compact"></div>
    </section>

    <section class="grid">
      <article class="panel">
        <div class="panel-title"><h2>{{ t.toolRanking }}</h2></div>
        <div class="ranking">
          <div v-for="item in toolRows" :key="item.name" class="rank-row">
            <span class="rank-name">{{ item.name }}</span>
            <span class="rank-track"><span class="rank-fill" :style="{ '--w': ((item.total_tokens / (toolRows[0]?.total_tokens || 1)) * 100) + '%' }"></span></span>
            <span class="rank-value">{{ fmt(item.total_tokens) }}</span>
          </div>
        </div>
      </article>

      <article class="panel">
        <div class="panel-title"><h2>{{ t.projectRanking }}</h2></div>
        <div class="ranking project-ranking">
          <div v-for="item in projectRows" :key="item.name" class="rank-row">
            <span class="rank-name" :title="item.name">{{ projectName(item.name) }}</span>
            <span class="rank-track"><span class="rank-fill" :style="{ '--w': ((item.total_tokens / (projectRows[0]?.total_tokens || 1)) * 100) + '%' }"></span></span>
            <span class="rank-value">{{ fmt(item.total_tokens) }}</span>
          </div>
        </div>
      </article>

      <article class="panel wide">
        <div class="panel-title">
          <div><h2>{{ t.modelDetails }}</h2></div>
          <div class="model-filters">
            <select v-model="selectedTool">
              <option value="">{{ t.allTools }}</option>
              <option v-for="tool in tools" :key="tool" :value="tool">{{ tool }}</option>
            </select>
            <select v-model="selectedModel">
              <option value="">{{ t.allModels }}</option>
              <option v-for="model in models" :key="model" :value="model">{{ model }}</option>
            </select>
          </div>
        </div>
        <div class="usage-table-wrap">
          <table class="usage-table">
            <thead>
              <tr>
                <th>{{ t.period }}</th>
                <th>{{ t.tool }}</th>
                <th>{{ t.model }}</th>
                <th class="cost-column">{{ t.cost }}</th>
                <th>{{ t.inputTokens }}</th>
                <th>{{ t.outputTokens }}</th>
                <th>{{ t.cacheRead }}</th>
                <th>{{ t.totalTokens }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in filteredRows" :key="`${row.period}-${row.tool}-${row.model}`" class="model-data-row">
                <td>{{ row.period }}</td>
                <td>{{ row.tool }}</td>
                <td>{{ row.model }}</td>
                <td class="cost-column">{{ money(row.cost) }}</td>
                <td class="cost-cell" @mouseenter="showTip($event, row, 'input', t.inputTokens, row.input)" @mousemove="moveTip" @mouseleave="hideTip">{{ fmt(row.input) }}</td>
                <td class="cost-cell" @mouseenter="showTip($event, row, 'output', t.outputTokens, row.output)" @mousemove="moveTip" @mouseleave="hideTip">{{ fmt(row.output) }}</td>
                <td class="cost-cell" @mouseenter="showTip($event, row, 'cache_read', t.cacheRead, row.cacheRead)" @mousemove="moveTip" @mouseleave="hideTip">{{ fmt(row.cacheRead) }}</td>
                <td class="cost-cell" @mouseenter="showTip($event, row, 'total', t.totalTokens, row.total)" @mousemove="moveTip" @mouseleave="hideTip">{{ fmt(row.total) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </section>

    <div v-if="activeTip" class="cost-tooltip" :style="{ left: `${activeTip.x + 14}px`, top: `${activeTip.y + 14}px` }">
      <div class="cost-tooltip-head">
        <span>{{ activeTip.label }}</span>
        <strong>{{ money(activeTip.amount) }}</strong>
      </div>
      <div class="cost-tooltip-grid">
        <span>{{ t.tokens }}</span>
        <strong>{{ fmt(activeTip.tokens) }}</strong>
        <span>{{ t.columnCost }}</span>
        <strong>{{ money(activeTip.amount) }}</strong>
        <span>{{ t.rowCost }}</span>
        <strong>{{ money(activeTip.rowCost) }}</strong>
      </div>
      <div class="cost-tooltip-items">
        <div class="cost-tooltip-item cost-tooltip-item-head">
          <span>{{ t.item }}</span>
          <span>{{ t.tokens }}</span>
          <span>{{ t.cost }}</span>
        </div>
        <div v-for="item in activeTip.items" :key="item.key" class="cost-tooltip-item" :class="{ active: item.key === activeTip.key }">
          <span>{{ item.label }}</span>
          <span>{{ fmt(item.tokens) }}</span>
          <strong>{{ money(item.amount) }}</strong>
        </div>
      </div>
      <p v-if="activeTip.note">{{ activeTip.note }}</p>
    </div>
  </main>
</template>
