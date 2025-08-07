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
