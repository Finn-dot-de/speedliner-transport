
import {buildMrRouteSelect, mrRouteUI, refreshMrOptions} from "./costum_route_select.js";

const routeSelect = document.getElementById("route");
let routeData = {};
const collateralInput = document.getElementById("collateral");
const collateralRow = document.getElementById("collateralRow");
const routeMeta = document.getElementById("routeMeta");

const result = document.getElementById("result");
const volumeInput = document.getElementById("volume");

let userInteracted = false;

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

expressInput?.addEventListener("input", () => {
    userInteracted = true;
    updateDays();
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
    calculator(); // direkt neu rechnen
});

volumeInput.addEventListener("input", () => {
    userInteracted = true;
    calculator();
});
collateralInput.addEventListener("input", () => {
    userInteracted = true;
    calculator();
});
expressInput?.addEventListener("input", () => {
    userInteracted = true;
    calculator();
});

export function calculator() {
    const selectedRouteId = routeSelect.value;
    const route = routeData[selectedRouteId];
    const isk = new Intl.NumberFormat("de-DE");

    if (!route) {
        if (!userInteracted) hideResult(); else showResult("Please select a route.");
        if (routeMeta) routeMeta.innerHTML = "";
        if (collateralRow) collateralRow.style.display = "block";
        updateDays();
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
        return;
    }
    if (!hideCollateral && isNaN(collateral)) {
        showResult("Please enter a valid collateral.", true);
        return;
    }
    if (volume > MAX_VOLUME) {
        showResult(`Maximum volume exceeded (${isk.format(MAX_VOLUME)} m³).`, true);
        return;
    }
    if (!hideCollateral && collateral > MAX_COLLATERAL) {
        showResult(`Maximum of collateral can be only 20B ISK`, true);
        return;
    }

    const collateralPercent = hideCollateral ? 0 : (volume <= (MAX_VOLUME / 2) ? 0.03 : 0.01);
    const volumeFee = volume * route.pricePerM3;
    const collateralFee = collateral * collateralPercent;
    const baseTotal = Math.round(volumeFee + collateralFee);

    const expressOn = !!expressInput?.checked;
    const finalTotal = expressOn ? baseTotal * 2 : baseTotal;

    const resultHtml = expressOn
        ? `Reward (Express): <span class="value">${isk.format(finalTotal)} ISK</span><br>
       <small>Basis: ${isk.format(baseTotal)} ISK · +100% Express</small>`
        : `Reward: <span class="value">${isk.format(finalTotal)} ISK</span>`;

    showResult(resultHtml);
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
