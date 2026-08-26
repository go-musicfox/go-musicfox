// musicfox WebUI frontend
// Vanilla JS, no build step, no external libraries. Talks to the WebUI
// backend exclusively over same-origin WebSocket (/ws) and HTTP (/api/*).
//
// Splitting: this file owns the connection, frame dispatch, snapshot/event
// rendering, cover art, lyrics and QR login. Player controls (progress bar,
// volume, transport, mode buttons, search) live in player.js and consume the
// shared state + send helper exported below.
"use strict";

/* ============================== helpers ============================== */

const $ = (id) => document.getElementById(id);

function fmtTime(sec) {
  if (!isFinite(sec) || sec < 0) sec = 0;
  sec = Math.floor(sec);
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return m + ":" + String(s).padStart(2, "0");
}

// Same-origin JSON GET. The backend contract mixes two payload shapes
// ({"ok":true,"data":{...}} and bare {"fragments":...}), so callers decide
// whether to unwrap ".data" — this helper only returns the parsed body.
async function fetchJSON(url) {
  const resp = await fetch(url, { credentials: "same-origin" });
  if (!resp.ok) throw new Error("HTTP " + resp.status);
  return resp.json();
}

/* ============================== shared state (also used by player.js) ============================== */

const state = {
  ws: null,
  seq: 0,
  connected: false,
  reconnectTimer: null,
  // Player fields consumed by player.js:
  songId: 0,
  mode: "顺序播放",
  state: "stopped",
  position: 0,
  duration: 0,
  volume: 100,
  user: "",
  playlistLen: 0,
  dragging: false,
  dragRatio: 0,
  seekRaf: null,
  // App fields:
  lyrics: { fragments: [], offsetMs: 0 },
  lyricIndex: -1,
  qrKey: "",
  qrTimer: null,
  qrModalOpen: false,
};

// Exports for player.js (loaded after this file).
window.state = state;
window.appInit = init;

/* ============================== DOM refs ============================== */

const el = {
  connDot: $("conn-dot"),
  connText: $("conn-text"),
  user: $("user"),
  loginBtn: $("login-btn"),
  cover: $("cover"),
  songName: $("song-name"),
  artist: $("artist"),
  album: $("album"),
  lyrics: $("lyrics"),
  lyricsEmpty: $("lyrics-empty"),
  statusHint: $("status-hint"),
  modal: $("login-modal"),
  modalClose: $("login-modal-close"),
  qrImg: $("login-qr"),
  qrStatus: $("login-qr-status"),
};

/* ============================== toast ============================== */

let hintTimer = null;

function toast(msg) {
  el.statusHint.textContent = msg;
  if (hintTimer) clearTimeout(hintTimer);
  hintTimer = setTimeout(() => {
    el.statusHint.textContent = "";
  }, 4000);
}

/* ============================== command channel ============================== */

const pending = new Map();

// send dispatches one command over the WS (or toasts "未连接" when the socket is
// down). player.js and this file share it via window.send.
function send(cmd, args, onResp) {
  const id = ++state.seq;
  const payload = { v: 1, id, cmd };
  if (args) payload.args = args;
  if (onResp) pending.set(id, onResp);
  if (state.ws && state.ws.readyState === WebSocket.OPEN) {
    try {
      state.ws.send(JSON.stringify(payload));
    } catch (e) {
      pending.delete(id);
      toast("发送失败");
    }
  } else {
    pending.delete(id);
    toast("未连接");
  }
  return id;
}

window.send = send;

function handleResponse(msg) {
  const cb = pending.get(msg.id);
  if (cb) {
    pending.delete(msg.id);
    cb(msg);
  }
  if (!msg.ok && msg.error) toast(msg.error);
}

/* ============================== WebSocket ============================== */

function wsURL() {
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  return proto + "//" + location.host + "/ws";
}

function connectWS() {
  const ws = new WebSocket(wsURL());
  state.ws = ws;
  ws.onopen = onWSOpen;
  ws.onmessage = onWSMessage;
  ws.onclose = onWSClose;
  ws.onerror = () => {
    try {
      ws.close();
    } catch (e) {
      /* noop */
    }
  };
}

function onWSOpen() {
  state.connected = true;
  setConnUI(true);
  toast("");
  // Force a full snapshot after (re)connect so the page never shows stale UI.
  send("status", null, (msg) => {
    if (msg.ok && msg.data) renderSnapshot(msg.data);
  });
  refreshCover();
  refreshLyrics();
}

function onWSClose() {
  state.connected = false;
  setConnUI(false);
  if (state.reconnectTimer) clearTimeout(state.reconnectTimer);
  state.reconnectTimer = setTimeout(connectWS, 5000);
}

