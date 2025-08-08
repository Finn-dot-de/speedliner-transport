export function renderPricing(routes) {
    const pricingBox = document.getElementById("pricingBox");

    pricingBox.innerHTML = `<h3>Pricing:</h3>`;


    // Defensive Checks
    if (!Array.isArray(routes)) {
        console.warn("Expected array for routes, got:", routes);
        pricingBox.innerHTML = "<p style='color: red;'>Could not load pricing info.</p>";
        return;
    }

    routes.forEach(route => {
        if (
            !route ||
            typeof route.from !== "string" ||
            typeof route.to !== "string" ||
            typeof route.pricePerM3 !== "number"
        ) {
            console.warn("Invalid route object:", route);
            return;
        }

        const routeName = `${route.from} ↔ ${route.to}`;

        const routeHtml = `
            <div style="margin-bottom: 1rem;">
                <strong>Route: ${routeName}</strong><br/>
                Up to 165.000m³: ${route.pricePerM3.toLocaleString()} ISK/m³ + 3% collateral fee<br/>
                165.000m³ and more: ${route.pricePerM3.toLocaleString()} ISK/m³ + 1% collateral fee
            </div>
        `;

        pricingBox.innerHTML += routeHtml;
    });
}
