import {copyContractName, setupAutoFormat} from "./utils.js";
import {calculator} from "./calculator.js";
import {loadUser} from "./user.js";
import {initMailModal} from "./mail.js";
import {CONTACT_NAME, contactNameEl, loadRoutes, ta} from "./costum_route_select.js";


setupAutoFormat("volume");
setupAutoFormat("collateral");

["route", "volume", "collateral"].forEach((id) => {
    document.getElementById(id).addEventListener("input", calculator);
});


if (contactNameEl) {
    contactNameEl.textContent = CONTACT_NAME;

    const copyBtn = document.getElementById("copyContactBtn");
    copyBtn?.addEventListener("click", async () => {
        try {
            await navigator.clipboard.writeText(CONTACT_NAME);
            const old = copyBtn.innerHTML;
            copyBtn.innerHTML = '<i class="fa-solid fa-check"></i> Copied';
            setTimeout(() => (copyBtn.innerHTML = old), 1400);
        } catch {
            alert("Could not copy the name to the clipboard.");
        }
    });
}

initMailModal({
    cooldownMin: 5,
    requiredRecipients: [
        {id: 92393462, type: "character"}

    ],
});


if (ta) {
    const grow = e => {
        e.target.style.height = 'auto';
        e.target.style.height = Math.min(e.target.scrollHeight, window.innerHeight * 0.6) + 'px';
    };
    ta.addEventListener('input', grow);
    grow({target: ta}); // initial
}


window.copyContractName = copyContractName;
loadRoutes();
loadUser();


