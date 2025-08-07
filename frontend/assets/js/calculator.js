const routeSelect = document.getElementById("route");
const pricingBox = document.querySelector(".calculator .result");
let routeData = {};

export function calculator() {
    const selectedRouteId = routeSelect.value;
    const route = routeData[selectedRouteId];
    const result = document.getElementById("result");

    const collateralRaw = document.getElementById("collateral").value;
    const collateral = parseInt(collateralRaw.replace(/\D/g, ""), 10);

    const volumeRaw = document.getElementById("volume").value;
    const volume = parseInt(volumeRaw.replace(/\D/g, ""), 10);

    const iskFormatter = new Intl.NumberFormat("de-DE");

    if (isNaN(volume) || isNaN(collateral)) {
        result.textContent = "Please enter valid values for volume and collateral.";
        return;
    }

    if (!route) {
        result.textContent = "Invalid route selected.";
        return;
    }

    if (volume > route.volumeMax) {
        result.textContent = `Maximum volume exceeded (${iskFormatter.format(route.volumeMax)} m³).`;
        return;
    }

    const collateralPercent = volume <= 165000 ? route.collateralFeePercent : 0.01;
    const volumeFee = volume * route.pricePerM3;
    const collateralFee = collateral * collateralPercent;
    const total = Math.round(volumeFee + collateralFee);

    result.textContent = `Reward: ${iskFormatter.format(total)} ISK`;
}

// Muss nach dem Laden der Routes aufgerufen werden!
export function setRoutesData(routes) {
    // Route-Daten in ein Lookup-Objekt umwandeln
    routeData = {};
    routes.forEach(route => {
        routeData[route.id] = route;

        // Auswahloptionen erzeugen
        const option = document.createElement("option");
        option.value = route.id;
        option.textContent = `${route.from} ↔ ${route.to}`;
        routeSelect.appendChild(option);
    });
}
