import { invoke, isTauri } from '@tauri-apps/api/core';
import { createIcons, Play, Square, Save, ExternalLink, Eye, EyeOff, Monitor, Server, Activity } from 'lucide';
import './style.css';

type Settings = { id: string; server: string; secret: string; target: string; tls: boolean; dashboard_port: number };
type Status = { running: boolean; ready: boolean; log: string; version: string };
const languages = {
  zh: { brand: '小鸡远程', workstation: '连接工作台', connection: '连接设置', relay: '中继服务器', device: '本机 ID', target: '目标设备 ID（可留空）', secret: '共享密钥', tls: 'TLS 加密连接', port: '本地面板端口', save: '保存配置', start: '启动客户端', stop: '停止客户端', dashboard: '打开会话面板', logs: '运行日志', idle: '未运行', starting: '正在连接', ready: '客户端运行中', saved: '配置已保存', empty: '暂无日志', preview: '浏览器预览：原生连接操作仅在桌面应用中可用', failed: '操作失败', version: '版本', show: '显示密钥', hide: '隐藏密钥', settings: '设备配置', local: '本地端口', status: '客户端状态' },
  en: { brand: 'ChickReomte', workstation: 'Connection workspace', connection: 'Connection settings', relay: 'Relay server', device: 'This device ID', target: 'Target device ID (optional)', secret: 'Shared secret', tls: 'TLS encryption', port: 'Local dashboard port', save: 'Save settings', start: 'Start client', stop: 'Stop client', dashboard: 'Open sessions', logs: 'Runtime log', idle: 'Stopped', starting: 'Connecting', ready: 'Client running', saved: 'Settings saved', empty: 'No log entries', preview: 'Browser preview: native connections are available in the desktop app only', failed: 'Operation failed', version: 'Version', show: 'Show secret', hide: 'Hide secret', settings: 'Device settings', local: 'Local port', status: 'Client status' }
};
let language: 'zh' | 'en' = localStorage.getItem('language') === 'en' ? 'en' : 'zh';
let settings: Settings = { id: '', server: '127.0.0.1:6154', secret: '', target: '', tls: true, dashboard_port: 8080 };
let status: Status = { running: false, ready: false, log: '', version: '1.0.20260919' };
let busy = false;
const root = document.querySelector<HTMLDivElement>('#app')!;
const $ = <T extends HTMLElement>(selector: string) => root.querySelector<T>(selector)!;
const icons = () => createIcons({ icons: { Play, Square, Save, ExternalLink, Eye, EyeOff, Monitor, Server, Activity } });

