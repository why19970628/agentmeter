const pageLang = document.documentElement.lang === "zh-CN" ? "zh-CN" : "en";
const numberLocale = pageLang === "zh-CN" ? "zh-CN" : "en-US";
const themeStorageKey = "agentmeter-theme";
const tokenDisplayStorageKey = "agentmeter-token-display";
const text = {
  en: {
    noTokens: "No local token records found",
    noData: "No data",
    noRequests: "No daily request data",
    requestSuffix: "requests",
    requestChartLabel: "Daily API request count line chart",
    tokenCompact: "Readable",
    tokenRaw: "Raw",
  },
  "zh-CN": {
    noTokens: "未读取到本地 token 记录",
    noData: "暂无数据",
    noRequests: "暂无每日请求数据",
    requestSuffix: "次请求",
    requestChartLabel: "每日 API 请求次数折线图",
    tokenCompact: "友好数量",
    tokenRaw: "原始数量",
  },
};

function readSavedTheme() {
  try {
    return localStorage.getItem(themeStorageKey);
  } catch (_) {
    return "";
  }
}

function saveTheme(theme) {
  try {
    localStorage.setItem(themeStorageKey, theme);
  } catch (_) {
    // Theme persistence is optional when storage is unavailable.
  }
}

function applyTheme(theme) {
  const nextTheme = theme === "light" ? "light" : "dark";
  document.documentElement.dataset.theme = nextTheme;
  document.body.dataset.theme = nextTheme;
  const toggle = document.querySelector("#themeToggle");
  if (toggle) {
    toggle.setAttribute("aria-pressed", String(nextTheme === "light"));
  }
}

function setupThemeToggle() {
  applyTheme(readSavedTheme());
  const toggle = document.querySelector("#themeToggle");
  if (!toggle) return;
  toggle.addEventListener("click", () => {
    const nextTheme = document.body.dataset.theme === "light" ? "dark" : "light";
    applyTheme(nextTheme);
    saveTheme(nextTheme);
  });
}

function formatNumber(value) {
  return new Intl.NumberFormat(numberLocale).format(value || 0);
}

function formatCompact(value) {
  return new Intl.NumberFormat(numberLocale, { notation: "compact", maximumFractionDigits: 1 }).format(value || 0);
}

function formatTokenValue(value, mode) {
  return mode === "raw" ? formatNumber(value) : formatCompact(value);
}