function onWSMessage(ev) {
  let msg;
  try {
    msg = JSON.parse(ev.data);
  } catch (e) {
    return;
  }
  if (!msg || typeof msg !== "object") return;
  if (msg.type === "snapshot") {
    renderSnapshot(msg.data);
    return;
  }
  if (msg.type === "event") {
    handleEvent(msg.event, msg.data);
    return;
  }
  if (msg.id !== undefined) handleResponse(msg);
}

/* ============================== event frames ============================== */

function handleEvent(name, data) {
  switch (name) {
    case "song_changed":
      if (data) {
        // A live song switch starts from the top of the track.
        if (applySong(data)) {
          state.position = 0;
          window.playerUI.renderProgress();
        }
      }
      refreshCover();
      refreshLyrics();
      break;
    case "state_changed":
      if (data && data.state) {
        state.state = data.state;
        window.playerUI.renderControls();
      }
      break;
    case "position":
      if (data && isFinite(data.positionSeconds)) {
        state.position = data.positionSeconds;
        if (!state.dragging) window.playerUI.renderProgress();
        highlightLyricAt(state.position * 1000);
      }
      break;
    case "startup_phase":
      toast(
        "启动中: " +
          (typeof data === "string" ? data : data && data.phase ? data.phase : ""),
      );
      break;
    case "login":
      closeLoginModal();
      send("status", null, (msg) => {
        if (msg.ok && msg.data) renderSnapshot(msg.data);
      });
      toast("登录成功");
      break;
  }
}

/* ============================== snapshot / status rendering ============================== */

function setConnUI(ok) {
  el.connDot.classList.toggle("off", !ok);
  el.connText.textContent = ok ? "已连接" : "未连接";
}

function renderSnapshot(d) {
  if (!d) return;
  if (d.state) state.state = d.state;
  if (isFinite(d.positionSeconds)) state.position = d.positionSeconds;
  if (isFinite(d.durationSeconds) && d.durationSeconds > 0) state.duration = d.durationSeconds;
  if (typeof d.volume === "number") state.volume = d.volume;
  if (typeof d.mode === "string") state.mode = d.mode;
  if (typeof d.user === "string") state.user = d.user;
  state.playlistLen =
    typeof d.playlistLen === "number"
      ? d.playlistLen
      : Array.isArray(d.playlist)
        ? d.playlist.length
        : state.playlistLen;
  if (d.song) applySong(d.song, d.durationSeconds);
  window.playerUI.renderAll();
  renderUser();
}

function applySong(song, topDuration) {
  if (!song || typeof song !== "object") return false;
  const id = Number(song.id) || 0;
  const changed = id !== state.songId;
  state.songId = id;
  if (isFinite(topDuration) && topDuration > 0) state.duration = topDuration;
  if (isFinite(song.durationSeconds) && song.durationSeconds > 0) state.duration = song.durationSeconds;
  el.songName.textContent = song.name || "未在播放";
  el.artist.textContent = song.artist || "";
  el.album.textContent = song.album || "";
  return changed;
}

function renderUser() {
  if (state.user) {
    el.user.textContent = state.user;
    el.loginBtn.textContent = "重新登录";
  } else {
    el.user.textContent = "未登录";
    el.loginBtn.textContent = "登录";
  }
}

/* ============================== cover art ============================== */

let coverToken = 0;

async function refreshCover() {
  const token = ++coverToken;
  try {
    const resp = await fetch("/api/albumart?size=300", { credentials: "same-origin" });
    if (!resp.ok) throw new Error("HTTP " + resp.status);
    const blob = await resp.blob();
    if (token !== coverToken) return; // superseded by a newer refresh
    if (el.cover.dataset.url) URL.revokeObjectURL(el.cover.dataset.url);
    const url = URL.createObjectURL(blob);
    el.cover.dataset.url = url;
    el.cover.src = url;
    el.cover.style.display = "block";
  } catch (e) {
    if (token === coverToken) el.cover.style.display = "none";
  }
}

/* ============================== lyrics ============================== */

async function refreshLyrics() {
  try {
    const body = await fetchJSON("/api/lyrics");
    const data = body.data || body;
    const frags = Array.isArray(data.fragments) ? data.fragments : [];
    state.lyrics = { fragments: frags, offsetMs: Number(data.offsetMs) || 0 };
    renderLyrics();
    highlightLyricAt(state.position * 1000);
  } catch (e) {
    state.lyrics = { fragments: [], offsetMs: 0 };
    renderLyrics();
  }
}

