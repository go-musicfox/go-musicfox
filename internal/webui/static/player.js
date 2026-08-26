// musicfox WebUI — player controls (progress / volume / transport / mode / search).
// Loaded after app.js, which owns the WebSocket channel, the shared `state`
// object and the `send` command helper (exposed on window). Everything here is
// plain DOM wiring on the shared player state; render entry points are exposed
// via window.playerUI for app.js to call when snapshots/events arrive.
//
// NOTE: top-level let/const live in the shared global lexical environment, so
// bindings here must not clash with app.js (`shared`/`playerEl` vs `state`/`el`).
"use strict";

const shared = window.state; // shared state owned by app.js
const sendCmd = window.send; // shared command helper owned by app.js

const byId = (id) => document.getElementById(id);

const playerEl = {
  playBtn: byId("btn-play"),
  prevBtn: byId("btn-prev"),
  nextBtn: byId("btn-next"),
  stopBtn: byId("btn-stop"),
  likeBtn: byId("btn-like"),
  repeatBtn: byId("btn-repeat"),
  shuffleBtn: byId("btn-shuffle"),
  progress: byId("progress"),
  progressFill: byId("progress-fill"),
  progressThumb: byId("progress-thumb"),
  timeCur: byId("progress-time-cur"),
  timeTotal: byId("progress-time-total"),
  volume: byId("volume"),
  volumeVal: byId("volume-val"),
  searchInput: byId("search-input"),
  searchBtn: byId("search-btn"),
  playlistInfo: byId("playlist-info"),
};

/* ============================== rendering ============================== */

function renderAll() {
  renderControls();
  renderProgress();
  renderVolume();
  renderModeButtons();
  renderPlaylistInfo();
}

function renderControls() {
  if (shared.state === "playing") {
    playerEl.playBtn.textContent = "暂停";
    playerEl.playBtn.classList.add("active");
  } else {
    playerEl.playBtn.textContent = "播放";
    playerEl.playBtn.classList.remove("active");
  }
}

function renderProgress() {
  const ratio = shared.duration > 0 ? Math.min(1, shared.position / shared.duration) : 0;
  playerEl.progressFill.style.width = ratio * 100 + "%";
  playerEl.progressThumb.style.left = ratio * 100 + "%";
  playerEl.timeCur.textContent = fmtTime(shared.position);
  playerEl.timeTotal.textContent = fmtTime(shared.duration);
}

function renderVolume() {
  const v = Math.max(0, Math.min(100, Math.round(shared.volume)));
  playerEl.volume.value = String(v);
  playerEl.volumeVal.textContent = v + "%";
}

function modeRepeat(mode) {
  if (mode === "单曲循环") return "one";
  if (mode === "列表循环") return "all";
  return "off";
}

function modeShuffle(mode) {
  return mode === "列表随机" || mode === "无限随机" || mode === "心动模式";
}

function renderModeButtons() {
  const rep = modeRepeat(shared.mode);
  playerEl.repeatBtn.textContent = rep === "one" ? "循环:单曲" : rep === "all" ? "循环:列表" : "循环:关";
  playerEl.repeatBtn.classList.toggle("active", rep !== "off");
  const shuf = modeShuffle(shared.mode);
  playerEl.shuffleBtn.textContent = shuf ? "随机:开" : "随机:关";
  playerEl.shuffleBtn.classList.toggle("active", shuf);
}

function renderPlaylistInfo() {
  playerEl.playlistInfo.textContent = "播放列表 " + (shared.playlistLen || 0) + " 首";
}

/* ============================== progress bar ============================== */

function ratioFromEvent(e) {
  const rect = playerEl.progress.getBoundingClientRect();
  if (rect.width <= 0) return 0;
  return Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width));
}

function seekPreviewTick() {
  shared.seekRaf = null;
  const pos = shared.dragRatio * shared.duration;
  const ratio = shared.duration > 0 ? Math.min(1, Math.max(0, pos / shared.duration)) : 0;
  playerEl.progressFill.style.width = ratio * 100 + "%";
  playerEl.progressThumb.style.left = ratio * 100 + "%";
  playerEl.timeCur.textContent = fmtTime(pos);
}

