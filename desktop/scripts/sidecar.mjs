import { execFileSync } from 'node:child_process';
import { mkdirSync, readFileSync } from 'node:fs';
import { dirname, resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { materializeLinks } from './assets.mjs';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..');
const pkg = JSON.parse(readFileSync(join(root, 'desktop/package.json')));
const target = process.env.TARGET_TRIPLE || /host: (.+)/.exec(execFileSync('rustc', ['-vV'], { encoding: 'utf8' }))[1];
const windows = target.includes('windows');
const arch = target.startsWith('aarch64') ? 'arm64' : 'amd64';
const goos = windows ? 'windows' : target.includes('apple') ? 'darwin' : 'linux';
// The aarch64 Go sidecar has no cgo sources. A static build keeps the AppImage
// copy runnable when linuxdeploy rewrites RUNPATH for bundled ELF files.
const cgoEnabled = goos === 'linux' && arch === 'arm64' ? '0' : '1';

if (windows) {
  const entries = execFileSync('git', ['ls-files', '-s', '-z', 'html'], { cwd: root, encoding: 'utf8' });
  materializeLinks(root, entries.split('\0').filter(entry => entry.startsWith('120000 ')).map(entry => entry.split('\t')[1]));
}
for (const name of ['shell', 'vnc', 'code', 'dashboard']) {
  const dest = name === 'dashboard' ? 'code/client/dashboard' : `code/client/rule/${name}`;
  execFileSync('go', ['run', 'contrib/bindata/main.go', '-pkg', name, '-o', `${dest}/assets.go`, '-prefix', `html/${name}`, `html/${name}/...`], { cwd: root, stdio: 'inherit' });
}
const output = join(root, `desktop/src-tauri/binaries/chickreomte-cli-${target}${windows ? '.exe' : ''}`);
mkdirSync(dirname(output), { recursive: true });
const hash = execFileSync('git', ['rev-parse', '--short', 'HEAD'], { cwd: root, encoding: 'utf8' }).trim();
const flags = `-s -w -X main.version=${pkg.version} -X main.gitHash=${hash}` + (windows ? ' -extldflags=-static' : '');
execFileSync('go', ['build', '-trimpath', '-ldflags', flags, '-o', output, './code/client'], {
  cwd: root, stdio: 'inherit', env: { ...process.env, CGO_ENABLED: cgoEnabled, GOOS: goos, GOARCH: arch, MACOSX_DEPLOYMENT_TARGET: '14.0' }
});
console.log(`Sidecar ready: ${output}`);