function fragTime(f) {
  const t =
    f && f.startTimeMs !== undefined
      ? f.startTimeMs
      : f && f.timeMs !== undefined
        ? f.timeMs
        : f && f.time !== undefined
          ? f.time
          : 0;
  return Number(t) || 0;
}

function renderLyrics() {
  const frags = state.lyrics.fragments;
  el.lyrics.replaceChildren();
  state.lyricIndex = -1;
  if (!frags.length) {
    el.lyricsEmpty.style.display = "block";
    return;
  }
  el.lyricsEmpty.style.display = "none";
  const nodes = frags.map((f, i) => {
    const div = document.createElement("div");
    div.className = "lyric-line";
    div.textContent = (f && (f.content || f.text)) || "";
    return div;
  });
  el.lyrics.replaceChildren(...nodes);
}

function highlightLyricAt(posMs) {
  const frags = state.lyrics.fragments;
  if (!frags.length) return;
  const eff = posMs + state.lyrics.offsetMs;
  let idx = -1;
  for (let i = 0; i < frags.length; i++) {
    if (fragTime(frags[i]) <= eff) idx = i;
    else break;
  }
  if (idx === state.lyricIndex) return;
  state.lyricIndex = idx;
  const lines = el.lyrics.children;
  for (let i = 0; i < lines.length; i++) lines[i].classList.toggle("active", i === idx);
  if (idx >= 0) {
    const line = lines[idx];
    const target = line.offsetTop - el.lyrics.clientHeight / 2 + line.clientHeight / 2;
    el.lyrics.scrollTo({ top: Math.max(0, target), behavior: "smooth" });
  }
}

/* ============================== login (QR) ============================== */

function openLoginModal() {
  if (state.qrModalOpen) return;
  state.qrModalOpen = true;
  el.modal.hidden = false;
  el.qrStatus.textContent = "获取二维码中…";
  el.qrImg.removeAttribute("src");
  requestQRKey();
}

function closeLoginModal() {
  if (!state.qrModalOpen) return;
  state.qrModalOpen = false;
  state.qrKey = "";
  stopQRPoll();
  el.modal.hidden = true;
}

function stopQRPoll() {
  if (state.qrTimer) {
    clearInterval(state.qrTimer);
    state.qrTimer = null;
  }
}

async function requestQRKey() {
  stopQRPoll();
  try {
    const body = await fetchJSON("/api/login/qr/key");
    const data = body.data || body;
    const key = data.uniKey || data.unikey || data.key;
    if (!key) throw new Error("no uniKey");
    state.qrKey = key;
    el.qrImg.src = "/api/login/qr/image?key=" + encodeURIComponent(key);
    el.qrStatus.textContent = "请使用网易云音乐 App 扫码登录";
    state.qrTimer = setInterval(pollQRStatus, 2000);
  } catch (e) {
    el.qrStatus.textContent = "二维码获取失败（后端可能未就绪）";
  }
}

async function pollQRStatus() {
  if (!state.qrKey || !state.qrModalOpen) return;
  try {
    const body = await fetchJSON("/api/login/qr/status?key=" + encodeURIComponent(state.qrKey));
    const data = body.data || body;
    const code = Number(data.code);
    if (code === 801) {
      el.qrStatus.textContent = "等待扫码…";
    } else if (code === 802) {
      el.qrStatus.textContent = "已扫描，请在手机上确认登录";
    } else if (code === 803) {
      el.qrStatus.textContent = "登录成功！";
      stopQRPoll();
      setTimeout(closeLoginModal, 800);
      send("status", null, (msg) => {
        if (msg.ok && msg.data) renderSnapshot(msg.data);
      });
    } else if (code === 800) {
      el.qrStatus.textContent = "二维码已过期，正在刷新…";
      requestQRKey();
    } else {
      el.qrStatus.textContent = "状态码 " + code;
    }
  } catch (e) {
    el.qrStatus.textContent = "状态查询失败";
  }
}

el.loginBtn.addEventListener("click", openLoginModal);
el.modalClose.addEventListener("click", closeLoginModal);
el.modal.addEventListener("click", (e) => {
  if (e.target === el.modal) closeLoginModal();
});

/* ============================== init ============================== */

async function refreshStatusHTTP() {
  // Fallback snapshot pull over HTTP for environments where the WS snapshot
  // broadcast is not delivered on connect.
  try {
    const body = await fetchJSON("/api/status");
    const data = body.data || body;
    if (data) renderSnapshot(data);
  } catch (e) {
    /* the WS channel covers the snapshot when /api is unavailable */
  }
}

// Called by player.js once both scripts are loaded and every DOM ref exists.
function init() {
  setConnUI(false);
  window.playerUI.renderAll();
  renderUser();
  connectWS();
  refreshStatusHTTP();
}