playerEl.progress.addEventListener("pointerdown", (e) => {
  if (shared.duration <= 0) return;
  shared.dragging = true;
  shared.dragRatio = ratioFromEvent(e);
  try {
    playerEl.progress.setPointerCapture(e.pointerId);
  } catch (err) {
    /* noop */
  }
  if (shared.seekRaf) cancelAnimationFrame(shared.seekRaf);
  shared.seekRaf = requestAnimationFrame(seekPreviewTick);
  e.preventDefault();
});

playerEl.progress.addEventListener("pointermove", (e) => {
  if (!shared.dragging) return;
  shared.dragRatio = ratioFromEvent(e);
  if (!shared.seekRaf) shared.seekRaf = requestAnimationFrame(seekPreviewTick);
});

function endSeek(e) {
  if (!shared.dragging) return;
  shared.dragging = false;
  if (shared.seekRaf) {
    cancelAnimationFrame(shared.seekRaf);
    shared.seekRaf = null;
  }
  const pos = shared.dragRatio * shared.duration;
  sendCmd("seek", { seconds: pos });
  shared.position = pos;
  renderProgress();
  try {
    playerEl.progress.releasePointerCapture(e.pointerId);
  } catch (err) {
    /* noop */
  }
}

playerEl.progress.addEventListener("pointerup", endSeek);
playerEl.progress.addEventListener("pointercancel", () => {
  shared.dragging = false;
});

// Keyboard seeking on the focused progress bar (accessibility).
playerEl.progress.addEventListener("keydown", (e) => {
  let pos = null;
  if (e.key === "ArrowLeft") pos = shared.position - 5;
  else if (e.key === "ArrowRight") pos = shared.position + 5;
  else if (e.key === "Home") pos = 0;
  else if (e.key === "End") pos = shared.duration;
  if (pos !== null) {
    e.preventDefault();
    pos = Math.max(0, Math.min(shared.duration, pos));
    sendCmd("seek", { seconds: pos });
    shared.position = pos;
    renderProgress();
  }
});

/* ============================== volume ============================== */

playerEl.volume.addEventListener("input", () => {
  playerEl.volumeVal.textContent = playerEl.volume.value + "%";
});

playerEl.volume.addEventListener("change", () => {
  const v = parseInt(playerEl.volume.value, 10);
  shared.volume = v;
  sendCmd("volume", { value: v });
});

/* ============================== transport / mode buttons ============================== */

playerEl.playBtn.addEventListener("click", () => {
  if (shared.state === "playing") sendCmd("pause");
  else sendCmd("toggle");
});

playerEl.prevBtn.addEventListener("click", () => sendCmd("prev"));
playerEl.nextBtn.addEventListener("click", () => sendCmd("next"));
playerEl.stopBtn.addEventListener("click", () => sendCmd("stop"));
playerEl.likeBtn.addEventListener("click", () => sendCmd("like"));

playerEl.repeatBtn.addEventListener("click", () => {
  const cur = modeRepeat(shared.mode);
  const next = cur === "off" ? "one" : cur === "one" ? "all" : "off";
  sendCmd("repeat", { mode: next });
  shared.mode = next === "one" ? "单曲循环" : next === "all" ? "列表循环" : "顺序播放";
  renderModeButtons();
});

playerEl.shuffleBtn.addEventListener("click", () => {
  const on = !modeShuffle(shared.mode);
  sendCmd("shuffle", { on });
  shared.mode = on ? "列表随机" : "列表循环";
  renderModeButtons();
});

/* ============================== play/search ============================== */

function doSearch() {
  const q = playerEl.searchInput.value.trim();
  if (!q) return;
  sendCmd("play", { query: q }, (msg) => {
    if (msg.ok) {
      toast("正在播放: " + ((msg.data && msg.data.name) || q));
      playerEl.searchInput.value = "";
    }
  });
}

playerEl.searchBtn.addEventListener("click", doSearch);
playerEl.searchInput.addEventListener("keydown", (e) => {
  if (e.key === "Enter") doSearch();
});

/* ============================== export ============================== */

window.playerUI = {
  renderAll,
  renderControls,
  renderProgress,
  renderVolume,
  renderModeButtons,
  renderPlaylistInfo,
};

// Both scripts are loaded (app.js first, then player.js): kick the app off now
// so the initial render + connection start once every DOM ref and API exists.
window.appInit();
