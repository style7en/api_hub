package gui

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <title>API Hub</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif; background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%); min-height: 100vh; color: #e2e8f0; }
    .wrapper { max-width: 800px; margin: 0 auto; padding: 32px 20px; }
    h1 { font-size: 28px; font-weight: 700; color: #f8fafc; margin-bottom: 24px; letter-spacing: -0.5px; }
    h1 span { color: #22d3ee; }
    .card { background: #1e293b; border: 1px solid #334155; border-radius: 14px; padding: 24px; margin-bottom: 16px; }
    .card-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
    .card-header h2 { font-size: 16px; font-weight: 600; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px; }
    .badge { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; padding: 4px 12px; border-radius: 20px; font-weight: 600; }
    .badge-running { background: #065f46; color: #6ee7b7; }
    .badge-stopped { background: #7f1d1d; color: #fca5a5; }
    .badge-dot { width: 7px; height: 7px; border-radius: 50%; display: inline-block; }
    .badge-running .badge-dot { background: #22c55e; box-shadow: 0 0 6px #22c55e; }
    .badge-stopped .badge-dot { background: #ef4444; }
    .form-group { margin-bottom: 16px; }
    label { display: block; font-size: 13px; font-weight: 600; color: #64748b; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.5px; }
    select { width: 100%; padding: 11px 14px; font-size: 15px; background: #0f172a; color: #e2e8f0; border: 1px solid #334155; border-radius: 10px; appearance: none; cursor: pointer; transition: border-color 0.15s; }
    select:focus { outline: none; border-color: #22d3ee; box-shadow: 0 0 0 3px rgba(34,211,238,0.1); }
    select:disabled { opacity: 0.4; cursor: not-allowed; background: #1a2332; }
    .btn-row { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 18px; }
    button { padding: 10px 22px; font-size: 14px; font-weight: 600; border: none; border-radius: 10px; cursor: pointer; transition: all 0.15s; }
    button:active { transform: scale(0.97); }
    .btn-start { background: #059669; color: white; }
    .btn-start:hover { background: #047857; }
    .btn-stop { background: #dc2626; color: white; }
    .btn-stop:hover { background: #b91c1c; }
    .btn-refresh { background: #334155; color: #cbd5e1; }
    .btn-refresh:hover { background: #475569; }
    .msg { margin-top: 10px; font-size: 13px; color: #22d3ee; min-height: 20px; }
    .client-info { display: none; }
    .client-info.visible { display: block; }
    .client-info h2 { font-size: 16px; font-weight: 600; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; }
    .info-row { display: flex; align-items: center; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #1e293b; }
    .info-row:last-child { border-bottom: none; }
    .info-label { font-size: 13px; color: #64748b; font-weight: 600; text-transform: uppercase; }
    .info-value { font-size: 14px; font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace; color: #22d3ee; word-break: break-all; text-align: right; max-width: 70%; }
    .log-panel { background: #0a0f1a; border-radius: 10px; padding: 14px; min-height: 100px; max-height: 260px; overflow-y: auto; font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace; font-size: 13px; color: #94a3b8; white-space: pre-wrap; line-height: 1.6; }
    .log-panel::-webkit-scrollbar { width: 6px; }
    .log-panel::-webkit-scrollbar-thumb { background: #334155; border-radius: 3px; }
    .footer { text-align: center; margin-top: 24px; font-size: 12px; color: #475569; }
  </style>
</head>
<body>
<div class="wrapper">
  <h1>API <span>Hub</span></h1>
  <div class="card">
    <div class="card-header">
      <h2>API 服务</h2>
      <div id="status-badge" class="badge badge-stopped"><span class="badge-dot"></span>已停止</div>
    </div>
    <div class="form-group">
      <label for="provider">Provider</label>
      <select id="provider"></select>
    </div>
    <div class="form-group">
      <label for="model">Model</label>
      <select id="model"></select>
    </div>
    <div class="btn-row">
      <button class="btn-start" onclick="startService()">启动服务</button>
      <button class="btn-stop" onclick="stopService()">停止服务</button>
      <button class="btn-refresh" onclick="loadState()">刷新</button>
    </div>
    <div class="msg" id="message"></div>
  </div>
  <div class="card client-info" id="client-panel">
    <h2>客户端连接</h2>
    <div class="info-row"><span class="info-label">Base URL</span><span class="info-value" id="info-url">http://127.0.0.1:8080/v1</span></div>
    <div class="info-row"><span class="info-label">API Key</span><span class="info-value" id="info-key">—</span></div>
    <div class="info-row"><span class="info-label">Model</span><span class="info-value" id="info-model">—</span></div>
  </div>
  <div class="card">
    <div class="card-header"><h2>日志</h2></div>
    <div class="log-panel" id="logs"></div>
  </div>
  <div class="footer">API Hub</div>
</div>
<script>
let state = null;
let userSelected = false;

async function request(path, options) {
  const res = await fetch(path, options);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

function render(data) {
  state = data;
  if (!data.service.running && !userSelected) {
    const providerSelect = document.getElementById('provider');
    const modelSelect = document.getElementById('model');
    providerSelect.innerHTML = '';
    Object.keys(data.providers).sort().forEach(name => {
      const option = document.createElement('option');
      option.value = name;
      option.textContent = name;
      providerSelect.appendChild(option);
    });
    providerSelect.value = data.defaults.provider || '';
    renderModels();
    modelSelect.value = data.defaults.model || '';
    providerSelect.disabled = false;
    modelSelect.disabled = false;
  } else if (data.service.running) {
    document.getElementById('provider').disabled = true;
    document.getElementById('model').disabled = true;
  }

  const badge = document.getElementById('status-badge');
  if (data.service.running) {
    badge.innerHTML = '<span class="badge-dot"></span>运行中';
    badge.className = 'badge badge-running';
    document.getElementById('client-panel').classList.add('visible');
    document.getElementById('info-url').textContent = data.client.base_url;
    document.getElementById('info-key').textContent = data.client.api_key;
    document.getElementById('info-model').textContent = data.client.model;
  } else {
    badge.innerHTML = '<span class="badge-dot"></span>已停止';
    badge.className = 'badge badge-stopped';
    document.getElementById('client-panel').classList.remove('visible');
  }
  document.getElementById('logs').textContent = (data.service.logs || []).join('\n');
}

function renderModels() {
  const provider = document.getElementById('provider').value;
  const modelSelect = document.getElementById('model');
  modelSelect.innerHTML = '';
  ((state && state.providers[provider] && state.providers[provider].models) || []).forEach(name => {
    const option = document.createElement('option');
    option.value = name;
    option.textContent = name;
    modelSelect.appendChild(option);
  });
}

async function loadState() {
  try { render(await request('/api/config')); document.getElementById('message').textContent = ''; } catch (err) { document.getElementById('message').textContent = err.message; }
}

async function saveDefaults() {
  if (!state || state.service.running) return;
  try {
    const provider = document.getElementById('provider').value;
    const model = document.getElementById('model').value;
    await request('/api/defaults', {method:'POST', body: JSON.stringify({provider, model})});
    document.getElementById('message').textContent = '已自动保存';
    setTimeout(function(){ document.getElementById('message').textContent = ''; }, 2000);
  } catch (err) { document.getElementById('message').textContent = err.message; }
}

async function startService() {
  try { render(await request('/api/service/start', {method:'POST'})); document.getElementById('message').textContent = '服务已启动'; } catch (err) { document.getElementById('message').textContent = err.message; }
}

async function stopService() {
  userSelected = false;
  try { render(await request('/api/service/stop', {method:'POST'})); document.getElementById('message').textContent = '服务已停止'; } catch (err) { document.getElementById('message').textContent = err.message; }
}

document.getElementById('provider').addEventListener('change', function() {
  userSelected = true;
  renderModels();
  saveDefaults();
});
document.getElementById('model').addEventListener('change', function() {
  userSelected = true;
  saveDefaults();
});
loadState();
setInterval(loadState, 3000);
</script>
</body>
</html>`
