export async function loadUser() {
    try {
        const res = await fetch("/app/me");
        if (!res.ok) throw new Error("Not logged in");

        const data = await res.json();
        const loginContainer = document.getElementById("loginContainer");
        if (!loginContainer) return;

        const imgUrl = `https://images.evetech.net/characters/${data.CharacterID}/portrait?size=128`;

        loginContainer.innerHTML = `
      <div id="userMenu" class="user-menu">
        <div class="user-menu-toggle">
          <img src="${imgUrl}" alt="Avatar" class="user-avatar" />
          <span class="user-name">${data.CharacterName}</span>
          <i id="arrow-down-avatar" class="fa-solid fa-chevron-down"></i>
        </div>
        <div id="logoutDropdown" class="user-dropdown">
          <a href="#" id="adminPanelBtn" style="display: none;">
            <i class="fa-solid fa-shield-halved"></i> Admin Panel
          </a>
          <a href="#" id="providerPanelBtn" style="display: none;">
            <i class="fa-solid fa-truck"></i> Provider
          </a>
          <a href="#" id="logoutLink">
            <i class="fa-solid fa-right-from-bracket"></i> Logout
          </a>

        </div>
      </div>
    `;

        const roleRes = await fetch("/app/role");
        if (roleRes.ok) {
            const { role } = await roleRes.json();
            if (["admin"].includes(role)) document.getElementById("adminPanelBtn").style.display = "block";
            if (["admin", "provider"].includes(role)) document.getElementById("providerPanelBtn").style.display = "block";
        }

        const userMenu = document.getElementById("userMenu");
        const dropdown = document.getElementById("logoutDropdown");

        userMenu.querySelector(".user-menu-toggle").addEventListener("click", (e) => {
            e.stopPropagation(); // verhindert window.close
            dropdown.classList.toggle("open");
        });

        window.addEventListener("click", () => dropdown.classList.remove("open"));

        document.getElementById("logoutLink").addEventListener("click", async (e) => {
            e.preventDefault();
            await logoutUser();
        });

        document.getElementById("adminPanelBtn").addEventListener("click", () => {
            window.location.href = "/users.html";
        });

        document.getElementById("providerPanelBtn").addEventListener("click", () => {
            window.location.href = "/routes.html";
        });

    } catch (err) {
        console.log("User not logged in:", err.message);
    }
}

export async function logoutUser() {
    try {
        await fetch("/app/logout");
        window.location.reload();
    } catch (err) {
        console.error("Logout failed", err);
    }
}
