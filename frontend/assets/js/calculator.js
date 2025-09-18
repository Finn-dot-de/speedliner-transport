import {buildMrRouteSelect, mrRouteUI, refreshMrOptions} from "./costum_route_select.js";
import {copyByElementText} from "./utils.js";

const routeSelect = document.getElementById("route");
let routeData = {};
const collateralInput = document.getElementById("collateral");
const collateralRow = document.getElementById("collateralRow");
const routeMeta = document.getElementById("routeMeta");

const result = document.getElementById("result");
const volumeInput = document.getElementById("volume");

let userInteracted = false;

// letzte Berechnung für den Express-Description-Text
let lastQuote = null;
// wurde das Modal für diese Express-Aktivierung schon gezeigt?
let expressModalShown = false;

const MAX_COLLATERAL = 20_000_000_000;
const MAX_VOLUME = 351_000;
const expressInput = document.getElementById("express");
const daysToCompleteEl = document.getElementById("daysToComplete");

const DAYS_STANDARD = 3;
const DAYS_EXPRESS = 1;

function updateDays() {
    if (!daysToCompleteEl) return;
    daysToCompleteEl.textContent = expressInput?.checked ? String(DAYS_EXPRESS) : String(DAYS_STANDARD);
}

updateDays();

// Umschalten Express -> neue Session, Modal darf wieder erscheinen
expressInput?.addEventListener("input", () => {
    userInteracted = true;
    updateDays();
    expressModalShown = false;
    calculator();
});

function hideResult() {
    result.classList.remove("is-visible", "error");
    result.innerHTML = "";
}

function showResult(html, isError = false) {
    result.classList.toggle("error", !!isError);
    result.innerHTML = html;
    result.classList.add("is-visible");
}

routeSelect.addEventListener("change", () => {
    userInteracted = true;
    const defaultOption = routeSelect.querySelector('option[value=""]');
    if (defaultOption) defaultOption.remove();
    if (mrRouteUI) refreshMrOptions(routeSelect, mrRouteUI);
    calculator();
});

volumeInput.addEventListener("input", () => {
    userInteracted = true;
    calculator();
});
collateralInput.addEventListener("input", () => {
    userInteracted = true;
    calculator();
});

export function calculator() {
    const selectedRouteId = routeSelect.value;
    const route = routeData[selectedRouteId];
    const iskFmt = new Intl.NumberFormat("de-DE");

    if (!route) {
        if (!userInteracted) hideResult(); else showResult("Please select a route.");
        if (routeMeta) routeMeta.innerHTML = "";
        if (collateralRow) collateralRow.style.display = "block";
        lastQuote = null;
        updateDays();
        updateExpressUI();
        maybeOpenExpressModal(); // falls später gültig wird
        return;
    }

    const isCorpRoute = route.visibility === "whitelist";
    const hideCollateral = !!route.noCollateral;

    if (routeMeta) {
        routeMeta.innerHTML = `
      ${isCorpRoute ? `<span class="badge-corp">🔒 Corp route</span>` : ""}
      ${route.noCollateral ? `<span class="badge-nocoll">No collateral</span>` : ""}
    `;
    }

    if (hideCollateral) {
        collateralInput.value = "";
        collateralInput.disabled = true;
        collateralInput.title = "For this route no collateral is required.";
        if (collateralRow) collateralRow.style.display = "none";
    } else {
        collateralInput.disabled = false;
        collateralInput.title = isCorpRoute ? "Corp route: Collateral applies." : "";
        if (collateralRow) collateralRow.style.display = "block";
    }

    const volumeRaw = volumeInput.value;
    const volume = parseInt((volumeRaw || "").replace(/\D/g, ""), 10);

    let collateral = 0;
    if (!hideCollateral) {
        const collateralRaw = collateralInput.value;
        collateral = parseInt((collateralRaw || "").replace(/\D/g, ""), 10);
    }

    if (isNaN(volume)) {
        showResult("Please enter valid values for volume.", true);
        lastQuote = null;
        updateExpressUI();
        maybeOpenExpressModal();
        return;
    }
    if (!hideCollateral && isNaN(collateral)) {
        showResult("Please enter a valid collateral.", true);
        lastQuote = null;
        updateExpressUI();
        maybeOpenExpressModal();
        return;
    }
    if (volume > MAX_VOLUME) {
        showResult(`Maximum volume exceeded (${iskFmt.format(MAX_VOLUME)} m³).`, true);
        lastQuote = null;
        updateExpressUI();
        maybeOpenExpressModal();
        return;
    }
    if (!hideCollateral && collateral > MAX_COLLATERAL) {
        showResult(`Maximum of collateral can be only 20B ISK`, true);
        lastQuote = null;
        updateExpressUI();
        maybeOpenExpressModal();
        return;
    }

    const collateralPercent = hideCollateral ? 0 : (volume <= (MAX_VOLUME / 2) ? 0.03 : 0.01);
    const volumeFee = volume * route.pricePerM3;
    const collateralFee = collateral * collateralPercent;
    const baseTotal = Math.round(volumeFee + collateralFee);

    const expressOn = !!expressInput?.checked;
    const finalTotal = expressOn ? baseTotal * 2 : baseTotal;

    const resultHtml = expressOn
        ? `Reward (Express): <span class="value">${iskFmt.format(finalTotal)} ISK</span><br>
       <small>Basis: ${iskFmt.format(baseTotal)} ISK · +100% Express</small>`
        : `Reward: <span class="value">${iskFmt.format(finalTotal)} ISK</span>`;

    showResult(resultHtml);

    // letzte Quote merken
    const routeLabel = routeSelect.options[routeSelect.selectedIndex]?.text?.split(" — ")[0] ?? "-";
    lastQuote = {
        routeLabel,
        volume,
        collateral: hideCollateral ? 0 : (Number.isFinite(collateral) ? collateral : 0),
        baseTotal,
        finalTotal,
        expressOn,
        days: expressOn ? 1 : 3
    };

    updateExpressUI();
    maybeOpenExpressModal();
}

