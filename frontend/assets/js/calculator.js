const routeSelect = document.getElementById("route");
let routeData = {};
const maxcollateral = 20_000_000_000
const maxvolume = 351_000
const collateralInput = document.getElementById("collateral");

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

    const iskFormatter = new Intl.NumberFormat("de-DE");

    if (!route) {
        result.textContent = "Invalid route selected.";
        return;
    }

    // Collateral-Feld je nach Route behandeln
    if (route.noCollateral) {
        collateralInput.value = "";
        collateralInput.disabled = true;
        collateralInput.title = "Für diese Route ist keine Sicherheit erforderlich.";
    } else {
        collateralInput.disabled = false;
        collateralInput.title = "";
    }

    // Werte lesen (wenn disabled/noCollateral => collateral = 0)
    const volumeRaw = document.getElementById("volume").value;
    const volume = parseInt((volumeRaw || "").replace(/\D/g, ""), 10);

    let collateral = 0;
    if (!route.noCollateral) {
        const collateralRaw = collateralInput.value;
        collateral = parseInt((collateralRaw || "").replace(/\D/g, ""), 10);
    }

    if (isNaN(volume)) {
        result.textContent = "Please enter valid values for volume.";
        return;
    }

    if (!route.noCollateral && isNaN(collateral)) {
        result.textContent = "Please enter a valid collateral.";
        return;
    }

    const maxcollateral = 20_000_000_000;
    const maxvolume = 351_000;

    if (volume > maxvolume) {
        result.textContent = `Maximum volume exceeded (${iskFormatter.format(maxvolume)} m³).`;
        return;
    }
    if (!route.noCollateral && collateral > maxcollateral) {
        result.textContent = `Maximum of collateral can be only 20B ISK`;
        return;
    }

    const collateralPercent = route.noCollateral
        ? 0
        : (volume <= (maxvolume / 2) ? 0.03 : 0.01);

    const volumeFee = volume * route.pricePerM3;
    const collateralFee = collateral * collateralPercent;
    const total = Math.round(volumeFee + collateralFee);

    result.textContent = `Reward: ${iskFormatter.format(total)} ISK`;
}

// nach dem Laden der Routes aufgerufen
export function setRoutesData(routes) {
    routeData = {};
    routeSelect.innerHTML = `<option value="">Select route...</option>`; // neu aufbauen

    routes.forEach(route => {
        routeData[route.id] = route;

        const option = document.createElement("option");
        option.value = route.id;
        option.textContent = `${route.from} ↔ ${route.to}`;
        if (route.noCollateral) {
            option.title = "Keine Sicherheit erforderlich";
        }
        routeSelect.appendChild(option);
    });

    // Wenn Route gewechselt wird, Felder/Tooltip aktualisieren
    routeSelect.addEventListener("change", () => {
        calculator(); // triggert die (de)aktivierung & Berechnung
    });
}
