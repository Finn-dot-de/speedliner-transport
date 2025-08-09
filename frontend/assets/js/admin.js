// Top-level await funktioniert nur mit type="module"
const redirectHome = () => {
  location.replace("/");
  // prevent further code running after redirect
  throw new Error("redirect");
};

try {
  const res = await fetch("/app/role", { credentials: "include" });
  if (!res.ok) redirectHome();

  const { role } = await res.json();
  if (role !== "admin") redirectHome();
} catch {
  redirectHome();
}

// User-Liste laden
async function loadUsers() {
  const res = await fetch("/app/users", { credentials: "include" });
  if (!res.ok) return; // optional: Fehlerbehandlung
  const users = await res.json();

  const tbody = document.querySelector("#userTable tbody");
  tbody.innerHTML = "";

  users.forEach((user) => {
    const tr = document.createElement("tr");

    // Portrait
    const portraitTd = document.createElement("td");
    portraitTd.innerHTML = `
      <img src="https://images.evetech.net/characters/${user.char_id}/portrait?size=64"
           alt="${user.name}" width="48" height="48"
           style="border-radius:8px;object-fit:cover;">
    `;
    tr.appendChild(portraitTd);

    // Name
    const nameTd = document.createElement("td");
    nameTd.textContent = user.name;
    tr.appendChild(nameTd);

    // Rolle Auswahl
    const roleTd = document.createElement("td");
    const select = document.createElement("select");
    ["user", "provider", "admin"].forEach((r) => {
      const opt = document.createElement("option");
      opt.value = r;
      opt.textContent = r;
      if (user.role === r) opt.selected = true;
      select.appendChild(opt);
    });
    roleTd.appendChild(select);
    tr.appendChild(roleTd);

    // Speichern Button
    const actionTd = document.createElement("td");
    const btn = document.createElement("button");
    btn.textContent = "Speichern";
    btn.onclick = async () => {
      const newRole = select.value;
      const resp = await fetch(`/app/users/${user.char_id}/role`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ role: newRole }),
      });
      if (!resp.ok) {
        alert("Fehler beim Speichern");
        return;
      }
      alert(`Rolle geändert zu: ${newRole}`);
    };
    actionTd.appendChild(btn);
    tr.appendChild(actionTd);

    tbody.appendChild(tr);
  });
}

// Seite füllen
await loadUsers();
