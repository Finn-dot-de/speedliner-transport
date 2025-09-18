export function copyContractName() {
    const name = document.getElementById("contractName").textContent;
    const copyIcon = document.getElementById("copyIcon");

    navigator.clipboard
        .writeText(name)
        .then(() => {
            copyIcon.classList.replace("fa-copy", "fa-check");
            copyIcon.title = "Copied!";
            copyIcon.style.color = "green";

            setTimeout(() => {
                copyIcon.classList.replace("fa-check", "fa-copy");
                copyIcon.title = "Name copied";
                copyIcon.style.color = "";
            }, 2000);
        })
        .catch((err) => {
            console.error("Error with copying:", err);
            copyIcon.title = "Error with copying";
            copyIcon.style.color = "red";

            setTimeout(() => {
                copyIcon.title = "Name copied";
                copyIcon.style.color = "";
            }, 2000);
        });
}

export function setupAutoFormat(inputId) {
    const input = document.getElementById(inputId);

    input.addEventListener("input", () => {
        const oldValue = input.value;
        const cursor = input.selectionStart;

        const raw = oldValue.replace(/\D/g, "");

        if (!raw) {
            input.value = "";
            return;
        }

        const formatted = raw.replace(/\B(?=(\d{3})+(?!\d))/g, ".");

        const leftCountBefore = (oldValue.slice(0, cursor).match(/\D/g) || []).length;
        const leftCountAfter = (formatted.slice(0, cursor).match(/\D/g) || []).length;
        const diff = leftCountAfter - leftCountBefore;

        input.value = formatted;

        const newCursor = cursor + diff;
        input.setSelectionRange(newCursor, newCursor);
    });
}

export function copyByElementText(elId, iconId) {
    const el = document.getElementById(elId);
    const icon = document.getElementById(iconId);
    const text = (el?.value ?? el?.textContent ?? "").trim();

    if (!text) {
        console.warn("Nothing to copy");
        return Promise.resolve();
    }

    const applySuccessState = (ico) => {
        if (!ico) return;
        // ursprünglichen FA-Style merken (regular/solid)
        const wasRegular = ico.classList.contains("fa-regular");
        ico.dataset.faStyle = wasRegular ? "regular" : "solid";

        // auf Check umschalten (immer solid)
        ico.classList.remove("fa-regular", "fa-copy");
        ico.classList.add("fa-solid", "fa-check");
        ico.title = "Copied!";
        ico.style.color = "green";

        setTimeout(() => {
            // zurück zu Copy + ursprünglichen Style
            ico.classList.remove("fa-check");
            ico.classList.add("fa-copy");

            if (ico.dataset.faStyle === "regular") {
                ico.classList.remove("fa-solid");
                ico.classList.add("fa-regular");
            } else {
                ico.classList.add("fa-solid"); // bleibt solid
            }

            ico.title = "Copy";
            ico.style.color = "";
            delete ico.dataset.faStyle;
        }, 2000);
    };

    const applyErrorState = (ico) => {
        if (!ico) return;
        ico.title = "Copy failed";
        ico.style.color = "red";
        setTimeout(() => {
            ico.title = "Copy";
            ico.style.color = "";
        }, 2000);
    };

    return navigator.clipboard
        .writeText(text)
        .then(() => applySuccessState(icon))
        .catch((err) => {
            console.error("Error with copying:", err);
            applyErrorState(icon);
        });
}