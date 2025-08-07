export async function loadUser() {
    try {
        const res = await fetch("/app/me");
        if (!res.ok) throw new Error("Not logged in");

        const data = await res.json();
        const loginContainer = document.getElementById("loginContainer");
        if (!loginContainer) return;

        const imgUrl = `https://images.evetech.net/characters/${data.CharacterID}/portrait?size=64`;

        loginContainer.innerHTML = `
      <div id="userMenu" style="position: relative; display: inline-block; cursor: pointer;">
        <div style="display: flex; align-items: center; gap: 0.5rem; background-color: #1f2833; padding: 0.5rem 1rem; border-radius: 6px;">
          <img src="${imgUrl}" alt="Avatar" style="width: 32px; height: 32px; border-radius: 50%;" />
          <span style="color: #66fcf1; font-weight: bold;">${data.CharacterName}</span>
        </div>
        <div id="logoutDropdown" style="display: none; position: absolute; right: 0; top: 100%; background-color: #0b0c10; border: 1px solid #45a29e; padding: 0.5rem; border-radius: 6px; margin-top: 0.5rem;">
          <a href="#" id="logoutLink" style="color: #c5c6c7; text-decoration: none;">Logout</a>
        </div>
      </div>
    `;

        const userMenu = document.getElementById("userMenu");
        const dropdown = document.getElementById("logoutDropdown");
        userMenu.addEventListener("click", () => {
            dropdown.style.display = dropdown.style.display === "block" ? "none" : "block";
        });

        document.getElementById("logoutLink").addEventListener("click", async (e) => {
            e.preventDefault();
            await logoutUser();
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
