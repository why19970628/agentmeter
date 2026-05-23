function formatNumber(value) {
  return new Intl.NumberFormat("zh-CN").format(value || 0);
}

function formatCompact(value) {
  return new Intl.NumberFormat("zh-CN", { notation: "compact", maximumFractionDigits: 1 }).format(value || 0);
}

function formatShortDate(value) {
  const parts = String(value || "").split("-");
  if (parts.length >= 3) {
    return `${Number(parts[1])}-${Number(parts[2])}`;
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
    chart.innerHTML = '<p class="empty">未读取到本地 token 记录</p>';
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
    label.innerHTML = `<span>${formatShortDate(item.key)}</span><strong>${formatCompact(item.total_tokens)}</strong>`;
    wrap.appendChild(bar);
    wrap.appendChild(label);
    chart.appendChild(wrap);
  }
}

function renderRankings() {
  for (const root of document.querySelectorAll(".ranking")) {
    const data = JSON.parse(root.dataset.ranking || "[]").slice(0, 10);
    const max = Math.max(...data.map((item) => item.total_tokens), 1);
    root.innerHTML = "";
    if (!data.length) {
      root.innerHTML = '<p class="empty">暂无数据</p>';
      continue;
    }
    for (const item of data) {
      const row = document.createElement("div");
      row.className = "rank-row";
      row.innerHTML = `
        <span class="rank-name" title="${item.name}">${item.name}</span>
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
    chart.innerHTML = '<p class="empty">暂无每日请求数据</p>';
    return;
  }

  const width = 1000;
  const height = 260;
  const pad = { top: 16, right: 22, bottom: 46, left: 58 };
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
        <title>${item.key}: ${formatNumber(item.request_count)} 次请求</title>
      </circle>
    `)
    .join("");

  chart.innerHTML = `
    <svg viewBox="0 0 ${width} ${height}" role="img" aria-label="每日 API 请求次数折线图">
      ${ticks}
      <path class="line-area" d="${area}"></path>
      <path class="line-path" d="${path}"></path>
      ${dots}
      ${labels}
    </svg>
  `;
}

const grainSelect = document.querySelector("#grainSelect");
if (grainSelect) {
  grainSelect.addEventListener("change", () => {
    const url = new URL(window.location.href);
    url.searchParams.set("grain", grainSelect.value);
    window.location.href = url.toString();
  });
}

renderTrend();
renderRequests();
renderRankings();
