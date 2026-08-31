const $ = selector => document.querySelector(selector);
const servicesRoot = $('#services-list');
const systemRoot = $('#system-stats');
const refreshButton = $('#refresh');
const errorBanner = $('#error-banner');
const toast = $('#toast');
const metricChoices = {cpu:'CPU', memory:'Memory', storage:'Storage', temperature:'Temperature', load:'System load', swap:'Swap'};
const detailChoices = {hostname:'Hostname', local_ip:'Local IP', os:'OS', uptime:'Uptime', kernel:'Kernel', architecture:'Architecture'};
const palette = ['#4ec669','#56b4d3','#8b77cf','#e08b55','#d7667d','#5a6778'];
const bundledIcons = new Set(['docker','jellyfin','plex','qbittorrent','transmission','nextcloud','paperless-ngx','portainer','uptime-kuma','adguard-home','dockge','n8n']);
const iconAliases = {
  'qbit':'qbittorrent', 'paperless':'paperless-ngx', 'paperlessngx':'paperless-ngx',
  'uptimekuma':'uptime-kuma', 'adguard':'adguard-home', 'adguardhome':'adguard-home'
};
let allServices = [];
let config = null;
let configYAML = '';
let lastStats = null;
let activeFilter = 'all';
let refreshTimer;
let toastTimer;
let validationTimer;
let validatedYAML = null;

