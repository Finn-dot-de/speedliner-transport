import {setupAutoFormat, copyContractName} from "./utils.js";
import {calculator, setRoutesData} from "./calculator.js";
import {renderPricing} from "./pricing.js";
import {loadUser} from "./user.js";


export async function loadRoutes() {
    try {
        const res = await fetch("/app/routes");
        const routes = await res.json();

        setRoutesData(routes);
        renderPricing(routes);

        document.getElementById("route").addEventListener("change", calculator);
        document.getElementById("volume").addEventListener("input", calculator);
        document.getElementById("collateral").addEventListener("input", calculator);

    } catch (err) {
        console.error("Failed to load routes:", err);
    }
}

setupAutoFormat("volume");
setupAutoFormat("collateral");

["route", "volume", "collateral"].forEach((id) => {
    document.getElementById(id).addEventListener("input", calculator);
});


window.copyContractName = copyContractName;
loadRoutes();
loadUser();
