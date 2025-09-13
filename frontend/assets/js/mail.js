const MAIL_COOKIE_NAME = "kjdlfghkdsfgl";

function setCookie(name, value, minutes){
    const d = new Date();
    d.setTime(d.getTime() + minutes * 60 * 1000);
    document.cookie = `${name}=${encodeURIComponent(value)}; expires=${d.toUTCString()}; path=/; samesite=lax`;
}
function getCookie(name){
    return document.cookie
        .split(";")
        .map(c => c.trim())
        .find(c => c.startsWith(name + "="))
        ?.split("=")[1] || null;
}
function getCooldownRemainingMs(){
    const until = getCookie(MAIL_COOKIE_NAME);
    if (!until) return 0;
    const ms = Number(decodeURIComponent(until)) - Date.now();
    return ms > 0 ? ms : 0;
}

/**
 * Init zentrales Mail-Modal mit festen Empfängern (keine User-Eingabe)
 * requiredRecipients: Array<{id:number, type:"character"|"corporation"|"alliance"|"mailing_list"}>
 */
export function initMailModal({
                                  cooldownMin = 5,
                                  requiredRecipients = [],
                                  extraRecipients = [],
                                  selectors = {
                                      overlay: "#mailModal",
                                      openBtn: "#openMailModalBtn",
                                      closeBtn: "#closeMailModalBtn",
                                      form: "#mailForm",
                                      sendBtn: "#sendMailBtn",
                                      cooldownInfo: "#mailCooldownInfo",
                                      feedback: "#mailFeedback",
                                      recipientsTags: "#recipientsTags",
                                      subject: "#mailSubject",
                                      body: "#mailBody",
                                  }
                              } = {}){
    const $ = (s)=>document.querySelector(s);

    const modal        = $(selectors.overlay);
    const openBtn      = $(selectors.openBtn);
    const closeBtn     = $(selectors.closeBtn);
    const form         = $(selectors.form);
    const sendBtn      = $(selectors.sendBtn);
    const cooldownInfo = $(selectors.cooldownInfo);
    const feedback     = $(selectors.feedback);
    const tagsWrap     = $(selectors.recipientsTags);
    const inpSubject   = $(selectors.subject);
    const taBody       = $(selectors.body);

    if (!modal || !openBtn || !closeBtn || !form) {
        console.warn("[mail] Modal elements missing – init aborted.");
        return;
    }

    const recipients = [
        ...(requiredRecipients || []).map(r => ({ id: Number(r.id), type: r.type || "character", locked: true })),
        ...(extraRecipients || []).map(r => ({ id: Number(r.id), type: r.type || "character", locked: true })),
    ].filter(r => Number.isFinite(r.id) && r.id > 0);

    function renderTags(){
        if (!tagsWrap) return;
        if (recipients.length === 0){
            tagsWrap.innerHTML = `<span style="color:var(--muted);">No receivers configured.</span>`;
            return;
        }
        tagsWrap.innerHTML = recipients.map(r =>
            `<span class="tag">🔒 <strong>${r.type}</strong> #${r.id}</span>`
        ).join("");
    }
    renderTags();

    let cooldownTimer = null;
    function updateCooldownUI(){
        const ms = getCooldownRemainingMs();
        const active = ms > 0;
        if (sendBtn) sendBtn.disabled = active;
        if (cooldownInfo){
            const m = String(Math.floor((ms/1000)/60)).padStart(2,"0");
            const s = String(Math.ceil((ms/1000)%60)).padStart(2,"0");
            cooldownInfo.textContent = active ? `Cooldown active: ${m}:${s}` : "";
        }
    }
    function startCooldownTicker(){
        if (cooldownTimer) clearInterval(cooldownTimer);
        cooldownTimer = setInterval(updateCooldownUI, 500);
        updateCooldownUI();
    }

    let lastFocused = null;
    function openModal(){
        lastFocused = document.activeElement;
        modal.classList.add("open");
        modal.setAttribute("aria-hidden", "false");
        startCooldownTicker();
        requestAnimationFrame(()=> inpSubject?.focus());
    }
    function closeModal(){
        modal.classList.remove("open");
        modal.setAttribute("aria-hidden", "true");
        if (lastFocused) lastFocused.focus();
    }
    openBtn.addEventListener("click", openModal);
    closeBtn.addEventListener("click", closeModal);
    modal.addEventListener("click", (e)=>{ if (e.target === modal) closeModal(); });
    document.addEventListener("keydown", (e)=>{ if (e.key === "Escape" && modal.classList.contains("open")) closeModal(); });

    form.addEventListener("submit", async (e)=>{
        e.preventDefault();
        if (feedback){ feedback.style.color = "#66fcf1"; feedback.textContent = "Sending …"; }

        if (getCooldownRemainingMs() > 0){
            if (feedback){ feedback.style.color = "#ff8080"; feedback.textContent = "Please wait until the cooldown has expired."; }
            return;
        }

        const subject = inpSubject.value.trim();
        const body    = taBody.value.trim();
        if (!subject || !body || recipients.length === 0){
            if (feedback){
                feedback.style.color = "#ff8080";
                feedback.textContent = !subject || !body
                    ? "Subject and message are mandatory."
                    : "No recipients are configured.";
            }
            return;
        }

        const payload = {
            autoApproveCspa: false,
            subject,
            body,
            recipients: recipients.map(r => ({ id: r.id, type: r.type }))
        };

        try{
            if (sendBtn) sendBtn.disabled = true;
            const res = await fetch("/app/mail", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify(payload)
            });
            if (!res.ok){
                const txt = await res.text().catch(()=> "");
                throw new Error(`${res.status} ${res.statusText} ${txt || ""}`.trim());
            }

            // Cooldown starten
            setCookie(MAIL_COOKIE_NAME, String(Date.now() + cooldownMin*60*1000), cooldownMin);
            startCooldownTicker();

            if (feedback){ feedback.style.color = "#66fcf1"; feedback.textContent = "Email sent. Thanks!"; }
            inpSubject.value = ""; taBody.value = "";
            setTimeout(closeModal, 900);
        }catch(err){
            if (feedback){ feedback.style.color = "#ff8080"; feedback.textContent = "Sending failed: " + err.message; }
        }finally{
            if (sendBtn) sendBtn.disabled = getCooldownRemainingMs() > 0;
        }
    });
}
