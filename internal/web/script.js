const state = {
  tracks: [],
  options: [],
  previewQuery: "",
  previewLinkType: "",
  layoutFields: null,
  config: null,
  jobsLoading: false,
  jobsTimer: null,
  arlCollapsed: null,
  settingsSnapshots: {},
};
const $ = (id) => document.getElementById(id);
const settingsSections = {
  downloads: {
    button: "saveDownloadsBtn",
    inputs: [
      "cfgConcurrency",
      "cfgTrackNumber",
      "cfgFallbackTrack",
      "cfgFallbackQuality",
    ],
    label: "Download",
  },
  layout: {
    button: "saveLayoutBtn",
    inputs: [
      "cfgLayoutTrack",
      "cfgLayoutAlbum",
      "cfgLayoutArtist",
      "cfgLayoutPlaylist",
    ],
    label: "Layout",
  },
  playlist: {
    button: "savePlaylistBtn",
    inputs: ["cfgResolveFullPath"],
    label: "Playlist",
  },
  cover: {
    button: "saveCoverBtn",
    inputs: [
      "cfgCoverMode",
      "cfgCoverFileName",
      "cfgCover128",
      "cfgCover320",
      "cfgCoverFlac",
    ],
    label: "Cover",
  },
};
const api = async (path, opts = {}) => {
  const res = await fetch(path, {
    headers: { "content-type": "application/json" },
    ...opts,
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: res.statusText }));
    const err = new Error(data.error || res.statusText);
    err.status = res.status;
    throw err;
  }
  if (res.status === 204) return null;
  return res.json();
};
const setMainMessage = (text) => {
  $("mainMessage").textContent = text || "";
};
function currentTheme() {
  return document.documentElement.dataset.theme === "dark" ? "dark" : "light";
}
function setTheme(theme) {
  const next = theme === "dark" ? "dark" : "light";
  document.documentElement.dataset.theme = next;
  $("themeToggle").textContent = next === "dark" ? "Light" : "Dark";
  $("themeToggle").setAttribute(
    "aria-label",
    "Switch to " + (next === "dark" ? "light" : "dark") + " theme",
  );
  try {
    localStorage.setItem("d-fi-theme", next);
  } catch (_) {}
}
function toggleTheme() {
  setTheme(currentTheme() === "dark" ? "light" : "dark");
}
function showToast(text, kind = "success") {
  if (!text) return;
  const toast = document.createElement("div");
  toast.className = "toast " + kind;
  toast.textContent = text;
  $("toasts").appendChild(toast);
  window.setTimeout(() => toast.remove(), kind === "error" ? 9000 : 3500);
}
const duration = (seconds) => {
  seconds = Number(seconds || 0);
  const min = Math.floor(seconds / 60);
  const sec = String(seconds % 60).padStart(2, "0");
  return min + ":" + sec;
};
async function loadConfig() {
  try {
    const data = await api("/api/config");
    const cfg = data.config;
    state.config = cfg;
    fillConfig(cfg);
    updateSession(data.session, data.hasArl);
    updateARLStatus(data.hasArl);
    if (data.session && data.session.error) {
      showToast(data.session.error, "error");
    }
  } catch (err) {
    showToast(err.message, "error");
  }
}
function fillConfig(cfg) {
  fillConfigSection("downloads", cfg);
  fillConfigSection("layout", cfg);
  fillConfigSection("playlist", cfg);
  fillConfigSection("cover", cfg);
  snapshotAllSettings();
}
function fillConfigSection(section, cfg) {
  if (section === "downloads") {
    $("cfgConcurrency").value = cfg.concurrency || 4;
    $("cfgTrackNumber").checked = !!cfg.trackNumber;
    $("cfgFallbackTrack").checked = !!cfg.fallbackTrack;
    $("cfgFallbackQuality").checked = !!cfg.fallbackQuality;
    return;
  }
  if (section === "layout") {
    $("cfgLayoutTrack").value = cfg.saveLayout?.track || "";
    $("cfgLayoutAlbum").value = cfg.saveLayout?.album || "";
    $("cfgLayoutArtist").value = cfg.saveLayout?.artist || "";
    $("cfgLayoutPlaylist").value = cfg.saveLayout?.playlist || "";
    return;
  }
  if (section === "playlist") {
    $("cfgResolveFullPath").checked = !!cfg.playlist?.resolveFullPath;
    return;
  }
  if (section === "cover") {
    $("cfgCover128").value = cfg.coverSize?.["128"] || 500;
    $("cfgCover320").value = cfg.coverSize?.["320"] || 500;
    $("cfgCoverFlac").value = cfg.coverSize?.flac || 1000;
    $("cfgCoverMode").value = cfg.cover?.mode || "embed";
    $("cfgCoverFileName").value = cfg.cover?.fileName || "cover.jpg";
  }
}
function readConfigSection(section) {
  if (section === "downloads") {
    return {
      concurrency: Number($("cfgConcurrency").value || 1),
      trackNumber: $("cfgTrackNumber").checked,
      fallbackTrack: $("cfgFallbackTrack").checked,
      fallbackQuality: $("cfgFallbackQuality").checked,
    };
  }
  if (section === "layout") {
    return {
      track: $("cfgLayoutTrack").value,
      album: $("cfgLayoutAlbum").value,
      artist: $("cfgLayoutArtist").value,
      playlist: $("cfgLayoutPlaylist").value,
    };
  }
  if (section === "playlist") {
    return {
      resolveFullPath: $("cfgResolveFullPath").checked,
    };
  }
  if (section === "cover") {
    return {
      coverSize: {
        128: Number($("cfgCover128").value || 500),
        320: Number($("cfgCover320").value || 500),
        flac: Number($("cfgCoverFlac").value || 1000),
      },
      cover: {
        mode: $("cfgCoverMode").value || "embed",
        fileName: $("cfgCoverFileName").value || "cover.jpg",
      },
    };
  }
  return {};
}
function snapshotAllSettings() {
  Object.keys(settingsSections).forEach(snapshotSettingsSection);
}
function snapshotSettingsSection(section) {
  state.settingsSnapshots[section] = JSON.stringify(readConfigSection(section));
  syncSettingsButton(section);
}
function syncSettingsButtons() {
  Object.keys(settingsSections).forEach(syncSettingsButton);
}
function syncSettingsButton(section) {
  const snapshot = state.settingsSnapshots[section];
  const current = JSON.stringify(readConfigSection(section));
  $(settingsSections[section].button).disabled = snapshot === current;
}
function cloneSavedConfig() {
  return JSON.parse(JSON.stringify(state.config || {}));
}
function readConfigForSection(section) {
  const cfg = cloneSavedConfig();
  const values = readConfigSection(section);
  cfg.cookies = {
    ...(cfg.cookies || {}),
    arl: "",
  };
  if (section === "downloads") {
    cfg.concurrency = values.concurrency;
    cfg.trackNumber = values.trackNumber;
    cfg.fallbackTrack = values.fallbackTrack;
    cfg.fallbackQuality = values.fallbackQuality;
  } else if (section === "layout") {
    cfg.saveLayout = values;
  } else if (section === "playlist") {
    cfg.playlist = values;
  } else if (section === "cover") {
    cfg.coverSize = values.coverSize;
    cfg.cover = values.cover;
  }
  return cfg;
}
function readConfigWithARL(arl) {
  const cfg = cloneSavedConfig();
  cfg.cookies = {
    ...(cfg.cookies || {}),
    arl,
  };
  return cfg;
}
function updateSession(session, hasArl) {
  const dot = $("sessionDot");
  const text = $("sessionText");
  dot.className = "dot" + (session && session.ready ? " ready" : "");
  if (session && session.ready)
    text.textContent = "Connected as " + session.userName;
  else if (session && session.error) text.textContent = "Connection failed";
  else if (hasArl) text.textContent = "ARL saved, not connected";
  else text.textContent = "Not connected";
}
function updateARLStatus(hasArl) {
  $("arlDot").className = "arl-dot" + (hasArl ? " saved" : "");
  $("arlText").textContent = hasArl ? "ARL saved" : "No ARL saved";
  if (state.arlCollapsed === null) {
    setARLCollapsed(hasArl);
  }
}
function updateSaveARLButton() {
  $("saveArlBtn").disabled = $("arl").value.trim() === "";
}
function setARLCollapsed(collapsed) {
  state.arlCollapsed = collapsed;
  $("arlBody").hidden = collapsed;
  $("toggleArlBtn").textContent = collapsed ? "Edit" : "Hide";
  $("toggleArlBtn").setAttribute("aria-expanded", String(!collapsed));
  document
    .querySelector(".arl-section")
    .classList.toggle("collapsed", collapsed);
}
function toggleARL() {
  setARLCollapsed(!state.arlCollapsed);
}
async function saveConfig(section) {
  try {
    const data = await api("/api/config", {
      method: "PUT",
      body: JSON.stringify(readConfigForSection(section)),
    });
    state.config = data.config;
    fillConfigSection(section, data.config);
    updateSession(data.session, data.hasArl);
    updateARLStatus(data.hasArl);
    if (data.session && data.session.error) {
      showToast(data.session.error, "error");
      return;
    }
    snapshotSettingsSection(section);
    syncSettingsButtons();
    showToast(settingsSections[section].label + " settings saved.");
  } catch (err) {
    showToast(err.message, "error");
  }
}
async function saveARL() {
  const arl = $("arl").value.trim();
  if (!arl) {
    showToast("Paste ARL first.", "error");
    return;
  }
  $("saveArlBtn").disabled = true;
  try {
    const data = await api("/api/config", {
      method: "PUT",
      body: JSON.stringify(readConfigWithARL(arl)),
    });
    state.config = data.config;
    $("arl").value = "";
    updateSaveARLButton();
    updateSession(data.session, data.hasArl);
    updateARLStatus(data.hasArl);
    if (data.hasArl) {
      setARLCollapsed(true);
    }
    if (data.session && data.session.error) {
      showToast(data.session.error, "error");
      return;
    }
    showToast("ARL saved.");
  } catch (err) {
    updateSaveARLButton();
    showToast(err.message, "error");
  }
}
function bindConfigControls() {
  Object.entries(settingsSections).forEach(([section, settings]) => {
    settings.inputs.forEach((id) => {
      $(id).addEventListener("input", () => syncSettingsButton(section));
      $(id).addEventListener("change", () => syncSettingsButton(section));
    });
  });
  $("arl").addEventListener("input", updateSaveARLButton);
  $("arl").addEventListener("keydown", (event) => {
    if (event.key === "Enter") saveARL();
  });
}
async function preview() {
  setMainMessage("Fetching preview...");
  try {
    state.previewQuery = "";
    state.previewLinkType = "";
    state.layoutFields = null;
    state.options = [];
    state.tracks = [];
    renderOptions();
    renderTracks();
    if (needsOptionSelection()) {
      const data = await api("/api/search-options", {
        method: "POST",
        body: JSON.stringify({
          type: $("queryType").value,
          query: $("query").value.trim(),
        }),
      });
      state.options = data.options || [];
      renderOptions();
      renderTracks();
      setMainMessage("");
      showToast(state.options.length + " results found.");
      return;
    }
    const query = buildQuery();
    const data = await api("/api/preview", {
      method: "POST",
      body: JSON.stringify({ query }),
    });
    state.previewQuery = query;
    state.previewLinkType = data.linkType || "";
    state.layoutFields = data.layoutFields || null;
    state.tracks = data.tracks || [];
    renderOptions();
    renderTracks();
    setMainMessage("");
    showToast(state.tracks.length + " tracks found.");
  } catch (err) {
    showToast(err.message, "error");
    setMainMessage("");
  }
}
function renderTracks() {
  const body = $("tracksBody");
  if (!state.tracks.length) {
    body.innerHTML = "";
    syncSelectAllTracks();
    return;
  }
  body.innerHTML = state.tracks
    .map(
      (track) =>
        "<tr>" +
        '<td class="col-select"><input type="checkbox" data-index="' +
        track.index +
        '" checked></td>' +
        '<td class="col-position">' +
        (track.position || track.index + 1) +
        "</td>" +
        '<td title="' +
        escapeHTML(track.title) +
        '">' +
        escapeHTML(track.title) +
        "</td>" +
        '<td title="' +
        escapeHTML(track.artist) +
        '">' +
        escapeHTML(track.artist) +
        "</td>" +
        '<td title="' +
        escapeHTML(track.album) +
        '">' +
        escapeHTML(track.album) +
        "</td>" +
        "<td>" +
        duration(track.duration) +
        "</td>" +
        "</tr>",
    )
    .join("");
  body.querySelectorAll("[data-index]").forEach((checkbox) => {
    checkbox.addEventListener("change", syncSelectAllTracks);
  });
  syncSelectAllTracks();
}
function syncSelectAllTracks() {
  const selectAll = $("selectAllTracks");
  const downloadBtn = $("downloadSelectedBtn");
  const selectionCount = $("selectionCount");
  const checkboxes = [...document.querySelectorAll("[data-index]")];
  const checked = checkboxes.filter((el) => el.checked).length;
  selectAll.disabled = checkboxes.length === 0;
  selectAll.checked = checkboxes.length > 0 && checked === checkboxes.length;
  selectAll.indeterminate = checked > 0 && checked < checkboxes.length;
  downloadBtn.disabled = checked === 0;
  selectionCount.textContent =
    checked + " of " + checkboxes.length + " selected";
}
function toggleAllTracks() {
  document
    .querySelectorAll("[data-index]")
    .forEach((checkbox) => (checkbox.checked = $("selectAllTracks").checked));
  syncSelectAllTracks();
}
function renderOptions() {
  const previewArea = $("previewArea");
  const box = $("optionsBox");
  const tracksBox = $("tracksBox");
  const previewActions = $("previewActions");
  if (!state.options.length) {
    box.hidden = true;
    box.innerHTML = "";
    tracksBox.hidden = state.tracks.length === 0;
    previewActions.hidden = state.tracks.length === 0;
    previewArea.hidden = state.tracks.length === 0;
    return;
  }
  previewArea.hidden = false;
  box.hidden = false;
  tracksBox.hidden = true;
  previewActions.hidden = true;
  box.innerHTML = state.options
    .map(
      (option, index) =>
        '<button class="option-row" type="button" data-option-index="' +
        index +
        '">' +
        '<div class="option-title">' +
        escapeHTML(option.title) +
        "</div>" +
        '<div class="option-desc">' +
        escapeHTML(option.description) +
        "</div>" +
        "</button>",
    )
    .join("");
  box.querySelectorAll("[data-option-index]").forEach((button) => {
    button.addEventListener("click", async () => {
      const option = state.options[Number(button.dataset.optionIndex)];
      state.previewQuery = option.url;
      setMainMessage("Fetching tracks...");
      try {
        const data = await api("/api/preview", {
          method: "POST",
          body: JSON.stringify({ query: option.url }),
        });
        state.tracks = data.tracks || [];
        state.previewLinkType = data.linkType || "";
        state.layoutFields = data.layoutFields || null;
        state.options = [];
        renderOptions();
        renderTracks();
        setMainMessage("");
        showToast(state.tracks.length + " tracks found.");
      } catch (err) {
        showToast(err.message, "error");
        setMainMessage("");
      }
    });
  });
}
async function startDownload() {
  if (!state.tracks.length) {
    showToast("Preview tracks first.", "error");
    return;
  }
  const selected = [...document.querySelectorAll("[data-index]:checked")].map(
    (el) => Number(el.dataset.index),
  );
  if (!selected.length) {
    showToast("Select at least one track.", "error");
    return;
  }
  const body = {
    query: state.previewQuery,
    quality: $("quality").value,
    tracks: selected,
  };
  setMainMessage("Starting download...");
  try {
    await api("/api/downloads", {
      method: "POST",
      body: JSON.stringify(body),
    });
    setMainMessage("");
    showToast("Download queued.");
    await loadJobs();
  } catch (err) {
    showToast(err.message, "error");
    setMainMessage("");
  }
}
async function loadJobs() {
  if (state.jobsLoading) return;
  state.jobsLoading = true;
  try {
    const data = await api("/api/jobs");
    const jobs = data.jobs || [];
    renderJobs(jobs);
    syncClearHistoryButton(jobs);
    scheduleJobsPoll(jobs.some(isActiveJob) ? 2000 : 5000);
  } catch (err) {
    $("jobs").innerHTML = '<p class="muted">Unable to load downloads</p>';
    syncClearHistoryButton([]);
    showToast(err.message, "error");
    scheduleJobsPoll(5000);
  } finally {
    state.jobsLoading = false;
  }
}
function scheduleJobsPoll(delay) {
  window.clearTimeout(state.jobsTimer);
  state.jobsTimer = window.setTimeout(loadJobs, delay);
}
function isActiveJob(job) {
  return (
    job.status === "queued" ||
    job.status === "running" ||
    job.status === "canceling"
  );
}
function isCancelableJob(job) {
  return job.status === "queued" || job.status === "running";
}
function syncClearHistoryButton(jobs) {
  $("clearHistoryBtn").disabled = !jobs.some((job) => !isActiveJob(job));
}
function buildQuery() {
  const value = $("query").value.trim();
  const type = $("queryType").value;
  if (!value || type === "auto" || looksLikeURL(value)) return value;
  return type + ":" + value;
}
function needsOptionSelection() {
  const value = $("query").value.trim();
  const type = $("queryType").value;
  return value && type !== "auto" && !looksLikeURL(value);
}
function looksLikeURL(value) {
  return (
    value.startsWith("http://") ||
    value.startsWith("https://") ||
    value.startsWith("spotify:")
  );
}
function openLayoutFields() {
  renderLayoutFields();
  $("layoutCopyStatus").textContent = "";
  $("layoutFieldsDialog").showModal();
}
function closeLayoutFields() {
  $("layoutFieldsDialog").close();
}
function renderLayoutFields() {
  const fields = state.layoutFields || defaultLayoutFields();
  $("alwaysFields").innerHTML = renderFieldButtons(fields.always || []);
  const current = fields.current || [];
  $("currentFieldsNote").textContent = current.length
    ? "Fields from the current " +
      (state.previewLinkType || "download") +
      " preview."
    : "Preview a URL or search result to see fields from that response.";
  $("currentFields").innerHTML = renderFieldButtons(current);
  document
    .querySelectorAll("[data-layout-field]")
    .forEach((button) =>
      button.addEventListener("click", () =>
        copyLayoutField(button.dataset.layoutField),
      ),
    );
}
function defaultLayoutFields() {
  return {
    always: [
      { key: "ALB_TITLE", scope: "track" },
      { key: "ART_NAME", scope: "track" },
      { key: "SNG_TITLE", scope: "track" },
      { key: "TRACK_NUMBER", scope: "special" },
      { key: "TRACK_POSITION", scope: "special" },
      { key: "NO_TRACK_NUMBER", scope: "special" },
      { key: "TITLE", scope: "playlist" },
    ],
    current: [],
  };
}
function renderFieldButtons(fields) {
  if (!fields.length) return "";
  return fields
    .map((field) => {
      const placeholder = "{" + field.key + "}";
      return (
        '<button class="field-btn" type="button" data-layout-field="' +
        escapeHTML(field.key) +
        '" title="' +
        escapeHTML(placeholder) +
        '">' +
        '<div class="field-key">' +
        escapeHTML(placeholder) +
        "</div>" +
        '<div class="field-meta">' +
        escapeHTML(field.scope || "") +
        (field.sample ? " · " + escapeHTML(field.sample) : "") +
        "</div>" +
        "</button>"
      );
    })
    .join("");
}
async function copyLayoutField(key) {
  const value = "{" + key + "}";
  try {
    await navigator.clipboard.writeText(value);
    $("layoutCopyStatus").textContent = value + " copied";
  } catch (_) {
    $("layoutCopyStatus").textContent = "Copy failed";
  }
}
function renderJobs(jobs) {
  const root = $("jobs");
  if (!jobs.length) {
    root.innerHTML = '<p class="muted">No downloads yet</p>';
    return;
  }
  root.innerHTML = jobs
    .slice()
    .reverse()
    .map((job) => {
      const files = (job.files || [])
        .slice(-4)
        .map((file) => "<div>" + escapeHTML(file) + "</div>")
        .join("");
      const pct = Math.max(0, Math.min(100, Number(job.progress || 0)));
      const line = jobLine(job);
      const action = isCancelableJob(job)
        ? '<button class="secondary" onclick="cancelJob(' +
          job.id +
          ')">Cancel</button>'
        : job.status === "canceling"
          ? '<button class="secondary" disabled>Canceling</button>'
          : "";
      return (
        '<div class="job">' +
        '<div class="job-head"><strong>#' +
        job.id +
        " " +
        escapeHTML(jobTitle(job)) +
        '</strong><span class="muted">' +
        job.doneTracks +
        "/" +
        job.totalTracks +
        "</span></div>" +
        '<div class="job-source muted" title="' +
        escapeHTML(job.source) +
        '">' +
        escapeHTML(line) +
        "</div>" +
        '<div class="progress"><div class="bar" style="width:' +
        pct.toFixed(1) +
        '%"></div></div>' +
        '<div class="job-actions"><span class="job-current">' +
        pct.toFixed(1) +
        "% complete" +
        "</span>" +
        action +
        "</div>" +
        (job.error
          ? '<div class="error">' + escapeHTML(job.error) + "</div>"
          : "") +
        (files ? '<div class="files">' + files + "</div>" : "") +
        "</div>"
      );
    })
    .join("");
}
function jobTitle(job) {
  if (job.status === "running") return "Downloading";
  if (job.status === "done") return "Done";
  if (job.status === "error") return "Error";
  if (job.status === "canceling") return "Canceling";
  if (job.status === "canceled") return "Canceled";
  if (job.status === "queued") return "Queued";
  return job.status;
}
function jobLine(job) {
  if (job.status === "done") {
    const dirs = uniqueDirs(job.files || []);
    return dirs.length ? "Saved in " + dirs.join(", ") : "Done";
  }
  if (job.status === "running") {
    return job.current ? "Downloading " + job.current : "Preparing download";
  }
  if (job.status === "queued") return "Waiting to start";
  if (job.status === "canceling") return "Canceling download";
  if (job.status === "error") return job.error || "Download failed";
  if (job.status === "canceled") return job.error || "Canceled";
  return job.source;
}
function uniqueDirs(paths) {
  return [...new Set(paths.map(dirName).filter(Boolean))];
}
function dirName(path) {
  const normalized = String(path || "").replaceAll("\\", "/");
  const index = normalized.lastIndexOf("/");
  return index > 0 ? normalized.slice(0, index) : ".";
}
async function cancelJob(id) {
  await api("/api/jobs/" + id + "/cancel", {
    method: "POST",
    body: "{}",
  }).catch((err) => showToast(err.message, "error"));
  await loadJobs();
}
async function clearHistory() {
  await api("/api/jobs", { method: "DELETE" }).catch((err) =>
    showToast(err.message, "error"),
  );
  showToast("History cleared.");
  await loadJobs();
}
function escapeHTML(value) {
  return String(value || "").replace(
    /[&<>"']/g,
    (ch) =>
      ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;",
      })[ch],
  );
}
bindConfigControls();
setTheme(currentTheme());
$("themeToggle").addEventListener("click", toggleTheme);
$("saveArlBtn").addEventListener("click", saveARL);
$("toggleArlBtn").addEventListener("click", toggleARL);
$("saveDownloadsBtn").addEventListener("click", () => saveConfig("downloads"));
$("saveLayoutBtn").addEventListener("click", () => saveConfig("layout"));
$("savePlaylistBtn").addEventListener("click", () => saveConfig("playlist"));
$("saveCoverBtn").addEventListener("click", () => saveConfig("cover"));
$("layoutFieldsBtn").addEventListener("click", openLayoutFields);
$("closeLayoutFieldsBtn").addEventListener("click", closeLayoutFields);
$("layoutFieldsDialog").addEventListener("click", (event) => {
  if (event.target === $("layoutFieldsDialog")) closeLayoutFields();
});
$("previewBtn").addEventListener("click", preview);
$("downloadSelectedBtn").addEventListener("click", startDownload);
$("clearHistoryBtn").addEventListener("click", clearHistory);
$("selectAllTracks").addEventListener("change", toggleAllTracks);
loadConfig();
loadJobs();
