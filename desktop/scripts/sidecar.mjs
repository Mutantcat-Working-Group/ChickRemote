import { execFileSync } from 'node:child_process';
import { mkdirSync, copyFileSync, readFileSync, existsSync, lstatSync, readdirSync, readlinkSync, unlinkSync } from 'node:fs';
import { dirname, resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..');
const pkg = JSON.parse(readFileSync(join(root, 'desktop/package.json')));
const target = process.env.TARGET_TRIPLE || /host: (.+)/.exec(execFileSync('rustc', ['-vV'], { encoding: 'utf8' }))[1];
const windows = target.includes('windows');
const arch = target.startsWith('aarch64') ? 'arm64' : 'amd64';
const goos = windows ? 'windows' : target.includes('apple') ? 'darwin' : 'linux';

// Windows checkouts may contain Git symlink placeholders instead of symlinks.
function materialize(dir) {
  for (const name of readdirSync(dir)) {
    const path = join(dir, name), stat = lstatSync(path);
    if (stat.isDirectory()) materialize(path);
    else if (stat.isSymbolicLink()) {
      const source = resolve(dirname(path), readlinkSync(path));
      unlinkSync(path); copyFileSync(source, path);
    } else if (windows && stat.size < 180) {
      const value = readFileSync(path, 'utf8').trim();
      if (/^(\.\.\/)+[^\r\n]+$/.test(value) && existsSync(resolve(dirname(path), value))) {
        copyFileSync(resolve(dirname(path), value), path);
      }
    }
  }
}
if (windows) materialize(join(root, 'html'));
for (const name of ['shell', 'vnc', 'code', 'dashboard']) {
  const dest = name === 'dashboard' ? 'code/client/dashboard' : `code/client/rule/${name}`;
  execFileSync('go', ['run', 'contrib/bindata/main.go', '-pkg', name, '-o', `${dest}/assets.go`, '-prefix', `html/${name}`, `html/${name}/...`], { cwd: root, stdio: 'inherit' });
}
const output = join(root, `desktop/src-tauri/binaries/chickreomte-cli-${target}${windows ? '.exe' : ''}`);
mkdirSync(dirname(output), { recursive: true });
const hash = execFileSync('git', ['rev-parse', '--short', 'HEAD'], { cwd: root, encoding: 'utf8' }).trim();
const flags = `-s -w -X main.version=${pkg.version} -X main.gitHash=${hash}` + (windows ? ' -extldflags=-static' : '');
execFileSync('go', ['build', '-trimpath', '-ldflags', flags, '-o', output, './code/client'], {
  cwd: root, stdio: 'inherit', env: { ...process.env, CGO_ENABLED: '1', GOOS: goos, GOARCH: arch, MACOSX_DEPLOYMENT_TARGET: '14.0' }
});
console.log(`Sidecar ready: ${output}`);
