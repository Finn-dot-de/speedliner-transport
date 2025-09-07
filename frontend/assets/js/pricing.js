export function renderPricing(routes) {
    const pricingBox = document.getElementById("pricingBox");
    pricingBox.innerHTML = `<h3>Pricing:</h3>`;

    if (!Array.isArray(routes)) {
        console.warn("Expected array for routes, got:", routes);
        pricingBox.innerHTML = "<p style='color: red;'>Could not load pricing info.</p>";
        return;
    }

    const fmt = (n) => Number(n).toLocaleString();

    routes.forEach(route => {
        if (!route || typeof route.from !== "string" || typeof route.to !== "string" || typeof route.pricePerM3 !== "number") {
            console.warn("Invalid route object:", route);
            return;
        }

        const name = `${route.from} ↔ ${route.to}`;
        const isCorp = route.visibility === "whitelist";
        const badges = [
            isCorp ? `<span class="badge-corp">🔒 Corp</span>` : "",
            route.noCollateral ? `<span class="badge-nocoll">No collateral</span>` : ""
        ].join(" ");

        // Text abhängig davon, ob Collateral greift
        const body = route.noCollateral
            ? `Price: ${fmt(route.pricePerM3)} ISK/m³`
            : `Up to 175.500m³: ${fmt(route.pricePerM3)} ISK/m³ + 3% collateral fee<br/>
         175.500m³ and more: ${fmt(route.pricePerM3)} ISK/m³ + 1% collateral fee`;

        pricingBox.insertAdjacentHTML("beforeend", `
      <div class="pricing-item" style="margin-bottom:1rem;">
        <strong>Route: ${name}</strong> ${badges}<br/>
        ${body}
      </div>
    `);
    });
}
