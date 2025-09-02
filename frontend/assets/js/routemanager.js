document.addEventListener("DOMContentLoaded", () => {
    initWhitelistUI();
    fetchRoutes();

    document.getElementById("routeForm").addEventListener("submit", async (e) => {
        e.preventDefault();
        await saveRoute();
    });
});

// ===== Routes =====
async function fetchRoutes() {
    const res = await fetch("/app/routes", { credentials: "include" });
    const data = await res.json();
    renderRoutes(data);
}

function renderRoutes(routes) {
    const tbody = document.querySelector("#routeTable tbody");
    tbody.innerHTML = "";

    routes.forEach(route => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
      <td>${route.from || "-"}</td>
      <td>${route.to || "-"}</td>
      <td>${route.pricePerM3 ?? 0} ISK</td>
      <td>
        ${route.visibility === 'whitelist'
            ? '<span class="badge" title="Nur ausgewählte Corps">Whitelist</span>'
            : '<span class="badge" title="Öffentlich">All</span>'}
        ${route.noCollateral ? '<span class="badge" title="Für diese Route ist keine Sicherheit nötig.">No collateral</span>' : ''}
        <button onclick="editRoute('${route.id}')" title="Bearbeiten">
          <i class="fa-solid fa-pen-to-square"></i>
        </button>
        <button onclick="deleteRoute('${route.id}')" title="Löschen">
          <i class="fa-solid fa-trash"></i>
        </button>
      </td>
    `;
        tr.dataset.route = JSON.stringify(route);
        tbody.appendChild(tr);
    });
}

function showRouteForm(editing = false) {
    document.getElementById("formTitle").textContent = editing ? "Route bearbeiten" : "Neue Route";
    document.getElementById("routeFormContainer").style.display = "block";

    if (!editing) {
        document.getElementById("routeForm").reset();
        document.getElementById("routeId").value = "";
        if (visibilitySelect) visibilitySelect.value = "all";
        if (whitelistBox) whitelistBox.style.display = "none";
        resetWhitelistUI();
    }
}

function editRoute(id) {
    const row = document.querySelector(`tr td button[onclick="editRoute('${id}')"]`).closest("tr");
    const route = JSON.parse(row.dataset.route);

    document.getElementById("routeId").value = route.id;
    document.getElementById("routeFrom").value = route.from;
    document.getElementById("routeTo").value = route.to;
    document.getElementById("routePricePerM3").value = route.pricePerM3;
    document.getElementById("routeNoCollateral").checked = !!route.noCollateral;

    // Visibility + Whitelist
    const vis = route.visibility || "all";
    visibilitySelect.value = vis;
    whitelistBox.style.display = vis === "whitelist" ? "block" : "none";
    resetWhitelistUI();

    // allowedCorps kann [id,...] oder [{corpId,name,ticker},...] sein
    if (Array.isArray(route.allowedCorps)) {
        route.allowedCorps.forEach(x => {
            if (typeof x === "number") {
                addSelectedCorp({ corpId: x, name: `Corp #${x}`, ticker: "" });
            } else if (x && typeof x === "object" && "corpId" in x) {
                addSelectedCorp(x);
            }
        });
        renderSelectedTags(Array.from(selectedCorps.values()));
    }

    showRouteForm(true);
}

async function saveRoute() {
    const id = document.getElementById("routeId").value;
    const route = {
        from: document.getElementById("routeFrom").value,
        to: document.getElementById("routeTo").value,
        pricePerM3: parseFloat(document.getElementById("routePricePerM3").value),
        noCollateral: document.getElementById("routeNoCollateral").checked,
        visibility: visibilitySelect.value,
        allowedCorps: visibilitySelect.value === "whitelist"
            ? Array.from(selectedCorps.keys()) // nur IDs ans Backend
            : []
    };

    const method = id ? "PUT" : "POST";
    const url = id ? `/app/routes/${id}` : "/app/routes";

    const res = await fetch(url, {
        method,
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(route),
    });

    if (!res.ok) {
        alert("Fehler beim Speichern der Route");
        return;
    }

    document.getElementById("routeForm").reset();
    document.getElementById("routeFormContainer").style.display = "none";
    await fetchRoutes();
}

async function deleteRoute(id) {
    if (!confirm("Wirklich löschen?")) return;

    const res = await fetch(`/app/routes/${id}`, { method: "DELETE", credentials: "include" });
    if (!res.ok) {
        alert("Fehler beim Löschen!");
        return;
    }
    await fetchRoutes();
}

// ===== Zugriff =====
async function checkAccess() {
    try {
        const res = await fetch('/app/role', { credentials: 'include' });
        if (!res.ok) { window.location.href = '/'; return; }
        const data = await res.json();
        if (!(data.role === 'admin' || data.role === 'provider')) {
            window.location.href = '/';
        }
    } catch (err) {
        console.error('Rollenprüfung fehlgeschlagen:', err);
        window.location.href = '/';
    }
}
checkAccess();