function parseNumber(value) {
  const parsed = Number(value || 0);
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatShortDate(value) {
  const parts = String(value || "").split("-");
  if (parts.length >= 3) {
    return `${Number(parts[1])}-${Number(parts[2])}`;
  }
  return value;
}

function formatRankingName(name, kind) {
  const value = String(name || "");
  if (kind === "project") {
    const normalized = value.replace(/\\/g, "/").replace(/\/+$/, "");
    const parts = normalized.split("/").filter(Boolean);
    return parts[parts.length - 1] || value;
  }
  return value;
}

function renderTrend() {
  const chart = document.querySelector("#trendChart");
  if (!chart) return;
  const data = JSON.parse(chart.dataset.values || "[]");
  const max = Math.max(...data.map((item) => item.total_tokens), 1);
  chart.innerHTML = "";
  if (!data.length) {
    chart.innerHTML = `<p class="empty">${text[pageLang].noTokens}</p>`;
    return;
  }
  const ticks = document.createElement("div");
  ticks.className = "bar-axis";
  for (const ratio of [1, 0.75, 0.5, 0.25, 0]) {
    const tick = document.createElement("span");
    tick.style.setProperty("--y", `${(1 - ratio) * 100}%`);
    tick.textContent = formatCompact(Math.round(max * ratio));
    ticks.appendChild(tick);
  }
  chart.appendChild(ticks);
  for (const item of data) {
    const wrap = document.createElement("div");
    wrap.className = "bar-item";
    const bar = document.createElement("div");
    bar.className = "bar";
    bar.style.setProperty("--h", `${Math.max(4, (item.total_tokens / max) * 100)}%`);
    bar.dataset.value = formatCompact(item.total_tokens);
    bar.title = `${item.key}: ${formatNumber(item.total_tokens)} tokens`;
    const label = document.createElement("div");
    label.className = "bar-label";
    label.innerHTML = `<span>${formatShortDate(item.key)}</span>`;
    wrap.appendChild(bar);
    wrap.appendChild(label);
    chart.appendChild(wrap);
  }
}

function renderRankings() {
  for (const root of document.querySelectorAll(".ranking")) {
    const kind = root.dataset.rankingKind || "";
    const data = JSON.parse(root.dataset.ranking || "[]");
    const rows = kind === "project" ? data : data.slice(0, 10);
    const max = Math.max(...data.map((item) => item.total_tokens), 1);
    root.innerHTML = "";
    if (!rows.length) {
      root.innerHTML = `<p class="empty">${text[pageLang].noData}</p>`;
      continue;
    }
    for (const item of rows) {
      const name = formatRankingName(item.name, kind);
      const row = document.createElement("div");
      row.className = "rank-row";
      row.innerHTML = `
        <span class="rank-name" title="${item.name}">${name}</span>
        <span class="rank-track"><span class="rank-fill" style="--w:${(item.total_tokens / max) * 100}%"></span></span>
        <span class="rank-value">${formatNumber(item.total_tokens)}</span>
      `;
      root.appendChild(row);
    }
  }
}

function renderRequests() {
  const chart = document.querySelector("#requestChart");
  if (!chart) return;
  const data = JSON.parse(chart.dataset.values || "[]");
  chart.innerHTML = "";
  if (!data.length) {
    chart.innerHTML = `<p class="empty">${text[pageLang].noRequests}</p>`;
    return;
  }

  const width = 1000;
  const height = 260;
  const pad = { top: 34, right: 22, bottom: 46, left: 58 };
  const max = Math.max(...data.map((item) => item.request_count), 1);
  const x = (idx) => {
    if (data.length === 1) return width / 2;
    return pad.left + (idx / (data.length - 1)) * (width - pad.left - pad.right);
  };
  const y = (value) => pad.top + (1 - value / max) * (height - pad.top - pad.bottom);
  const points = data.map((item, idx) => [x(idx), y(item.request_count)]);
  const path = points.map((point, idx) => `${idx === 0 ? "M" : "L"} ${point[0].toFixed(1)} ${point[1].toFixed(1)}`).join(" ");
  const area = `${path} L ${points[points.length - 1][0].toFixed(1)} ${height - pad.bottom} L ${points[0][0].toFixed(1)} ${height - pad.bottom} Z`;
  const ticks = [0, 0.25, 0.5, 0.75, 1]
    .map((ratio) => {
      const value = Math.round(max * ratio);
      const yy = y(value);
      return `
        <line class="line-grid" x1="${pad.left}" y1="${yy.toFixed(1)}" x2="${width - pad.right}" y2="${yy.toFixed(1)}"></line>
        <text class="line-axis-label" x="${pad.left - 10}" y="${(yy + 4).toFixed(1)}" text-anchor="end">${formatNumber(value)}</text>
      `;
    })
    .join("");
  const labels = data
    .map((item, idx) => {
      const xx = x(idx).toFixed(1);
      return `<text class="line-label" x="${xx}" y="${height - 12}" text-anchor="end" transform="rotate(-35 ${xx} ${height - 12})">${formatShortDate(item.key)}</text>`;
    })
    .join("");
  const dots = data
    .map((item, idx) => `
      <circle class="line-dot" cx="${x(idx).toFixed(1)}" cy="${y(item.request_count).toFixed(1)}" r="4">
        <title>${item.key}: ${formatNumber(item.request_count)} ${text[pageLang].requestSuffix}</title>
      </circle>
    `)
    .join("");
  const pointValues = data
    .map((item, idx) => {
      const xx = x(idx).toFixed(1);
      const yy = Math.max(12, y(item.request_count) - 10).toFixed(1);
      return `<text class="line-point-value" x="${xx}" y="${yy}" text-anchor="middle">${formatNumber(item.request_count)}</text>`;
    })
    .join("");

  chart.innerHTML = `
    <svg viewBox="0 0 ${width} ${height}" role="img" aria-label="${text[pageLang].requestChartLabel}">
      ${ticks}
      <path class="line-area" d="${area}"></path>
      <path class="line-path" d="${path}"></path>
      ${dots}
      ${pointValues}
      ${labels}
    </svg>
  `;
}

function uniqueSorted(values) {
  return [...new Set(values.filter(Boolean))].sort((a, b) => a.localeCompare(b));
}

function setFilterOptions(select, values, allLabel, selectedValue) {
  select.innerHTML = "";
  const all = document.createElement("option");
  all.value = "";
  all.textContent = allLabel;
  select.appendChild(all);
  for (const value of values) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = value;
    select.appendChild(option);
  }
  select.value = values.includes(selectedValue) ? selectedValue : "";
}

function recalculateModelTotal(rows, totalRow) {
  if (!totalRow) return;
  const totals = {
    inputTokens: 0,
    outputTokens: 0,
    cacheReadTokens: 0,
    cacheWriteTokens: 0,
    reasoningTokens: 0,
    toolTokens: 0,
    totalTokens: 0,
    estUsd: 0,
  };
  for (const row of rows) {
    totals.inputTokens += parseNumber(row.dataset.inputTokens);
    totals.outputTokens += parseNumber(row.dataset.outputTokens);
    totals.cacheReadTokens += parseNumber(row.dataset.cacheReadTokens);
    totals.cacheWriteTokens += parseNumber(row.dataset.cacheWriteTokens);
    totals.reasoningTokens += parseNumber(row.dataset.reasoningTokens);
    totals.toolTokens += parseNumber(row.dataset.toolTokens);
    totals.totalTokens += parseNumber(row.dataset.totalTokens);
    totals.estUsd += parseNumber(row.dataset.estUsd);
  }

  const cells = totalRow.children;
  if (cells.length < 11) return;
  const tokenTotals = [
    totals.inputTokens,
    totals.outputTokens,
    totals.cacheReadTokens,
    totals.cacheWriteTokens,
    totals.reasoningTokens,
    totals.toolTokens,
    totals.totalTokens,
  ];
  tokenTotals.forEach((value, index) => {
    const cell = cells[index + 3];
    cell.dataset.tokenValue = String(value);
    cell.title = formatNumber(value);
  });
  cells[10].textContent = `$${totals.estUsd.toFixed(3)}`;
  applyTokenDisplayMode(currentTokenDisplayMode());
}

