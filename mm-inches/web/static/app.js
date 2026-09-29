const inputs = [...document.querySelectorAll(".measurement input")];
const decimalPlaces = document.querySelector("#decimal-places");
const fractionDenominator = document.querySelector("#fraction-denominator");
const status = document.querySelector("#status");
const unitToggles = [...document.querySelectorAll("[data-unit-toggle]")];
const selectedUnitCount = document.querySelector("#selected-unit-count");
const restoreDefaultUnits = document.querySelector("#restore-default-units");
const defaultUnits = ["millimeters", "fractional-inches", "decimal-inches"];
const unitPreferenceKey = "benchmarks.displayedUnits";
let sourceInput = document.querySelector("#millimeters");
let requestNumber = 0;
let debounceTimer;

function savedUnits() {
  try {
    const saved = JSON.parse(window.localStorage.getItem(unitPreferenceKey));
    const available = new Set(unitToggles.map((toggle) => toggle.dataset.unitToggle));
    if (Array.isArray(saved)) {
      const valid = saved.filter((unit) => available.has(unit));
      if (valid.length) return valid;
    }
  } catch (_) {
    // Browser storage is optional; the default selection still works without it.
  }
  return defaultUnits;
}

function storeUnits(units) {
  try {
    window.localStorage.setItem(unitPreferenceKey, JSON.stringify(units));
  } catch (_) {
    // Keep the current selection for this page when storage is unavailable.
  }
}

function applyDisplayedUnits({ save = false } = {}) {
  const selected = unitToggles.filter((toggle) => toggle.checked).map((toggle) => toggle.dataset.unitToggle);
  const selectedSet = new Set(selected);
  const visibleRows = [];

  document.querySelectorAll(".measurement").forEach((row) => {
    row.hidden = !selectedSet.has(row.dataset.unit);
    row.classList.remove("last-visible");
    if (!row.hidden) visibleRows.push(row);
  });
  visibleRows.at(-1)?.classList.add("last-visible");
  selectedUnitCount.textContent = selected.length;

  if (sourceInput.closest(".measurement").hidden) {
    const nextInput = visibleRows[0]?.querySelector("input");
    if (nextInput?.value.trim()) setActive(nextInput);
  }
  if (save) storeUnits(selected);
}

function setActive(input) {
  sourceInput = input;
  document.querySelectorAll(".measurement").forEach((row) => {
    row.classList.toggle("active", row.contains(input));
  });
}

async function convert() {
  const thisRequest = ++requestNumber;
  const value = sourceInput.value.trim();
  if (!value) {
    status.textContent = "Enter a value to convert.";
    status.className = "status";
    return;
  }

  try {
    const response = await fetch("/api/convert", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        unit: sourceInput.id,
        value,
        decimalPlaces: Number(decimalPlaces.value),
        fractionDenominator: Number(fractionDenominator.value),
      }),
    });
    const data = await response.json();
    if (thisRequest !== requestNumber) return;
    if (!response.ok) throw new Error(data.error || "That value could not be converted.");

    inputs.forEach((input) => {
      if (input !== sourceInput) input.value = data.values[input.id];
    });
    if (sourceInput.closest(".measurement").hidden) {
      const nextInput = document.querySelector(".measurement:not([hidden]) input");
      if (nextInput) setActive(nextInput);
    }
    status.textContent = `Using ${decimalPlaces.value} decimal places and the nearest 1/${fractionDenominator.value} inch.`;
    status.className = "status";
  } catch (error) {
    if (thisRequest !== requestNumber) return;
    status.textContent = error.message;
    status.className = "status error";
  }
}

function queueConversion() {
  window.clearTimeout(debounceTimer);
  debounceTimer = window.setTimeout(convert, 120);
}

inputs.forEach((input) => {
  input.addEventListener("focus", () => setActive(input));
  input.addEventListener("input", () => {
    setActive(input);
    queueConversion();
  });
});

decimalPlaces.addEventListener("change", convert);
fractionDenominator.addEventListener("change", convert);

const initialUnits = new Set(savedUnits());
unitToggles.forEach((toggle) => {
  toggle.checked = initialUnits.has(toggle.dataset.unitToggle);
  toggle.addEventListener("change", () => {
    const selectedCount = unitToggles.filter((candidate) => candidate.checked).length;
    if (selectedCount === 0) {
      toggle.checked = true;
      status.textContent = "Keep at least one unit displayed.";
      status.className = "status error";
      return;
    }
    applyDisplayedUnits({ save: true });
  });
});

restoreDefaultUnits.addEventListener("click", () => {
  const defaults = new Set(defaultUnits);
  unitToggles.forEach((toggle) => {
    toggle.checked = defaults.has(toggle.dataset.unitToggle);
  });
  applyDisplayedUnits({ save: true });
});

applyDisplayedUnits();
convert();
