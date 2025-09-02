const routeSelect = document.getElementById("route");
let routeData = {};
const collateralInput = document.getElementById("collateral");
const collateralRow = document.getElementById("collateralRow");
const routeMeta = document.getElementById("routeMeta");

const MAX_COLLATERAL = 20_000_000_000;
const MAX_VOLUME = 351_000;


routeSelect.addEventListener("change", () => {
    const defaultOption = routeSelect.querySelector("option[value='']");
    if (defaultOption) {
        defaultOption.remove();
    }
});

export function calculator() {
    const selectedRouteId = routeSelect.value;
    const route = routeData[selectedRouteId];
    const result = document.getElementById("result");
    const isk = new Intl.NumberFormat("de-DE");

    if (!route) {
        // Collateral wieder zeigen, falls keine Route gewählt
        if (collateralRow) collateralRow.style.display = "block";
        result.textContent = "Please select a route.";
        if (routeMeta) routeMeta.innerHTML = "";
        return;
    }

    const isCorpRoute = route.visibility === "whitelist";
    const hideCollateral = !!route.noCollateral;

    // Meta-Badges unter der Route anzeigen
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


    // Eingaben lesen
    const volumeRaw = document.getElementById("volume").value;
    const volume = parseInt((volumeRaw || "").replace(/\D/g, ""), 10);

    let collateral = 0;
    if (!hideCollateral) {
        const collateralRaw = collateralInput.value;
        collateral = parseInt((collateralRaw || "").replace(/\D/g, ""), 10);
    }

    if (isNaN(volume)) {
        result.textContent = "Please enter valid values for volume.";
        return;
    }
    if (!hideCollateral && isNaN(collateral)) {
        result.textContent = "Please enter a valid collateral.";
        return;
    }

    if (volume > MAX_VOLUME) {
        result.textContent = `Maximum volume exceeded (${isk.format(MAX_VOLUME)} m³).`;
        return;
    }
    if (!hideCollateral && collateral > MAX_COLLATERAL) {
        result.textContent = `Maximum of collateral can be only 20B ISK`;
        return;
    }

    const collateralPercent = hideCollateral
        ? 0
        : (volume <= (MAX_VOLUME / 2) ? 0.03 : 0.01);

    const volumeFee = volume * route.pricePerM3;
    const collateralFee = collateral * collateralPercent;
    const total = Math.round(volumeFee + collateralFee);

    result.textContent = `Reward: ${isk.format(total)} ISK`;
}

export function setRoutesData(routes) {
    routeData = {};
    routeSelect.innerHTML = `<option value="">Select route...</option>`;

    routes.forEach(route => {
        routeData[route.id] = route;

        const isCorpRoute = route.visibility === "whitelist";
        const flags = [
            isCorpRoute ? "🔒 Corp" : null,
            route.noCollateral ? "No collateral" : null
        ].filter(Boolean).join(" · ");

        const option = document.createElement("option");
        option.value = route.id;
        option.textContent = `${route.from} ↔ ${route.to}${flags ? "  —  " + flags : ""}`;
        option.title = flags || "";
        routeSelect.appendChild(option);
    });

    routeSelect.addEventListener("change", calculator);
}

