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
  .stats { display: grid; grid-template-columns: 1fr 1fr; gap: 18px 28px; }
  @media (max-width: 560px) { .stats { grid-template-columns: 1fr; } }
  ul.chart { list-style: none; margin: 0; padding: 0; }
  ul.chart li {
    display: grid; grid-template-columns: 1.4em 1fr auto;
    align-items: baseline; gap: 8px; padding: 2px 0;
  }
  ul.chart .rank { color: var(--amber); text-align: right; }
  ul.chart .who { color: var(--fg); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  ul.chart .bar { color: var(--accent); letter-spacing: -1px; }
  ul.chart .n { color: var(--dim); }
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
  .invite {
    position: fixed; top: 28px; right: 16px; z-index: 10;
    display: flex; flex-direction: column; align-items: center; gap: 6px;
    width: 135px; padding: 11px;
    background: #0c1119; border: 1px solid var(--border); border-radius: 8px;
    box-shadow: 0 8px 24px rgba(0,0,0,.5);
    opacity: .55; transition: opacity .2s ease;
  }
  .invite:hover { opacity: 1; }
  .qr-img {
    width: 100px; height: 100px; flex-shrink: 0; display: none;
    image-rendering: pixelated;
    background: #fff; padding: 0; border-radius: 3px;
  }
  .invite-meta { font-size: 12px; color: var(--dim); line-height: 1.5; text-align: center; }
  .invite-meta .k { display: block; text-transform: uppercase; letter-spacing: .1em; }
  .inv-pass { color: var(--accent); word-break: break-word; }
  .invite-meta .hint { color: var(--dim); }
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

      <div class="section">
        <h2>~/.stats &#8212; session</h2>
        <div class="stats">
          <div>
            <h2>&#9733; top djs</h2>
            <ul class="chart" id="djs"></ul>
            <div id="djs-empty" class="idle">no plays yet</div>
          </div>
          <div>
            <h2>&#9836; session artists</h2>
            <ul class="chart" id="artists"></ul>
            <div id="artists-empty" class="idle">no plays yet</div>
          </div>
        </div>
      </div>

      <div class="foot">
        <span id="count">tracks: 0</span>
        <span id="played">played: 0</span>
        <span id="updated">updated: --:--:--</span>
      </div>
    </div>
  </div>

  <div class="invite" id="invite-section" style="display:none">
    <a id="qr-link" href="#" target="_blank" rel="noopener" title="Open bot in Telegram">
      <img id="qr-img" class="qr-img" src="/qr" alt="Scan to join">
    </a>
    <div class="invite-meta">
      <div><span class="k">pass</span><span id="inv-pass" class="inv-pass"></span></div>
    </div>
  </div>

<script>
(function () {
  var nowEl = document.getElementById("now");
  var queueEl = document.getElementById("queue");
  var emptyEl = document.getElementById("queue-empty");
  var countEl = document.getElementById("count");
  var playedEl = document.getElementById("played");
  var updatedEl = document.getElementById("updated");
  var djsEl = document.getElementById("djs");
  var djsEmptyEl = document.getElementById("djs-empty");
  var artistsEl = document.getElementById("artists");
  var artistsEmptyEl = document.getElementById("artists-empty");

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
    if (np.elapsed || np.duration) {
      var prog = fmtClock(np.elapsed || 0);
      if (np.duration) prog += " / " + fmtClock(np.duration);
      if (np.paused) prog = "⏸ paused — " + prog;
      box.appendChild(el("div", "meta", prog));
    }
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

  function renderChart(listEl, emptyEl, rows) {
    listEl.innerHTML = "";
    if (!rows || rows.length === 0) {
      emptyEl.style.display = "block";
      return;
    }
    emptyEl.style.display = "none";
    var max = 0;
    for (var i = 0; i < rows.length; i++) { if (rows[i].count > max) max = rows[i].count; }
    for (var j = 0; j < rows.length; j++) {
      var row = rows[j];
      var li = el("li");
      li.appendChild(el("span", "rank", "#" + (j + 1)));
      var who = el("span", "who", row.name);
      who.title = row.name;
      li.appendChild(who);
      var width = max > 0 ? Math.round((row.count / max) * 10) : 0;
      var bar = max > 0 && width < 1 ? 1 : width;
      var meter = el("span", "bar", repeat(String.fromCharCode(9608), bar));
      meter.appendChild(el("span", "n", " " + row.count));
      li.appendChild(meter);
      listEl.appendChild(li);
    }
  }

  function repeat(ch, n) {
    var s = "";
    for (var i = 0; i < n; i++) s += ch;
    return s;
  }

  function pad(n) { return (n < 10 ? "0" : "") + n; }
  function fmtClock(s) {
    s = Math.max(0, Math.floor(s));
    var h = Math.floor(s / 3600), m = Math.floor(s / 60) % 60, r = pad(s % 60);
    return h > 0 ? h + ":" + pad(m) + ":" + r : m + ":" + r;
  }
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
        renderChart(djsEl, djsEmptyEl, s.contributors);
        renderChart(artistsEl, artistsEmptyEl, s.artists);
        var total = (s.queue ? s.queue.length : 0) + (s.now_playing ? 1 : 0);
        countEl.textContent = "tracks: " + total;
        countEl.className = "";
        playedEl.textContent = "played: " + (s.played || 0);
        updatedEl.textContent = "updated: " + stamp();
      })
      .catch(function () {
        countEl.textContent = "connection lost";
        countEl.className = "err";
      });
  }

  refresh();
  setInterval(refresh, 1000);

  fetch("/api/invite", {cache: "no-store"})
    .then(function(r) { return r.json(); })
    .then(function(inv) {
      if (!inv.passphrase && !inv.bot_link) return;
      document.getElementById("invite-section").style.display = "";
      if (inv.passphrase) {
        document.getElementById("inv-pass").textContent = inv.passphrase;
      }
      if (inv.bot_link) {
        document.getElementById("qr-link").href = inv.bot_link;
        var img = document.getElementById("qr-img");
        img.style.display = "block";
        img.onerror = function() { img.style.display = "none"; };
      }
    })
    .catch(function() {});
})();
</script>
</body>
</html>`
