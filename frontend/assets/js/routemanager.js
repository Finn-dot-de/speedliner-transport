document.addEventListener("DOMContentLoaded", () => {
    fetchRoutes();

    document.getElementById("routeForm").addEventListener("submit", async (e) => {
        e.preventDefault();
        await saveRoute();
    });
});

async function fetchRoutes() {
    const res = await fetch("/app/routes");
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

    showRouteForm(true);
}

async function saveRoute() {
    const id = document.getElementById("routeId").value;
    const route = {
        from: document.getElementById("routeFrom").value,
        to: document.getElementById("routeTo").value,
        pricePerM3: parseFloat(document.getElementById("routePricePerM3").value),
        noCollateral: document.getElementById("routeNoCollateral").checked
    };

    const method = id ? "PUT" : "POST";
    const url = id ? `/app/routes/${id}` : "/app/routes";

    const res = await fetch(url, {
        method,
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

    const res = await fetch(`/app/routes/${id}`, { method: "DELETE" });
    if (!res.ok) {
        alert("Fehler beim Löschen!");
        return;
    }

    await fetchRoutes();
}

async function checkAccess() {
  try {
    const res = await fetch('/app/role', { credentials: 'include' });

    if (!res.ok) {
      // Nicht eingeloggt → zurück zur Startseite
      window.location.href = '/';
      return;
    }

    const data = await res.json();
    if (!(data.role === 'admin' || data.role === 'provider')) {
      // Keine Berechtigung → zurück zur Startseite
      window.location.href = '/';
    }
  } catch (err) {
    console.error('Rollenprüfung fehlgeschlagen:', err);
    window.location.href = '/';
  }
}

// Expose globally for onclick handlers
window.showRouteForm = showRouteForm;
window.editRoute = editRoute;
window.deleteRoute = deleteRoute;
checkAccess();