const escapeHTML = value => String(value ?? '').replace(/[&<>'"]/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'}[c]));
const bytes = value => {
  if (!Number.isFinite(value) || value <= 0) return '0 B';
  const units = ['B','KiB','MiB','GiB','TiB'];
  const i = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** i).toFixed(i > 1 ? 1 : 0)} ${units[i]}`;
};
const percent = value => `${Math.max(0, Number(value) || 0).toFixed(1)}%`;
const isRunning = service => service.status === 'running' || service.status === 'active';
const clamp = value => Math.min(100, Math.max(0, Number(value) || 0));
const clone = value => JSON.parse(JSON.stringify(value));
const iconSVG = name => `<svg aria-hidden="true"><use href="/icons.svg#${name}"></use></svg>`;

servicesRoot.addEventListener('error', event => {
  if (event.target.matches('.custom-icon img')) event.target.parentElement.classList.add('failed');
}, true);

function normalizedIconName(value) {
  return String(value || '').toLowerCase().replace(/[^a-z0-9]+/g, '');
}

function serviceIcon(service) {
  const requested = String(service.icon || 'auto').trim();
  const fallback = String(service.name || service.type || '?').split(/\s+/).map(word => word[0]).join('').slice(0, 2).toUpperCase();
  if (/^https?:\/\//i.test(requested)) {
    return `<span class="service-icon custom-icon"><img src="${escapeHTML(requested)}" alt=""><span>${escapeHTML(fallback)}</span></span>`;
  }
  let candidate = requested === '' || requested.toLowerCase() === 'auto' ? normalizedIconName(`${service.name} ${service.container || ''} ${service.unit || ''}`) : normalizedIconName(requested);
  candidate = iconAliases[candidate] || [...bundledIcons].find(name => candidate.includes(normalizedIconName(name))) || candidate;
  if (bundledIcons.has(candidate)) return `<span class="service-icon">${iconSVG(candidate)}</span>`;
  if ((service.type || '').toLowerCase() === 'docker') return `<span class="service-icon">${iconSVG('docker')}</span>`;
  return `<span class="service-icon fallback-icon">${escapeHTML(fallback)}</span>`;
}

function duration(seconds) {
  seconds = Math.max(0, Math.floor(Number(seconds) || 0));
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor(seconds % 86400 / 3600);
  const minutes = Math.floor(seconds % 3600 / 60);
  return `${days ? `${days}d ` : ''}${hours}h ${minutes}m`;
}

function showToast(message, isError = false) {
  clearTimeout(toastTimer);
  toast.textContent = message;
  toast.className = `toast show${isError ? ' error' : ''}`;
  toastTimer = setTimeout(() => { toast.className = 'toast'; }, 2800);
}

async function request(path, options) {
  const response = await fetch(path, options);
  let body = {};
  try { body = await response.json(); } catch (_) { /* response has no JSON */ }
  if (!response.ok) throw new Error(body.error || `Request failed (${response.status})`);
  return body;
}

function applyAppearance() {
  if (!config) return;
  document.documentElement.style.setProperty('--wallpaper', config.dashboard.background);
  $('meta[name="theme-color"]').content = config.dashboard.background;
  const saved = localStorage.getItem('switchboard-theme');
  let theme = saved || config.dashboard.theme;
  if (theme === 'system') theme = matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  document.documentElement.dataset.theme = theme;
  $('#theme-toggle').textContent = theme === 'dark' ? '☀' : '◐';
}

function metricModel(name, stats) {
  const ram = stats.memoryTotalBytes ? stats.memoryUsedBytes / stats.memoryTotalBytes * 100 : 0;
  const disk = stats.diskTotalBytes ? stats.diskUsedBytes / stats.diskTotalBytes * 100 : 0;
  const swap = stats.swapTotalBytes ? stats.swapUsedBytes / stats.swapTotalBytes * 100 : 0;
  const models = {
    cpu: ['cpu','CPU',percent(stats.cpuPercent),`${stats.cpuCores || '—'} cores`,stats.cpuPercent],
    memory: ['memory','RAM',`${bytes(stats.memoryFreeBytes)} free`,`${bytes(stats.memoryUsedBytes)} used`,ram],
    storage: ['storage','Storage',`${bytes(stats.diskFreeBytes)} free`,`${bytes(stats.diskUsedBytes)} used`,disk],
    temperature: ['temperature','Temperature',stats.temperatureCelsius ? `${stats.temperatureCelsius.toFixed(1)}°C` : 'Unavailable','CPU sensor',stats.temperatureCelsius],
    load: ['load','System load',(Number(stats.loadOne) || 0).toFixed(2),'1 minute average',stats.cpuCores ? stats.loadOne / stats.cpuCores * 100 : 0],
    swap: ['swap','Swap',stats.swapTotalBytes ? `${bytes(stats.swapFreeBytes)} free` : 'Disabled',stats.swapTotalBytes ? `${bytes(stats.swapUsedBytes)} used` : 'No swap configured',swap]
  };
  return models[name];
}

function renderSystem(stats) {
  lastStats = stats;
  const selected = config?.dashboard.overview || ['cpu','memory','storage'];
  systemRoot.innerHTML = selected.map(name => {
    const model = metricModel(name, stats);
    return `<article class="metric"><span class="metric-icon">${iconSVG(model[0])}</span><div class="metric-copy"><div class="metric-line"><strong>${model[1]}</strong><b>${model[2]}</b></div><small>${model[3]}</small><progress class="meter" max="100" value="${clamp(model[4])}"></progress></div></article>`;
  }).join('');
  $('#uptime').textContent = `Uptime: ${duration(stats.uptimeSeconds)}`;
  const details = {
    hostname: stats.hostname || '—', local_ip: stats.localIp || '—', os: stats.os || '—',
    uptime: duration(stats.uptimeSeconds), kernel: stats.kernel || '—', architecture: stats.architecture || '—'
  };
  $('#system-details').innerHTML = (config?.dashboard.systemDetails || ['hostname','local_ip','os']).map(key => `<div><dt>${detailChoices[key]}</dt><dd title="${escapeHTML(details[key])}">${escapeHTML(details[key])}</dd></div>`).join('');
}

function serviceCard(service) {
  const running = isRunning(service);
  const encoded = encodeURIComponent(service.name);
  const link = service.href ? `<a class="service-link" href="${escapeHTML(service.href)}" target="_blank" rel="noopener noreferrer">${escapeHTML(service.name)}</a>` : `<span class="service-link">${escapeHTML(service.name)}</span>`;
  const largestMemory = Math.max(1, ...allServices.map(item => Number(item.memoryBytes) || 0));
  return `<article class="card" data-service="${escapeHTML(service.name)}">
    <div class="card-top">${serviceIcon(service)}<div>${link}<p class="description">${escapeHTML(service.description || service.type)}</p><span class="state-line ${running ? 'running' : ''}"><i></i>${escapeHTML(service.status)}</span></div><input class="switch" type="checkbox" data-action="autostart" data-name="${encoded}" ${service.autostart ? 'checked' : ''} title="Toggle autostart" aria-label="Autostart ${escapeHTML(service.name)}"></div>
    <div class="card-bottom"><div class="card-stats"><span class="mini-stat">CPU <strong>${running ? percent(service.cpuPercent) : '—'}</strong><progress class="mini-meter" max="100" value="${running ? clamp(service.cpuPercent) : 0}"></progress></span><span class="mini-stat">RAM <strong>${running ? bytes(service.memoryBytes) : '—'}</strong><progress class="mini-meter" max="100" value="${running ? clamp((Number(service.memoryBytes) || 0) / largestMemory * 100) : 0}"></progress></span></div><div class="controls"><button class="control" data-action="start" data-name="${encoded}" ${running ? 'disabled' : ''} title="Start">▶</button><button class="control stop" data-action="stop" data-name="${encoded}" ${!running ? 'disabled' : ''} title="Stop">■</button><button class="control" data-action="restart" data-name="${encoded}" ${!running ? 'disabled' : ''} title="Restart">↻</button></div></div>
    ${service.error ? `<div class="service-error" title="${escapeHTML(service.error)}">${escapeHTML(service.error)}</div>` : ''}</article>`;
}

function renderServices(services) {
  allServices = services;
  const visible = activeFilter === 'all' ? services : services.filter(service => service.type === activeFilter);
  const groups = new Map();
  visible.forEach(service => {
    if (!groups.has(service.group)) groups.set(service.group, []);
    groups.get(service.group).push(service);
  });
  if (!visible.length) {
    servicesRoot.innerHTML = `<div class="loading">No ${activeFilter === 'all' ? '' : `${activeFilter} `}services configured.</div>`;
    return;
  }
  servicesRoot.innerHTML = [...groups].map(([group, items]) => `<section class="group"><h3 class="group-heading">${escapeHTML(group)} <span>${items.length}</span></h3><div class="cards">${items.map(serviceCard).join('')}</div></section>`).join('');
}

function setRefreshTimer() {
  clearInterval(refreshTimer);
  refreshTimer = setInterval(loadDashboard, Math.max(5, config?.dashboard.refreshSeconds || 30) * 1000);
}

async function loadDashboard() {
  refreshButton.classList.add('spinning');
  errorBanner.hidden = true;
  const [servicesResult, systemResult] = await Promise.allSettled([request('/api/services'), request('/api/system')]);
  if (servicesResult.status === 'fulfilled') renderServices(servicesResult.value);
  if (systemResult.status === 'fulfilled') renderSystem(systemResult.value);
  const errors = [servicesResult, systemResult].filter(result => result.status === 'rejected').map(result => result.reason.message);
  if (errors.length) { errorBanner.textContent = errors.join(' · '); errorBanner.hidden = false; }
  $('#updated').textContent = `Updated ${new Date().toLocaleTimeString([], {hour:'2-digit',minute:'2-digit'})}`;
  refreshButton.classList.remove('spinning');
}

function renderSettings() {
  if (!config) return;
  $('#setting-listen').value = config.listen;
  $('#setting-refresh').value = config.dashboard.refreshSeconds;
  $('#setting-theme').value = config.dashboard.theme;
  $('#setting-background').value = config.dashboard.background;
  $('#color-palette').innerHTML = palette.map((color, index) => `<button class="color-choice color-${index}${color.toLowerCase() === config.dashboard.background.toLowerCase() ? ' active' : ''}" type="button" data-color="${color}" title="${color}" aria-label="Use ${color}"></button>`).join('');
  $('#overview-options').innerHTML = Object.entries(metricChoices).map(([key,label]) => `<label><input type="checkbox" name="overview" value="${key}" ${config.dashboard.overview.includes(key) ? 'checked' : ''}>${label}</label>`).join('');
  $('#detail-options').innerHTML = Object.entries(detailChoices).map(([key,label]) => `<label><input type="checkbox" name="detail" value="${key}" ${config.dashboard.systemDetails.includes(key) ? 'checked' : ''}>${label}</label>`).join('');
  $('#configured-services').innerHTML = config.services.length ? config.services.map((service,index) => `<div class="configured-row"><strong>${escapeHTML(service.name)}</strong><span>${escapeHTML(service.type)}</span><span>${escapeHTML(service.container || service.unit)}</span><div class="row-actions"><button type="button" data-edit-service="${index}">Edit</button><button class="danger" type="button" data-delete-service="${index}">Delete</button></div></div>`).join('') : '<div class="loading">No services configured.</div>';
  $('#yaml-editor').value = configYAML;
  validateYAML();
}

async function saveConfiguration(nextConfig) {
  const result = await request('/api/config', {method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({config:nextConfig})});
  config = result.config;
  configYAML = result.yaml;
  applyAppearance();
  renderSettings();
  setRefreshTimer();
  await loadDashboard();
  return result;
}

function openServiceDialog(index = -1) {
  const service = index >= 0 ? config.services[index] : {name:'',type:'docker',container:'',href:'',description:'',group:'Other'};
  $('#service-index').value = index;
  $('#service-dialog-title').textContent = index >= 0 ? 'Edit service' : 'Add service';
  $('#service-name').value = service.name || '';
  $('#service-icon').value = service.icon || 'auto';
  $('#service-type').value = service.type || 'docker';
  $('#service-target').value = service.container || service.unit || '';
  $('#service-group').value = service.group || 'Other';
  $('#service-href').value = service.href || '';
  $('#service-description').value = service.description || '';
  $('#service-form-error').hidden = true;
  updateTargetLabel();
  $('#service-dialog').showModal();
}

function updateTargetLabel() {
  const docker = $('#service-type').value === 'docker';
  $('#target-label').firstChild.textContent = docker ? 'Container name ' : 'systemd unit ';
}

async function validateYAML() {
  clearTimeout(validationTimer);
  const status = $('#yaml-status');
  status.className = 'validation-status';
  status.textContent = 'Checking YAML…';
  $('#save-yaml').disabled = true;
  validationTimer = setTimeout(async () => {
    try {
      validatedYAML = await request('/api/config/validate', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({yaml:$('#yaml-editor').value})});
      status.className = 'validation-status valid';
      status.textContent = 'Valid YAML — safe to save.';
      $('#save-yaml').disabled = false;
    } catch (error) {
      validatedYAML = null;
      status.className = 'validation-status invalid';
      status.textContent = error.message;
    }
  }, 350);
}

servicesRoot.addEventListener('click', async event => {
  const button = event.target.closest('button[data-action]');
  if (!button) return;
  const card = button.closest('.card');
  card.classList.add('busy');
  try {
    await request(`/api/services/${button.dataset.name}/${button.dataset.action}`, {method:'POST'});
    showToast(`${decodeURIComponent(button.dataset.name)} ${button.dataset.action} requested`);
    await loadDashboard();
  } catch (error) { showToast(error.message, true); card.classList.remove('busy'); }
});

servicesRoot.addEventListener('change', async event => {
  const input = event.target.closest('input[data-action="autostart"]');
  if (!input) return;
  const card = input.closest('.card');
  card.classList.add('busy');
  try {
    await request(`/api/services/${input.dataset.name}/autostart`, {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({enabled:input.checked})});
    showToast(`Autostart ${input.checked ? 'enabled' : 'disabled'}`);
    await loadDashboard();
  } catch (error) { input.checked = !input.checked; showToast(error.message, true); card.classList.remove('busy'); }
});

document.querySelectorAll('[data-view]').forEach(button => button.addEventListener('click', () => {
  const view = button.dataset.view;
  if (button.dataset.filter) activeFilter = button.dataset.filter;
  else if (view === 'overview') activeFilter = 'all';
  document.querySelectorAll('.nav-item').forEach(item => item.classList.remove('active'));
  button.classList.add('active');
  $('#overview-view').hidden = view !== 'overview';
  $('#settings-view').hidden = view !== 'settings';
  if (view === 'overview') renderServices(allServices); else renderSettings();
}));

document.querySelectorAll('[data-settings-tab]').forEach(button => button.addEventListener('click', () => {
  document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
  button.classList.add('active');
  const yaml = button.dataset.settingsTab === 'yaml';
  $('#settings-form').hidden = yaml;
  $('#yaml-panel').hidden = !yaml;
  if (yaml) { $('#yaml-editor').value = configYAML; validateYAML(); }
}));

$('#settings-form').addEventListener('submit', async event => {
  event.preventDefault();
  const next = clone(config);
  next.listen = $('#setting-listen').value.trim();
  next.dashboard.refreshSeconds = Number($('#setting-refresh').value);
  next.dashboard.theme = $('#setting-theme').value;
  next.dashboard.background = $('#setting-background').value;
  next.dashboard.overview = [...document.querySelectorAll('input[name="overview"]:checked')].map(input => input.value);
  next.dashboard.systemDetails = [...document.querySelectorAll('input[name="detail"]:checked')].map(input => input.value);
  try { await saveConfiguration(next); showToast('Settings saved'); } catch (error) { showToast(error.message, true); }
});

$('#color-palette').addEventListener('click', event => {
  const button = event.target.closest('[data-color]');
  if (!button) return;
  $('#setting-background').value = button.dataset.color;
  document.querySelectorAll('.color-choice').forEach(item => item.classList.toggle('active', item === button));
});

$('#configured-services').addEventListener('click', async event => {
  const edit = event.target.closest('[data-edit-service]');
  if (edit) { openServiceDialog(Number(edit.dataset.editService)); return; }
  const remove = event.target.closest('[data-delete-service]');
  if (!remove) return;
  const index = Number(remove.dataset.deleteService);
  if (!confirm(`Remove ${config.services[index].name} from Switchboard? This does not stop or uninstall it.`)) return;
  const next = clone(config);
  next.services.splice(index, 1);
  try { await saveConfiguration(next); showToast('Service removed'); } catch (error) { showToast(error.message, true); }
});

$('#service-type').addEventListener('change', updateTargetLabel);
$('#add-service').addEventListener('click', () => openServiceDialog());
$('#settings-add-service').addEventListener('click', () => openServiceDialog());
$('#service-form').addEventListener('submit', async event => {
  event.preventDefault();
  if (event.submitter?.value === 'cancel') { $('#service-dialog').close(); return; }
  if (!event.currentTarget.reportValidity()) return;
  const type = $('#service-type').value;
  const service = {name:$('#service-name').value.trim(),icon:$('#service-icon').value.trim() || 'auto',type,href:$('#service-href').value.trim(),description:$('#service-description').value.trim(),group:$('#service-group').value.trim()};
  service[type === 'docker' ? 'container' : 'unit'] = $('#service-target').value.trim();
  const index = Number($('#service-index').value);
  const next = clone(config);
  if (index >= 0) next.services[index] = service; else next.services.push(service);
  try {
    await saveConfiguration(next);
    $('#service-dialog').close();
    showToast(index >= 0 ? 'Service updated' : 'Service added');
  } catch (error) { $('#service-form-error').textContent = error.message; $('#service-form-error').hidden = false; }
});

$('#yaml-editor').addEventListener('input', validateYAML);
$('#format-yaml').addEventListener('click', () => {
  if (!validatedYAML) return;
  $('#yaml-editor').value = validatedYAML.yaml;
  validateYAML();
});
$('#save-yaml').addEventListener('click', async () => {
  if (!validatedYAML) return;
  try {
    const result = await request('/api/config', {method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({yaml:$('#yaml-editor').value})});
    config = result.config; configYAML = result.yaml;
    applyAppearance(); renderSettings(); setRefreshTimer(); await loadDashboard();
    showToast('YAML saved');
  } catch (error) { showToast(error.message, true); validateYAML(); }
});

refreshButton.addEventListener('click', loadDashboard);
$('#theme-toggle').addEventListener('click', () => {
  const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
  localStorage.setItem('switchboard-theme', next);
  document.documentElement.dataset.theme = next;
  $('#theme-toggle').textContent = next === 'dark' ? '☀' : '◐';
});

async function initialize() {
  try {
    const result = await request('/api/config');
    config = result.config;
    configYAML = result.yaml;
    applyAppearance();
    renderSettings();
    setRefreshTimer();
    await loadDashboard();
  } catch (error) {
    errorBanner.textContent = error.message;
    errorBanner.hidden = false;
    showToast(error.message, true);
  }
}

initialize();
