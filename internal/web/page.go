package web

// indexHTML is the read-only dashboard. It polls /api/state and renders a
// terminal-style view of the current track and the queue.
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>music-kwewe-bot</title>
<style>
  :root {
    --bg: #0a0e14;
    --fg: #c6ffd6;
    --dim: #4b6e57;
    --accent: #00ff9c;
    --amber: #ffcb6b;
    --red: #ff6b6b;
    --border: #1c2b22;
  }
  * { box-sizing: border-box; }
  html, body {
    margin: 0; height: 100%;
    background: var(--bg);
    color: var(--fg);
    font-family: ui-monospace, "SF Mono", SFMono-Regular, Menlo, Monaco,
                 "Cascadia Code", "Roboto Mono", "Liberation Mono", monospace;
    font-size: 14px; line-height: 1.55;
  }
  body {
    display: flex; justify-content: center;
    padding: 28px 16px;
  }
  .term {
    width: 100%; max-width: 820px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: linear-gradient(180deg, #0c1119, #0a0e14);
    box-shadow: 0 0 0 1px #000, 0 20px 60px rgba(0,0,0,.6);
    overflow: hidden;
  }
  .bar {
    display: flex; align-items: center; gap: 8px;
    padding: 10px 14px;
    background: #0e141d;
    border-bottom: 1px solid var(--border);
  }
  .dot { width: 11px; height: 11px; border-radius: 50%; display: inline-block; }
  .dot.r { background: #ff5f56; } .dot.y { background: #ffbd2e; } .dot.g { background: #27c93f; }
  .bar .title { margin-left: 8px; color: var(--dim); font-size: 12px; }
  .screen { padding: 18px 20px 22px; }
  .prompt { color: var(--dim); margin: 0 0 6px; }
  .prompt b { color: var(--accent); font-weight: 600; }
  .section { margin: 18px 0 0; }
  h2 {
    font-size: 12px; letter-spacing: .12em; text-transform: uppercase;
    color: var(--dim); margin: 0 0 8px; font-weight: 600;
  }
  .now {
    border-left: 2px solid var(--accent);
    padding: 6px 0 6px 12px;
  }
  .now .t { color: var(--accent); }
  .now .meta, .row .meta { color: var(--dim); }
  .idle { color: var(--dim); }
  ol.queue { list-style: none; margin: 0; padding: 0; }
  ol.queue li {
    display: flex; gap: 12px; padding: 3px 0;
    border-bottom: 1px dashed var(--border);
  }
  ol.queue li:last-child { border-bottom: 0; }
  .idx { color: var(--amber); min-width: 2.2em; text-align: right; }
  .row .t { color: var(--fg); }
  a { color: inherit; text-decoration: none; }
  a:hover { text-decoration: underline; }
  .foot {
    margin-top: 20px; padding-top: 10px;
    border-top: 1px solid var(--border);
    color: var(--dim); font-size: 12px;
    display: flex; justify-content: space-between; gap: 12px; flex-wrap: wrap;
  }
  .cursor {
    display: inline-block; width: 8px; height: 1.05em;
    background: var(--accent); vertical-align: text-bottom;
    margin-left: 2px; animation: blink 1s steps(1) infinite;
  }
  @keyframes blink { 50% { opacity: 0; } }
  .err { color: var(--red); }
</style>
</head>
<body>
  <div class="term">
    <div class="bar">
      <span class="dot r"></span><span class="dot y"></span><span class="dot g"></span>
      <span class="title">music-kwewe-bot — dashboard (read-only)</span>
    </div>
    <div class="screen">
      <p class="prompt">music-kwewe-bot ~ $ <b>status</b><span class="cursor"></span></p>

      <div class="section">
        <h2>&#9654; now playing</h2>
        <div id="now"><span class="idle">loading...</span></div>
      </div>

      <div class="section">
        <h2># queue</h2>
        <ol class="queue" id="queue"></ol>
        <div id="queue-empty" class="idle" style="display:none">queue is empty</div>
      </div>

      <div class="foot">
        <span id="count">tracks: 0</span>
        <span id="updated">updated: --:--:--</span>
      </div>
    </div>
  </div>

<script>
(function () {
  var nowEl = document.getElementById("now");
  var queueEl = document.getElementById("queue");
  var emptyEl = document.getElementById("queue-empty");
  var countEl = document.getElementById("count");
  var updatedEl = document.getElementById("updated");

  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text != null) e.textContent = text;
    return e;
  }

  function linkOrText(track, cls) {
    if (track.url) {
      var a = el("a", cls, track.title);
      a.href = track.url;
      a.target = "_blank";
      a.rel = "noopener";
      return a;
    }
    return el("span", cls, track.title);
  }

  function renderNow(np) {
    nowEl.innerHTML = "";
    if (!np) {
      nowEl.appendChild(el("span", "idle", "nothing playing"));
      return;
    }
    var box = el("div", "now");
    box.appendChild(linkOrText(np, "t"));
    if (np.added_by) box.appendChild(el("div", "meta", "added by " + np.added_by));
    nowEl.appendChild(box);
  }

  function renderQueue(items) {
    queueEl.innerHTML = "";
    if (!items || items.length === 0) {
      emptyEl.style.display = "block";
      return;
    }
    emptyEl.style.display = "none";
    for (var i = 0; i < items.length; i++) {
      var li = el("li", "row");
      li.appendChild(el("span", "idx", (i + 1) + "."));
      var body = el("span");
      body.appendChild(linkOrText(items[i], "t"));
      if (items[i].added_by) {
        body.appendChild(el("span", "meta", "  " + String.fromCharCode(8212) + " " + items[i].added_by));
      }
      li.appendChild(body);
      queueEl.appendChild(li);
    }
  }

  function pad(n) { return (n < 10 ? "0" : "") + n; }
  function stamp() {
    var d = new Date();
    return pad(d.getHours()) + ":" + pad(d.getMinutes()) + ":" + pad(d.getSeconds());
  }

  function refresh() {
    fetch("/api/state", { cache: "no-store" })
      .then(function (r) { return r.json(); })
      .then(function (s) {
        renderNow(s.now_playing);
        renderQueue(s.queue);
        var total = (s.queue ? s.queue.length : 0) + (s.now_playing ? 1 : 0);
        countEl.textContent = "tracks: " + total;
        countEl.className = "";
        updatedEl.textContent = "updated: " + stamp();
      })
      .catch(function () {
        countEl.textContent = "connection lost";
        countEl.className = "err";
      });
  }

  refresh();
  setInterval(refresh, 2000);
})();
</script>
</body>
</html>`