function currentTokenDisplayMode() {
  return document.body.dataset.tokenDisplay === "raw" ? "raw" : "compact";
}

function readSavedTokenDisplayMode() {
  try {
    return localStorage.getItem(tokenDisplayStorageKey);
  } catch (_) {
    return "";
  }
}

function saveTokenDisplayMode(mode) {
  try {
    localStorage.setItem(tokenDisplayStorageKey, mode);
  } catch (_) {
    // Token display persistence is optional when storage is unavailable.
  }
}

function applyTokenDisplayMode(mode) {
  const nextMode = mode === "raw" ? "raw" : "compact";
  document.body.dataset.tokenDisplay = nextMode;
  for (const node of document.querySelectorAll("[data-token-value]")) {
    const value = parseNumber(node.dataset.tokenValue);
    node.textContent = formatTokenValue(value, nextMode);
    node.title = formatNumber(value);
  }
  const toggle = document.querySelector("#tokenDisplayToggle");
  if (toggle) {
    toggle.dataset.tokenMode = nextMode;
    toggle.setAttribute("aria-pressed", String(nextMode === "raw"));
    toggle.textContent = nextMode === "raw" ? text[pageLang].tokenRaw : text[pageLang].tokenCompact;
  }
}

function setupTokenDisplayToggle() {
  applyTokenDisplayMode(readSavedTokenDisplayMode());
  const toggle = document.querySelector("#tokenDisplayToggle");
  if (!toggle) return;
  toggle.addEventListener("click", () => {
    const nextMode = currentTokenDisplayMode() === "raw" ? "compact" : "raw";
    applyTokenDisplayMode(nextMode);
    saveTokenDisplayMode(nextMode);
  });
}

function applyModelFilters() {
  const toolFilter = document.querySelector("#modelToolFilter");
  const modelFilter = document.querySelector("#modelNameFilter");
  const rows = [...document.querySelectorAll(".model-data-row")];
  const totalRow = document.querySelector(".usage-table .total-row");
  if (!toolFilter || !modelFilter || !rows.length) return;

  const selectedTool = toolFilter.value;
  const modelValues = uniqueSorted(rows.filter((row) => !selectedTool || row.dataset.tool === selectedTool).map((row) => row.dataset.model));
  setFilterOptions(modelFilter, modelValues, modelFilter.dataset.allLabel || "All Models", modelFilter.value);

  const selectedModel = modelFilter.value;
  const visibleRows = [];
  for (const row of rows) {
    const show = (!selectedTool || row.dataset.tool === selectedTool) && (!selectedModel || row.dataset.model === selectedModel);
    row.hidden = !show;
    if (show) visibleRows.push(row);
  }
  recalculateModelTotal(visibleRows, totalRow);
}

function setupModelFilters() {
  const toolFilter = document.querySelector("#modelToolFilter");
  const modelFilter = document.querySelector("#modelNameFilter");
  const rows = [...document.querySelectorAll(".model-data-row")];
  if (!toolFilter || !modelFilter || !rows.length) return;

  setFilterOptions(toolFilter, uniqueSorted(rows.map((row) => row.dataset.tool)), toolFilter.dataset.allLabel || "All Tools", toolFilter.value);
  setFilterOptions(modelFilter, uniqueSorted(rows.map((row) => row.dataset.model)), modelFilter.dataset.allLabel || "All Models", modelFilter.value);
  toolFilter.addEventListener("change", applyModelFilters);
  modelFilter.addEventListener("change", applyModelFilters);
}

const grainSelect = document.querySelector("#grainSelect");
if (grainSelect) {
  grainSelect.addEventListener("change", () => {
    const url = new URL(window.location.href);
    url.searchParams.set("grain", grainSelect.value);
    window.location.href = url.toString();
  });
}

const langToggle = document.querySelector("#langToggle");
if (langToggle) {
  langToggle.addEventListener("click", () => {
    const url = new URL(window.location.href);
    url.searchParams.set("lang", langToggle.dataset.nextLang || "en");
    window.location.href = url.toString();
  });
}

renderTrend();
renderRequests();
renderRankings();
setupModelFilters();
setupTokenDisplayToggle();
setupThemeToggle();
