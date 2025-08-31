export async function fetchRoutes() {
    const res = await fetch("/app/routes");
    if (!res.ok) throw new Error("Failed to fetch routes");
    return await res.json();
}