function render() {
  const t = languages[language];
  document.documentElement.lang = language === 'zh' ? 'zh-CN' : 'en';
  root.innerHTML = `<header><div class="identity"><img src="/icon.png" alt=""><div><strong>${t.brand}</strong><span>${language === 'zh' ? 'ChickReomte' : '小鸡远程'}</span></div></div><select id="language" aria-label="Language"><option value="zh">简体中文</option><option value="en">English</option></select></header>
  <main><div class="page-heading"><div><span class="eyebrow">${t.workstation}</span><h1>${t.connection}</h1></div><span id="status" class="status"></span></div>
  <div class="workspace"><section class="configuration"><h2><i data-lucide="monitor"></i>${t.settings}</h2><form id="settings"><fieldset><div class="fields"><label>${t.device}<input id="id" autocomplete="off" maxlength="64" required></label><label>${t.target}<input id="target" autocomplete="off" maxlength="64"></label><label class="wide">${t.relay}<div class="input-icon"><i data-lucide="server"></i><input id="server" placeholder="relay.example.com:6154" required></div></label><label class="wide">${t.secret}<div class="secret-input"><input id="secret" type="password" autocomplete="off" minlength="16" maxlength="4096" required><button id="reveal" type="button" class="icon-button" title="${t.show}" aria-label="${t.show}"><i data-lucide="eye"></i></button></div></label><label>${t.port}<input id="dashboard_port" type="number" min="1024" max="65535" required></label><label class="toggle"><input id="tls" type="checkbox"><span>${t.tls}</span></label></div><div class="form-footer"><button class="secondary" id="save" type="submit"><i data-lucide="save"></i>${t.save}</button></div></fieldset></form></section>
  <aside><h2><i data-lucide="activity"></i>${t.status}</h2><div class="connection-state"><span id="indicator"></span><strong id="state-label"></strong></div><dl><div><dt>${t.local}</dt><dd id="port-label"></dd></div><div><dt>${t.version}</dt><dd>${status.version}</dd></div></dl><div class="actions"><button id="start" class="primary"><i data-lucide="play"></i>${t.start}</button><button id="stop" class="secondary"><i data-lucide="square"></i>${t.stop}</button><button id="dashboard" class="secondary"><i data-lucide="external-link"></i>${t.dashboard}</button></div></aside></div>
  <div id="message" role="status" aria-live="polite"></div><section class="logs"><h2>${t.logs}</h2><pre id="logs"></pre></section></main><footer>ChickReomte <span>org.mutantcat.chickreomte</span></footer>`;
  $<HTMLSelectElement>('#language').value = language;
  for (const key of ['id', 'server', 'secret', 'target', 'dashboard_port'] as const) $<HTMLInputElement>(`#${key}`).value = String(settings[key]);
  $<HTMLInputElement>('#tls').checked = settings.tls;
  $('#language').addEventListener('change', () => { settings = readForm(); language = $<HTMLSelectElement>('#language').value as 'zh' | 'en'; localStorage.setItem('language', language); render(); });
  $('#reveal').addEventListener('click', () => {
    const input = $<HTMLInputElement>('#secret'), show = input.type === 'password';
    input.type = show ? 'text' : 'password';
    $('#reveal').innerHTML = `<i data-lucide="${show ? 'eye-off' : 'eye'}"></i>`;
    $('#reveal').title = show ? t.hide : t.show;
    $('#reveal').setAttribute('aria-label', show ? t.hide : t.show); icons();
  });
  $('#settings').addEventListener('submit', (event) => { event.preventDefault(); void action(async () => { settings = readForm(); await invoke('save_settings', { settings }); message(t.saved); }); });
  $('#start').addEventListener('click', () => { if (!$<HTMLFormElement>('#settings').reportValidity()) return; void action(async () => { settings = readForm(); await invoke('save_settings', { settings }); await invoke('start_client'); }); });
  $('#stop').addEventListener('click', () => void action(() => invoke('stop_client')));
  $('#dashboard').addEventListener('click', () => void action(() => invoke('open_dashboard')));
  icons(); update();
  if (!isTauri()) message(t.preview);
}

function readForm(): Settings {
  return { id: $<HTMLInputElement>('#id').value.trim(), server: $<HTMLInputElement>('#server').value.trim(), secret: $<HTMLInputElement>('#secret').value, target: $<HTMLInputElement>('#target').value.trim(), tls: $<HTMLInputElement>('#tls').checked, dashboard_port: Number($<HTMLInputElement>('#dashboard_port').value) };
}
function message(text: string, error = false) { $('#message').textContent = text; $('#message').className = error ? 'error' : ''; }
function update() {
  const t = languages[language], text = status.ready ? t.ready : status.running ? t.starting : t.idle;
  $('#status').textContent = text; $('#status').className = `status ${status.ready ? 'online' : ''}`;
  $('#state-label').textContent = text; $('#indicator').className = status.ready ? 'online' : status.running ? 'pending' : '';
  $('#port-label').textContent = `127.0.0.1:${settings.dashboard_port}`;
  $('#logs').textContent = status.log || t.empty;
  $<HTMLFieldSetElement>('fieldset').disabled = status.running || busy;
  $<HTMLButtonElement>('#start').disabled = status.running || busy || !isTauri();
  $<HTMLButtonElement>('#stop').disabled = !status.running || busy;
  $<HTMLButtonElement>('#dashboard').disabled = !status.ready || busy;
}
async function action(work: () => Promise<unknown>) {
  if (busy) return;
  if (!isTauri()) { message(languages[language].preview, true); return; }
  busy = true; message(''); update();
  try { await work(); await refresh(); } catch (error) { message(`${languages[language].failed}: ${String(error)}`, true); }
  finally { busy = false; update(); }
}
async function refresh() {
  if (!isTauri()) return;
  status = await invoke<Status>('client_status'); update();
}
render();
if (isTauri()) {
  try { settings = await invoke<Settings>('load_settings'); render(); await refresh(); }
  catch (error) { message(String(error), true); }
  let polling = false;
  setInterval(async () => { if (polling || busy) return; polling = true; try { await refresh(); } catch (error) { message(String(error), true); } finally { polling = false; } }, 1500);
}