export function setRoutesData(routes) {
    routeData = {};
    routeSelect.innerHTML = `<option value="">Select route...</option>`;

    routes.forEach(route => {
        routeData[route.id] = route;
        const isCorpRoute = route.visibility === "whitelist";
        const flags = [isCorpRoute ? "🔒 Corp" : null, route.noCollateral ? "No collateral" : null]
            .filter(Boolean).join(" · ");

        const option = document.createElement("option");
        option.value = route.id;
        option.textContent = `${route.from} ↔ ${route.to}${flags ? " — " + flags : ""}`;
        option.title = flags || "";
        routeSelect.appendChild(option);
    });

    routeSelect.addEventListener("change", calculator);
    buildMrRouteSelect(routeSelect);
}

// ----- Express Inline UI (Hinweis + Copy) -----
const expressHint = document.getElementById("expressHint");
const copyExpressBtn = document.getElementById("copyExpressBtn");
const contractExpressDesc = document.getElementById("contractExpressDesc");

function setVisible(el, on) {
    if (el) el.style.display = on ? "" : "none";
}

function buildExpressDescription() {
    const iskFmt = new Intl.NumberFormat("de-DE");
    const routeLabel = routeSelect.options[routeSelect.selectedIndex]?.text?.split(" — ")[0] ?? "-";

    const lines = [
        `EXPRESS — PRIORITY COURIER`,
        `Route: ${routeLabel}`
    ];

    if (lastQuote && lastQuote.expressOn) {
        lines.push(`Reward: ${iskFmt.format(lastQuote.finalTotal)} ISK`);
        if (typeof lastQuote.collateral === "number" && lastQuote.collateral > 0) {
            lines.push(`Collateral: ${iskFmt.format(lastQuote.collateral)} ISK`);
        }
        if (typeof lastQuote.volume === "number" && !Number.isNaN(lastQuote.volume)) {
            lines.push(`Volume: ${iskFmt.format(lastQuote.volume)} m³`);
        }
        lines.push(`Days to complete: 1`);
    } else {
        lines.push(`Days to complete: 1`);
    }

    return lines.join("\n");
}

function updateExpressUI() {
    const route = routeData[routeSelect.value];
    const on = !!expressInput?.checked && !!route;
    setVisible(expressHint, on);
    setVisible(copyExpressBtn, on);
    if (!on) return;

    contractExpressDesc.value = buildExpressDescription();
}

