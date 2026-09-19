import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';
import { windowsVersion, checkTag } from './version.mjs';
import { fileURLToPath } from 'node:url';

const pkg = JSON.parse(readFileSync(new URL('../package.json', import.meta.url)));
const config = JSON.parse(readFileSync(new URL('../src-tauri/tauri.conf.json', import.meta.url)));
if (pkg.version !== config.version) throw new Error('Package and Tauri versions differ');
if (process.env.GITHUB_REF_TYPE === 'tag') checkTag(process.env.GITHUB_REF_NAME, pkg.version);
const args = ['build', ...process.argv.slice(2)];
if (process.platform === 'win32') {
  const override = 'src-tauri/tauri.windows.generated.json';
  writeFileSync(override, JSON.stringify({ version: windowsVersion(pkg.version) }));
  args.push('--config', override);
}
const cli = new URL('../node_modules/@tauri-apps/cli/tauri.js', import.meta.url);
execFileSync(process.execPath, [fileURLToPath(cli), ...args], { stdio: 'inherit' });
