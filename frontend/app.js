const apiBaseUrl = "http://localhost:8080/api";

const alertForm = document.querySelector("#alert-form");
const statusText = document.querySelector("#status");
const refreshButton = document.querySelector("#refresh");
const metricsGrid = document.querySelector("#metrics");
const agentTable = document.querySelector("#agent-table");

const fields = {
  enabled: document.querySelector("#alerts-enabled"),
  webhook: document.querySelector("#discord-webhook"),
  cpu: document.querySelector("#cpu-threshold"),
  ram: document.querySelector("#ram-threshold"),
  disk: document.querySelector("#disk-threshold"),
  cooldown: document.querySelector("#cooldown-seconds"),
};

refreshButton.addEventListener("click", loadData);
alertForm.addEventListener("submit", saveAlerts);

async function saveAlerts(event) {
  event.preventDefault();

  const payload = {
    enabled: fields.enabled.checked,
    discord_webhook_url: fields.webhook.value.trim(),
    cpu_percent: numberValue(fields.cpu),
    ram_percent: numberValue(fields.ram),
    disk_percent: numberValue(fields.disk),
    cooldown_seconds: numberValue(fields.cooldown),
  };

  try {
    const response = await fetch(`${apiBaseUrl}/alerts`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(`Backend Status ${response.status}`);
    }

    statusText.textContent = "Alert-Einstellungen gespeichert.";
  } catch (error) {
    statusText.textContent = `Fehler: ${error.message}`;
  }
}

async function loadAlerts() {
  const response = await fetch(`${apiBaseUrl}/alerts`);
  if (!response.ok) {
    throw new Error(`Backend Status ${response.status}`);
  }

  const config = await response.json();
  fields.enabled.checked = Boolean(config.enabled);
  fields.webhook.value = config.discord_webhook_url || "";
  fields.cpu.value = config.cpu_percent || 90;
  fields.ram.value = config.ram_percent || 90;
  fields.disk.value = config.disk_percent || 90;
  fields.cooldown.value = config.cooldown_seconds || 300;
}

async function loadData() {
  try {
    const response = await fetch(`${apiBaseUrl}/data`);
    if (!response.ok) {
      throw new Error(`Backend Status ${response.status}`);
    }

    const data = await response.json();
    renderAgents(latestByAgent(data));
  } catch (error) {
    statusText.textContent = `Backend nicht erreichbar: ${error.message}`;
  }
}

function latestByAgent(items) {
  const latest = new Map();

  for (const item of items) {
    const current = latest.get(item.agent_id);
    if (!current || new Date(item.timestamp) > new Date(current.timestamp)) {
      latest.set(item.agent_id, item);
    }
  }

  return [...latest.values()].sort((a, b) => a.agent_id.localeCompare(b.agent_id));
}

function renderAgents(agents) {
  if (agents.length === 0) {
    metricsGrid.innerHTML = `<article class="empty">Noch keine Agent-Daten empfangen.</article>`;
    agentTable.innerHTML = "";
    return;
  }

  metricsGrid.innerHTML = agents
    .map((agent) => {
      const meta = agent.meta || {};
      return `
        <article class="metric-card">
          <div>
            <p class="agent-name">${escapeHTML(agent.agent_id)}</p>
            <p class="muted">${escapeHTML(meta.hostname || "unknown")}</p>
          </div>
          <div class="metric-row">
            ${metricPill("CPU", meta.cpu_percent)}
            ${metricPill("RAM", meta.ram_percent)}
            ${metricPill("Disk", meta.disk_percent)}
          </div>
        </article>
      `;
    })
    .join("");

  agentTable.innerHTML = agents
    .map((agent) => {
      const meta = agent.meta || {};
      return `
        <tr>
          <td>${escapeHTML(agent.agent_id)}</td>
          <td>${formatPercent(meta.cpu_percent)}</td>
          <td>${formatPercent(meta.ram_percent)}</td>
          <td>${formatPercent(meta.disk_percent)}</td>
          <td>${formatTime(agent.timestamp)}</td>
        </tr>
      `;
    })
    .join("");
}

function metricPill(label, value) {
  const numeric = Number(value || 0);
  const level = numeric >= 90 ? "danger" : numeric >= 75 ? "warn" : "ok";
  return `<span class="pill ${level}">${label} ${formatPercent(numeric)}</span>`;
}

function numberValue(input) {
  return Number(input.value || 0);
}

function formatPercent(value) {
  const numeric = Number(value);
  return Number.isFinite(numeric) ? `${numeric.toFixed(1)}%` : "-";
}

function formatTime(value) {
  if (!value) {
    return "-";
  }
  return new Intl.DateTimeFormat("de-DE", {
    dateStyle: "short",
    timeStyle: "medium",
  }).format(new Date(value));
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

Promise.all([loadAlerts(), loadData()]).catch((error) => {
  statusText.textContent = `Startfehler: ${error.message}`;
});

setInterval(loadData, 10000);