copyExpressBtn?.addEventListener("click", async () => {
    // sicherstellen, dass der Text frisch ist
    updateExpressUI();

    // 1) immer kopieren
    await copyByElementText("contractExpressDesc", "copyExpressIcon");

    // 2) falls Modal NICHT offen ist: senden (wenn gültig)
    const modalOpen = expressModal?.classList.contains("open");
    const canSend = !!expressInput?.checked;

    // wenn noch keine Quote (z. B. gerade Collateral zuletzt getippt), erst rechnen
    if (!lastQuote) calculator();

    if (!modalOpen && canSend && lastQuote) {
        // nutzt deine bestehende Logik inkl. Cooldown
        await sendExpressOnce();
    }
});


routeSelect.addEventListener("change", updateExpressUI);
expressInput?.addEventListener("input", updateExpressUI);

// initial
updateExpressUI();

// ===== Express Modal =====
const expressModal = document.getElementById("expressModal");
const expressModalText = document.getElementById("expressModalText");
const expressConfirmBtn = document.getElementById("expressConfirmBtn");
const expressConfirmIcon = document.getElementById("expressConfirmIcon");
const expressCancelBtn = document.getElementById("expressCancelBtn");
const expressModalClose = document.getElementById("expressModalClose");

function openExpressModal() {
    if (!expressModal) return;
    expressModalText.textContent = buildExpressDescription();
    expressModal.classList.add("open");
    expressModal.setAttribute("aria-hidden", "false");
}

function closeExpressModal() {
    if (!expressModal) return;
    expressModal.classList.remove("open");
    expressModal.setAttribute("aria-hidden", "true");
}

let hadValidQuote = false; // neu

function maybeOpenExpressModal() {
    const nowValid = !!lastQuote;
    if (expressInput?.checked && nowValid && !hadValidQuote && !expressModalShown && !expressModal?.classList.contains("open")) {
        openExpressModal();
        expressModalShown = true;
    }
    hadValidQuote = nowValid; // State für nächsten Durchlauf merken
}

let t;
collateralInput.addEventListener("input", () => {
    userInteracted = true;
    clearTimeout(t);
    t = setTimeout(calculator, 120); // 120ms Debounce
});

collateralInput.addEventListener("change", () => {
    userInteracted = true;
    calculator();
});

expressConfirmBtn?.addEventListener("click", () => {
    copyByElementText("expressModalText", "expressConfirmIcon");
    expressModalShown = true;
    sendExpressOnce();
    closeExpressModal();
});

expressCancelBtn?.addEventListener("click", () => {
    cancelExpressAndClose();
});

expressInput?.addEventListener("change", () => {
    if (!expressInput.checked) return;
    const route = routeData[routeSelect.value];
    const vol = parseInt((volumeInput.value || "").replace(/\D/g, ""), 10);
    if (route && !Number.isNaN(vol)) {
        if (lastQuote) {
            openExpressModal();
            expressModalShown = true;
        }
    }
});

function cancelExpressAndClose() {
    if (expressInput) {
        expressInput.checked = false;
        updateDays();
        calculator();
    }
    expressModalShown = false;
    closeExpressModal();
}

// X -> wie Cancel
expressModalClose?.addEventListener("click", cancelExpressAndClose);

// Overlay-Klick -> wie Cancel
expressModal?.addEventListener("click", (e) => {
    if (e.target === expressModal) cancelExpressAndClose();
});

// ESC -> wie Cancel
document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && expressModal?.classList.contains("open")) {
        cancelExpressAndClose();
    }
});

// ===== Express Send (once per 5 min) =====
const EXPRESS_COOLDOWN_MIN = 2;
const EXPRESS_COOKIE = "express_mail_sent_at";
let expressIsSending = false;

function setCookie(name, value, minutes) {
    const d = new Date();
    d.setTime(d.getTime() + minutes * 60 * 1000);
    document.cookie = `${name}=${encodeURIComponent(value)};expires=${d.toUTCString()};path=/;SameSite=Lax`;
}

function getCookie(name) {
    const row = document.cookie
        .split("; ")
        .find((r) => r.startsWith(name + "="));
    if (!row) return null;
    const val = row.split("=").slice(1).join("=");
    try {
        return decodeURIComponent(val);
    } catch {
        return val;
    }
}

function getCooldownMsLeft() {
    const ts = parseInt(getCookie(EXPRESS_COOKIE) || "0", 10);
    if (!ts) return 0;
    const left = ts + EXPRESS_COOLDOWN_MIN * 60 * 1000 - Date.now();
    return left > 0 ? left : 0;
}

