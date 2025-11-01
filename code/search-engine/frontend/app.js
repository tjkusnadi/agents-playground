const apiBase = window.location.origin.includes("http")
  ? `${window.location.origin}/api`
  : "/api";

let currentPage = 1;
let totalPages = 1;
let currentQuery = "";
let currentPageSize = 5;

const resultsContainer = document.getElementById("results");
const pageInfo = document.getElementById("page-info");
const prevPageBtn = document.getElementById("prev-page");
const nextPageBtn = document.getElementById("next-page");

async function searchMovies() {
  const params = new URLSearchParams({
    page: currentPage,
    pageSize: currentPageSize,
  });
  if (currentQuery.trim()) {
    params.set("q", currentQuery.trim());
  }

  togglePaginationButtons(true);
  try {
    const response = await fetch(`${apiBase}/movies?${params.toString()}`);
    if (!response.ok) {
      throw new Error("Search failed");
    }
    const data = await response.json();
    renderResults(data.movies);
    updatePagination(data.pagination);
  } catch (error) {
    resultsContainer.innerHTML = `<p class="error">${error.message}</p>`;
    pageInfo.textContent = "";
  } finally {
    togglePaginationButtons(false);
  }
}

function renderResults(movies) {
  resultsContainer.innerHTML = "";
  if (!movies || movies.length === 0) {
    resultsContainer.innerHTML = "<p>No movies found. Try another search.</p>";
    return;
  }

  const template = document.getElementById("movie-template");
  movies.forEach((movie) => {
    const node = template.content.cloneNode(true);
    node.querySelector(".title").textContent = movie.title;
    node.querySelector(
      ".meta"
    ).textContent = `${movie.genre || "Unknown genre"} • Rating ${
      movie.rating ?? "n/a"
    } • ${movie.release_year || "Year n/a"}`;
    node.querySelector(".description").textContent = movie.description || "";
    node.querySelector(
      ".identifier"
    ).textContent = `Document ID: ${movie.id}`;
    resultsContainer.appendChild(node);
  });
}

function updatePagination(pagination) {
  if (!pagination) {
    pageInfo.textContent = "";
    prevPageBtn.disabled = true;
    nextPageBtn.disabled = true;
    return;
  }
  currentPage = pagination.page;
  totalPages = pagination.total_pages || 1;
  currentPageSize = pagination.page_size || currentPageSize;

  pageInfo.textContent = `Page ${currentPage} of ${Math.max(totalPages, 1)} (${pagination.total_hits} results)`;
  prevPageBtn.disabled = currentPage <= 1;
  nextPageBtn.disabled = currentPage >= totalPages;
}

function togglePaginationButtons(disabled) {
  prevPageBtn.disabled = disabled || currentPage <= 1;
  nextPageBtn.disabled = disabled || currentPage >= totalPages;
}

function readForm(form) {
  const data = new FormData(form);
  const payload = {};
  data.forEach((value, key) => {
    if (value === "") return;
    if (key === "rating" || key === "release_year") {
      const numeric = Number(value);
      if (!Number.isNaN(numeric)) {
        payload[key] = numeric;
      }
    } else {
      payload[key] = value;
    }
  });
  return payload;
}

function setStatus(target, message, type = "") {
  const el = document.querySelector(`.status[data-target="${target}"]`);
  if (!el) return;
  el.textContent = message;
  el.classList.remove("success", "error");
  if (type) {
    el.classList.add(type);
  }
}

async function handleCreate(event) {
  event.preventDefault();
  const payload = readForm(event.target);
  setStatus("create", "Creating movie...");
  try {
    const response = await fetch(`${apiBase}/movies`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || "Unable to create movie");
    }
    const movie = await response.json();
    setStatus("create", `Created movie with id ${movie.id}`, "success");
    event.target.reset();
    searchMovies();
  } catch (error) {
    setStatus("create", error.message, "error");
  }
}

async function handleLoadMovie() {
  const form = document.getElementById("update-form");
  const id = form.querySelector('input[name="id"]').value.trim();
  if (!id) {
    setStatus("update", "Enter a movie ID to load", "error");
    return;
  }
  setStatus("update", "Loading movie...");
  try {
    const response = await fetch(`${apiBase}/movies/${id}`);
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || "Movie not found");
    }
    const movie = await response.json();
    form.querySelector('input[name="title"]').value = movie.title || "";
    form.querySelector('textarea[name="description"]').value =
      movie.description || "";
    form.querySelector('input[name="genre"]').value = movie.genre || "";
    form.querySelector('input[name="rating"]').value =
      movie.rating ?? "";
    form.querySelector('input[name="release_year"]').value =
      movie.release_year ?? "";
    setStatus("update", "Movie loaded", "success");
  } catch (error) {
    setStatus("update", error.message, "error");
  }
}

async function handleUpdate(event) {
  event.preventDefault();
  const form = event.target;
  const id = form.querySelector('input[name="id"]').value.trim();
  if (!id) {
    setStatus("update", "Movie ID is required", "error");
    return;
  }
  const payload = readForm(form);
  delete payload.id;
  setStatus("update", "Updating movie...");
  try {
    const response = await fetch(`${apiBase}/movies/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || "Unable to update movie");
    }
    setStatus("update", "Movie updated", "success");
    searchMovies();
  } catch (error) {
    setStatus("update", error.message, "error");
  }
}

async function handleDelete(event) {
  event.preventDefault();
  const id = event.target.querySelector('input[name="id"]').value.trim();
  if (!id) {
    setStatus("delete", "Movie ID is required", "error");
    return;
  }
  setStatus("delete", "Deleting movie...");
  try {
    const response = await fetch(`${apiBase}/movies/${id}`, {
      method: "DELETE",
    });
    if (response.status === 404) {
      throw new Error("Movie not found");
    }
    if (!response.ok) {
      throw new Error("Unable to delete movie");
    }
    setStatus("delete", "Movie deleted", "success");
    event.target.reset();
    searchMovies();
  } catch (error) {
    setStatus("delete", error.message, "error");
  }
}

function setupEventListeners() {
  document.getElementById("search-form").addEventListener("submit", (event) => {
    event.preventDefault();
    currentQuery = document.getElementById("search-query").value;
    currentPageSize = Number(document.getElementById("page-size").value);
    currentPage = 1;
    searchMovies();
  });

  prevPageBtn.addEventListener("click", () => {
    if (currentPage > 1) {
      currentPage -= 1;
      searchMovies();
    }
  });

  nextPageBtn.addEventListener("click", () => {
    if (currentPage < totalPages) {
      currentPage += 1;
      searchMovies();
    }
  });

  document.getElementById("create-form").addEventListener("submit", handleCreate);
  document.getElementById("update-form").addEventListener("submit", handleUpdate);
  document.getElementById("delete-form").addEventListener("submit", handleDelete);
  document.getElementById("load-movie").addEventListener("click", handleLoadMovie);
}

document.addEventListener("DOMContentLoaded", () => {
  setupEventListeners();
  searchMovies();
});
