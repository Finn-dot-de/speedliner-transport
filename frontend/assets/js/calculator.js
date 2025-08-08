const routeSelect = document.getElementById("route");
let routeData = {};
const maxcollateral = 10_000_000_000
const maxvolume = 337000

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

    if (volume > maxvolume) {
        result.textContent = `Maximum volume exceeded (${iskFormatter.format(maxvolume)} m³).`;
        return;
    }

    if (collateral > maxcollateral) {
        result.textContent = `Maximum of collateral can be only 10B ISK`;
        return;
    }

    const collateralPercent = volume <= 165000 ? 0.03 : 0.01;
    console.log(`${collateralPercent}%`);
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
