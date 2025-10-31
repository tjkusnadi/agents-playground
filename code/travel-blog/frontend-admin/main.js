const API_BASE = new URL("/api", window.location.origin).toString();

const countriesList = document.getElementById("countriesList");
const countryTemplate = document.getElementById("countryTemplate");
const placeTemplate = document.getElementById("placeTemplate");
const refreshBtn = document.getElementById("refreshBtn");
const adminAlerts = document.getElementById("adminAlerts");

async function fetchJSON(url, options = {}) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed with status ${response.status}`);
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

function renderAlert(type, message) {
  const el = document.createElement("div");
  el.className = `alert ${type}`;
  el.textContent = message;
  adminAlerts.prepend(el);
  setTimeout(() => {
    el.remove();
  }, 4000);
}

function formatDate(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function renderCountries(data) {
  countriesList.innerHTML = "";
  const select = document.getElementById("placeCountry");
  select.innerHTML = "";

  if (!data.length) {
    countriesList.innerHTML = '<p class="empty">No destinations yet. Add your first country below!</p>';
    const option = document.createElement("option");
    option.value = "";
    option.textContent = "Add a country first";
    select.appendChild(option);
    select.disabled = true;
    return;
  }

  select.disabled = false;

  for (const country of data) {
    const clone = countryTemplate.content.cloneNode(true);
    clone.querySelector(".country-name").textContent = country.name;
    clone.querySelector(".country-description").textContent = country.description || "";

    const placesList = clone.querySelector(".places");
    if (!country.places || !country.places.length) {
      const empty = document.createElement("li");
      empty.className = "place-item";
      empty.innerHTML = "<p class=\"place-description\">No places added yet.</p>";
      placesList.appendChild(empty);
    } else {
      for (const place of country.places) {
        const placeNode = placeTemplate.content.cloneNode(true);
        placeNode.querySelector(".place-name").textContent = place.name;
        const metaPieces = [place.category];
        if (place.city) metaPieces.push(place.city);
        if (place.visited_at) metaPieces.push(`Visited ${formatDate(place.visited_at)}`);
        placeNode.querySelector(".place-meta").textContent = metaPieces.filter(Boolean).join(" • ");
        placeNode.querySelector(".place-description").textContent = place.description || "";
        placesList.appendChild(placeNode);
      }
    }

    countriesList.appendChild(clone);

    const option = document.createElement("option");
    option.value = country.id;
    option.textContent = country.name;
    select.appendChild(option);
  }
}

async function loadCountries() {
  refreshBtn.disabled = true;
  refreshBtn.textContent = "Loading...";
  try {
    const data = await fetchJSON(`${API_BASE}/countries`);
    renderCountries(data);
  } catch (error) {
    renderAlert("error", error.message);
  } finally {
    refreshBtn.disabled = false;
    refreshBtn.textContent = "Refresh";
  }
}

async function handleCountrySubmit(event) {
  event.preventDefault();
  const payload = {
    name: document.getElementById("countryName").value.trim(),
    description: document.getElementById("countryDescription").value.trim(),
  };

  if (!payload.name) {
    renderAlert("error", "Country name is required");
    return;
  }

  try {
    await fetchJSON(`${API_BASE}/countries`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
    renderAlert("success", `Added ${payload.name}`);
    event.target.reset();
    await loadCountries();
  } catch (error) {
    renderAlert("error", error.message);
  }
}

async function handlePlaceSubmit(event) {
  event.preventDefault();
  const countryId = document.getElementById("placeCountry").value;
  const payload = {
    name: document.getElementById("placeName").value.trim(),
    category: document.getElementById("placeCategory").value.trim(),
    city: document.getElementById("placeCity").value.trim(),
    description: document.getElementById("placeDescription").value.trim(),
    visited_at: document.getElementById("placeVisitedAt").value || undefined,
  };

  if (!countryId || !payload.name || !payload.category) {
    renderAlert("error", "Please fill out the required fields");
    return;
  }

  try {
    await fetchJSON(`${API_BASE}/countries/${countryId}/places`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
    renderAlert("success", `Added ${payload.name}`);
    event.target.reset();
    await loadCountries();
  } catch (error) {
    renderAlert("error", error.message);
  }
}

refreshBtn.addEventListener("click", loadCountries);
document.getElementById("countryForm").addEventListener("submit", handleCountrySubmit);
document.getElementById("placeForm").addEventListener("submit", handlePlaceSubmit);

loadCountries();
