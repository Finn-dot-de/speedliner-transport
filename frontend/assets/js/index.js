import {copyContractName, setupAutoFormat} from "./utils.js";
import {calculator, setRoutesData} from "./calculator.js";
import {renderPricing} from "./pricing.js";
import {loadUser} from "./user.js";


export async function loadRoutes() {
    try {
        const res = await fetch("/app/routes", {credentials: "include"});
        const routes = await res.json();
        setRoutesData(routes);
        renderPricing(routes);
        ["route", "volume", "collateral"].forEach(id =>
            document.getElementById(id).addEventListener("input", calculator)
        );
    } catch (err) {
        console.error("Failed to load routes:", err);
    }
}

setupAutoFormat("volume");
setupAutoFormat("collateral");

["route", "volume", "collateral"].forEach((id) => {
    document.getElementById(id).addEventListener("input", calculator);
});


/* ===== Custom Select für #route ===== */
export let mrRouteUI = null;

export function buildMrRouteSelect(selectEl) {
    const wrap = selectEl.closest('.select-wrap') || selectEl.parentElement;

    if (mrRouteUI?.wrap === wrap) {
        return refreshMrOptions(selectEl, mrRouteUI);
    }

    selectEl.classList.add('is-hidden');

    const trigger = document.createElement('button');
    trigger.type = 'button';
    trigger.className = 'mr-select-trigger';
    trigger.setAttribute('aria-haspopup', 'listbox');
    trigger.setAttribute('aria-expanded', 'false');

    const list = document.createElement('ul');
    list.className = 'mr-select-list';
    list.setAttribute('role', 'listbox');

    wrap.appendChild(trigger);
    wrap.appendChild(list);

    mrRouteUI = {wrap, selectEl, trigger, list};
    refreshMrOptions(selectEl, mrRouteUI);

    trigger.addEventListener('click', () => {
        const open = list.classList.toggle('open');
        trigger.setAttribute('aria-expanded', open ? 'true' : 'false');
        if (open) {
            const sel = list.querySelector('.mr-option.is-selected') || list.querySelector('.mr-option');
            sel?.scrollIntoView({block: 'nearest'});
            sel?.focus();
        }
    });

    document.addEventListener('click', (e) => {
        if (!wrap.contains(e.target)) {
            list.classList.remove('open');
            trigger.setAttribute('aria-expanded', 'false');
        }
    });

    trigger.addEventListener('keydown', (e) => {
        if (e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            trigger.click();
        }
    });

    list.addEventListener('keydown', (e) => {
        const items = Array.from(list.querySelectorAll('.mr-option:not(.is-disabled)'));
        const i = items.indexOf(document.activeElement);
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            (items[i + 1] || items[0]).focus();
        }
        if (e.key === 'ArrowUp') {
            e.preventDefault();
            (items[i - 1] || items[items.length - 1]).focus();
        }
        if (e.key === 'Enter') {
            e.preventDefault();
            document.activeElement?.click();
        }
        if (e.key === 'Escape') {
            e.preventDefault();
            trigger.click();
            trigger.focus();
        }
    });
}

export function refreshMrOptions(selectEl, ui) {
    const {trigger, list} = ui;

    const selected = selectEl.options[selectEl.selectedIndex];

    const text = (o) => (o?.textContent || '');
    const hasCorp = /Corp/.test(text(selected));
    const hasNoColl = /No collateral/.test(text(selected));
    const baseLabel = (text(selected) || 'Select route...').replace(/\s+—\s+.*/, '').trim();

    trigger.innerHTML = `
    <span class="mr-trigger-left">
      <span class="mr-label">${baseLabel || 'Select route...'}</span>
      <span class="mr-flags">
        ${hasCorp ? '<span class="badge-corp">🔒 Corp</span>' : ''}
        ${hasNoColl ? '<span class="badge-nocoll">No collateral</span>' : ''}
      </span>
    </span>
  `;

    list.innerHTML = '';
    [...selectEl.options].forEach(opt => {
        if (opt.value === '') return; // Placeholder ausblenden

        const li = document.createElement('li');
        li.className = 'mr-option';
        li.setAttribute('tabindex', '0');
        li.setAttribute('role', 'option');
        if (opt.disabled) li.classList.add('is-disabled');
        if (opt.selected) li.classList.add('is-selected');

        const label = (opt.textContent || '').replace(/\s+—\s+.*/, '').trim();
        const corp = /Corp/.test(opt.textContent || '');
        const nocoll = /No collateral/.test(opt.textContent || '');

        li.innerHTML = `
      <span class="text">${label}</span>
      <span style="margin-left:auto; display:inline-flex; gap:.35rem;">
        ${corp ? '<span class="badge-corp">🔒 Corp</span>' : ''}
        ${nocoll ? '<span class="badge-nocoll">No collateral</span>' : ''}
      </span>
    `;

        li.addEventListener('click', () => {
            if (opt.disabled) return;
            selectEl.value = opt.value;
            selectEl.dispatchEvent(new Event('change', {bubbles: true}));
            list.classList.remove('open');
            trigger.setAttribute('aria-expanded', 'false');

            refreshMrOptions(selectEl, ui);
        });

        list.appendChild(li);
    });
}

const CONTACT_NAME = "Apple Adven";

const contactNameEl = document.getElementById("contactDisplayName");
if (contactNameEl) {
    contactNameEl.textContent = CONTACT_NAME;

    const copyBtn = document.getElementById("copyContactBtn");
    copyBtn?.addEventListener("click", async () => {
        try {
            await navigator.clipboard.writeText(CONTACT_NAME);
            const old = copyBtn.innerHTML;
            copyBtn.innerHTML = '<i class="fa-solid fa-check"></i> Kopiert';
            setTimeout(() => (copyBtn.innerHTML = old), 1400);
        } catch {
            alert("Could not copy the name to the clipboard.");
        }
    });
}

window.copyContractName = copyContractName;
loadRoutes();
loadUser();