// ===== Whitelist-UI =====
let selectedCorps = new Map(); // corpId -> {corpId,ticker,name}
let visibilitySelect, whitelistBox, corpSearchInput, corpSearchResults, selectedCorpsBox;

function initWhitelistUI() {
    // Elemente NACH DOMContentLoaded holen
    visibilitySelect  = document.getElementById("routeVisibility");
    whitelistBox      = document.getElementById("whitelistBox");
    corpSearchInput   = document.getElementById("corpSearch");
    corpSearchResults = document.getElementById("corpSearchResults");
    selectedCorpsBox  = document.getElementById("selectedCorps");

    visibilitySelect?.addEventListener("change", () => {
        const v = visibilitySelect.value;
        whitelistBox.style.display = v === "whitelist" ? "block" : "none";
    });

    corpSearchInput?.addEventListener("input", debounce(() => {
        searchCorps(corpSearchInput.value.trim());
    }, 300));

    document.addEventListener("click", (e) => {
        if (corpSearchResults &&
            !corpSearchResults.contains(e.target) &&
            e.target !== corpSearchInput) {
            corpSearchResults.classList.remove("open");
        }
    });
}

function resetWhitelistUI() {
    selectedCorps.clear();
    if (selectedCorpsBox) selectedCorpsBox.innerHTML = "";
    if (corpSearchInput) corpSearchInput.value = "";
    if (corpSearchResults) {
        corpSearchResults.classList.remove("open");
        corpSearchResults.innerHTML = "";
    }
}

function addSelectedCorp(item) {
    if (!item || typeof item.corpId !== "number") return;
    if (!selectedCorps.has(item.corpId)) {
        selectedCorps.set(item.corpId, {
            corpId: item.corpId,
            name: item.name || `Corp #${item.corpId}`,
            ticker: item.ticker || ""
        });
    }
}

function removeSelectedCorp(id) {
    selectedCorps.delete(id);
    renderSelectedTags(Array.from(selectedCorps.values()));
}

// ---- Tags rendern (ersetzt deine aktuelle renderSelectedTags) ----
function renderSelectedTags(items) {
    selectedCorpsBox.innerHTML = "";
    items.forEach(({ corpId, ticker, name }) => {
        const el = document.createElement("div");
        el.className = "tag tag-corp";
        el.innerHTML = `
      <img class="avatar" src="${corpLogoUrl(corpId,32)}" alt="" loading="lazy"
           onerror="this.style.display='none'">
      <div class="meta">
        <span class="ticker">${ticker ? `[${ticker}]` : ""}</span>
        <span class="name">${name || corpId}</span>
      </div>
      <button type="button" class="tag-remove" title="Entfernen" aria-label="Entfernen">&times;</button>
    `;
        el.querySelector(".tag-remove").addEventListener("click", () => removeSelectedCorp(corpId));
        selectedCorpsBox.appendChild(el);
    });
}

// Suche
async function searchCorps(q) {
    if (!q || q.length < 2) {
        corpSearchResults?.classList.remove("open");
        if (corpSearchResults) corpSearchResults.innerHTML = "";
        return;
    }
    try {
        const res = await fetch(`/app/corps?q=${encodeURIComponent(q)}`, { credentials: "include" });
        const items = await res.json(); // [{corpId, ticker, name}]
        if (!corpSearchResults) return;
        corpSearchResults.innerHTML = items.map(it => `
      <div class="item" data-id="${it.corpId}" data-ticker="${it.ticker || ""}" data-name="${it.name || ""}">
        ${it.ticker ? `[${it.ticker}] ` : ""}${it.name}
      </div>`).join("");
        corpSearchResults.classList.add("open");

        corpSearchResults.querySelectorAll(".item").forEach(el => {
            el.addEventListener("click", () => {
                const corpId = parseInt(el.dataset.id, 10);
                addSelectedCorp({ corpId, ticker: el.dataset.ticker, name: el.dataset.name });
                renderSelectedTags(Array.from(selectedCorps.values()));
                if (corpSearchInput) corpSearchInput.value = "";
                corpSearchResults.classList.remove("open");
                corpSearchResults.innerHTML = "";
            });
        });
    } catch (e) {
        console.error("Corps search failed:", e);
    }
}

// helper
const corpLogoUrl = (id, size=32) =>
    `https://images.evetech.net/corporations/${id}/logo?size=${size}`;


// Utils
function debounce(fn, wait = 300) {
    let t;
    return (...args) => {
        clearTimeout(t);
        t = setTimeout(() => fn(...args), wait);
    };
}

// Expose globally for onclick handlers
window.showRouteForm = showRouteForm;
window.editRoute = editRoute;
window.deleteRoute = deleteRoute;