function fmtMMSS(ms) {
    const s = Math.ceil(ms / 1000);
    const m = Math.floor(s / 60);
    const r = s % 60;
    return `${m}:${String(r).padStart(2, "0")}`;
}


async function fetchMe() {
    try {
        const res = await fetch("/app/me", { credentials: "include" });

        if (!res.ok) {
            return { CharacterID: 808, CharacterName: "Unknown" };
        }

        const data = await res.json();
        return {
            CharacterID:  data?.CharacterID  ?? "Unknown",
            CharacterName: data?.CharacterName ?? "Unknown",
        };
    } catch {
        return { CharacterID: "Unknown", CharacterName: "Unknown" };
    }
}

function routeLabelToArrow(label) {
    return (label || "-").split(" — ")[0].replace("↔", "→").trim();
}

function buildExpressPayload(meInfo) {
    const routeLabel = routeSelect.options[routeSelect.selectedIndex]?.text ?? "-";
    const routeStr = routeLabelToArrow(routeLabel);

    const reward = Number(lastQuote?.finalTotal ?? 0);

    const vol = Number(
        lastQuote?.volume ??
        (parseInt(String(volumeInput.value || "").replace(/\D/g, ""), 10) || 0)
    );

    const coll = Number(
        lastQuote?.collateral ??
        (parseInt(String(collateralInput.value || "").replace(/\D/g, ""), 10) || 0)
    );

    return {
        express: true,
        route: routeStr,
        rewardISK: reward,
        reward_isk: reward,
        volumeM3: vol,
        volume_m3: vol,
        collateralISK: coll,
        collateral_isk: coll,
        notes: "EXPRESS Contract kommt demnächst vom Piloten:",
        customer_char_id: 10000,
        customer_char_name: meInfo?.CharacterName,
    };
}

export async function sendExpressOnce() {
    const btn = document.getElementById("expressConfirmBtn");
    const icon = document.getElementById("expressConfirmIcon");

    // Cooldown aktiv?
    const left = getCooldownMsLeft();
    if (left > 0) {
        // Nichts senden. Kurze visuelle Rückmeldung am Button, aber keine Aktion ansonsten.
        if (btn) {
            const oldHtml = btn.innerHTML;
            btn.disabled = true;
            btn.innerHTML = `Cooldown ${fmtMMSS(left)}`;
            setTimeout(() => {
                btn.disabled = false;
                btn.innerHTML = oldHtml;
            }, 1500);
        }
        return; // NICHT kopieren, NICHT schließen – wie gewünscht "nichts passiert"
    }

    if (expressIsSending) return;
    expressIsSending = true;
    if (btn) btn.disabled = true;
    if (icon) {
        icon.classList.remove("fa-copy");
        icon.classList.add("fa-spinner", "fa-spin");
    }

    try {
        // Nutzerinfo holen (optional)
        let me = await fetchMe();

        // Payload bauen
        const payload = buildExpressPayload(me);

        // Senden
        const res = await fetch("/app/express/mail", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            credentials: "include",
            body: JSON.stringify(payload),
        });


        if (res.status === 201) {
            setCookie(EXPRESS_COOKIE, String(Date.now()), EXPRESS_COOLDOWN_MIN);

            await copyByElementText("expressModalText", "expressConfirmIcon");
            expressModalShown = true;
            closeExpressModal();

            // Done-Icon
            if (icon) {
                icon.classList.remove("fa-spinner", "fa-spin");
                icon.classList.add("fa-check");
            }
        } else {
            const txt = await res.text().catch(() => "");
            console.error("Express send failed:", res.status, txt);
            if (btn) {
                const oldHtml = btn.innerHTML;
                btn.innerHTML = `Error ${res.status}`;
                setTimeout(() => (btn.innerHTML = oldHtml), 1500);
            }
            if (icon) {
                icon.classList.remove("fa-spinner", "fa-spin");
                icon.classList.add("fa-copy");
            }
        }
    } catch (e) {
        console.error("Express send error:", e);
        if (icon) {
            icon.classList.remove("fa-spinner", "fa-spin");
            icon.classList.add("fa-copy");
        }
    } finally {
        expressIsSending = false;
        if (btn) btn.disabled = false;
    }
}
