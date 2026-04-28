const backendUrl = "http://localhost:8080/api/data";

const form = document.querySelector("#data-form");
const statusText = document.querySelector("#status");
const output = document.querySelector("#output");
const refreshButton = document.querySelector("#refresh");

form.addEventListener("submit", async (event) => {
  event.preventDefault();

  const payload = {
    agent_id: document.querySelector("#agent-id").value,
    message: document.querySelector("#message").value,
    meta: {
      source: "frontend",
    },
  };

  try {
    const response = await fetch(backendUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(`Backend Status ${response.status}`);
    }

    statusText.textContent = "Gesendet.";
    await loadData();
  } catch (error) {
    statusText.textContent = `Fehler: ${error.message}`;
  }
});

refreshButton.addEventListener("click", loadData);

async function loadData() {
  const response = await fetch(backendUrl);
  const data = await response.json();
  output.textContent = JSON.stringify(data, null, 2);
}

loadData().catch((error) => {
  statusText.textContent = `Backend nicht erreichbar: ${error.message}`;
});
