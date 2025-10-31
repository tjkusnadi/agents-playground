const API_BASE = new URL("/api", window.location.origin).toString();

const countriesList = document.getElementById("countriesList");
const countryTemplate = document.getElementById("countryTemplate");
const placeTemplate = document.getElementById("placeTemplate");
const refreshBtn = document.getElementById("refreshBtn");

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

  if (!data.length) {
    countriesList.innerHTML = '<p class="empty">No destinations yet. Check back soon!</p>';
    return;
  }

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
  }
}

async function loadCountries() {
  refreshBtn.disabled = true;
  refreshBtn.textContent = "Loading...";
  try {
    const data = await fetchJSON(`${API_BASE}/countries`);
    renderCountries(data);
  } catch (error) {
    countriesList.innerHTML = `<p class="empty error">${error.message}</p>`;
  } finally {
    refreshBtn.disabled = false;
    refreshBtn.textContent = "Refresh";
  }
}

refreshBtn.addEventListener("click", loadCountries);

loadCountries();
