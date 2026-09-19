import assert from 'node:assert/strict';
import { execFileSync, spawn } from 'node:child_process';
import { once } from 'node:events';
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { createServer, connect } from 'node:net';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { setTimeout as delay } from 'node:timers/promises';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..');
const dir = mkdtempSync(join(tmpdir(), 'chickreomte-smoke-'));
const target = /host: (.+)/.exec(execFileSync('rustc', ['-vV'], { encoding: 'utf8' }))[1];
const suffix = process.platform === 'win32' ? '.exe' : '';
const children = [];
let output = '';
async function freePort() {
  const server = createServer().listen(0, '127.0.0.1');
  await once(server, 'listening');
  const port = server.address().port;
  await new Promise(resolve => server.close(resolve));
  return port;
}
async function waitFor(probe) {
  for (let i = 0; i < 100; i++) {
    if (children.some(child => child.exitCode !== null)) throw new Error(`Process exited: ${output}`);
    try { return await probe(); } catch { await delay(100); }
  }
  throw new Error(`Client readiness timed out: ${output}`);
}
function start(exe, config) {
  const child = spawn(exe, ['--conf', config], { cwd: dir, env: { ...process.env, CHICKREOMTE_DESKTOP: '1' } });
  child.stdout.on('data', data => { output += data; });
  child.stderr.on('data', data => { output += data; });
  children.push(child);
}
try {
  const relay = join(dir, `relay${suffix}`);
  execFileSync('go', ['build', '-o', relay, './code/server'], { cwd: root, stdio: 'inherit' });
  const relayPort = await freePort();
  const dashboardPort = await freePort();
  const secret = 'local-smoke-test-secret';
  const relayConfig = join(dir, 'server.yaml');
  writeFileSync(relayConfig, JSON.stringify({ listen: relayPort, secret, log: { dir: join(dir, 'server-logs') } }));
  start(relay, relayConfig);
  await waitFor(() => new Promise((resolve, reject) => {
    const socket = connect(relayPort, '127.0.0.1');
    socket.once('connect', () => { socket.destroy(); resolve(); });
    socket.once('error', reject);
  }));
  const config = join(dir, 'client.yaml');
  writeFileSync(config, JSON.stringify({ id: 'smoke-client', server: `127.0.0.1:${relayPort}`, secret,
    ssl: { enabled: false }, dashboard: { enabled: true, listen: '127.0.0.1', port: dashboardPort },
    log: { dir: join(dir, 'client-logs') }, codedir: join(dir, 'code'), rules: [] }));
  const binary = join(root, `desktop/src-tauri/binaries/chickreomte-cli-${target}${suffix}`);
  const version = JSON.parse(readFileSync(join(root, 'desktop/package.json'))).version;
  assert.match(execFileSync(binary, ['version'], { encoding: 'utf8' }), new RegExp(version.replaceAll('.', '\\.')));
  start(binary, config);
  const info = await waitFor(async () => {
    const response = await fetch(`http://127.0.0.1:${dashboardPort}/api/info`, { signal: AbortSignal.timeout(1000) });
    assert.equal(response.status, 200);
    return response.json();
  });
  assert.deepEqual(info, { rules: 0, virtual_links: 0, sessions: 0 });
  const page = await fetch(`http://127.0.0.1:${dashboardPort}/`);
  assert.ok((await page.text()).includes(`v${version}</title>`));
  for (const asset of ['/AdminLTE/css/adminlte.min.css', '/jquery/jquery-3.6.3.min.js', '/js/common.js']) {
    const response = await fetch(`http://127.0.0.1:${dashboardPort}${asset}`);
    assert.equal(response.status, 200, asset);
    assert.ok((await response.text()).length > 100, asset);
  }
  console.log(`Sidecar ${version}: relay handshake, embedded dashboard and API passed`);
} finally {
  await Promise.all(children.map(async child => {
    if (child.exitCode === null) {
      const exited = once(child, 'exit');
      child.kill('SIGKILL');
      await exited;
    }
  }));
  rmSync(dir, { recursive: true, force: true });
}
