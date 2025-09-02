
const redirectHome = () => {
  location.replace("/");

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


async function loadUsers() {
  const res = await fetch("/app/users", { credentials: "include" });
  if (!res.ok) return;
  const users = await res.json();

  const tbody = document.querySelector("#userTable tbody");
  tbody.innerHTML = "";

  users.forEach((user) => {
    const tr = document.createElement("tr");

    const portraitTd = document.createElement("td");
    portraitTd.innerHTML = `
      <img src="https://images.evetech.net/characters/${user.char_id}/portrait?size=64"
           alt="${user.name}" width="48" height="48"
           style="border-radius:8px;object-fit:cover;">
    `;
    tr.appendChild(portraitTd);

    const nameTd = document.createElement("td");
    nameTd.textContent = user.name;
    tr.appendChild(nameTd);

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

    const actionTd = document.createElement("td");
    const btn = document.createElement("button");
    btn.textContent = "Save";
    btn.onclick = async () => {
      const newRole = select.value;
      const resp = await fetch(`/app/users/${user.char_id}/role`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ role: newRole }),
      });
      if (!resp.ok) {
        alert("Error to save");
        return;
      }
      alert(`Role change to: ${newRole}`);
    };
    actionTd.appendChild(btn);
    tr.appendChild(actionTd);

    tbody.appendChild(tr);
  });
}

await loadUsers();
