import { readdirSync, mkdirSync, copyFileSync, writeFileSync, readFileSync, mkdtempSync, rmdirSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
const version = JSON.parse(readFileSync('package.json')).version;
const platform = process.env.ARTIFACT_PLATFORM;
if (!platform) throw new Error('ARTIFACT_PLATFORM is required');
const extension = platform.startsWith('windows') ? '.exe' : platform.startsWith('macos') ? '.dmg' : '.AppImage';
const folder = extension === '.exe' ? 'nsis' : extension === '.dmg' ? 'dmg' : 'appimage';
const source = `src-tauri/target/release/bundle/${folder}`;
const files = readdirSync(source).filter(name => name.endsWith(extension));
if (files.length !== 1) throw new Error(`Expected one installer, got ${files.length}`);
mkdirSync('artifacts', { recursive: true });
const name = `ChickReomte_${version}_${platform}${extension}`;
const output = join('artifacts', name);
copyFileSync(join(source, files[0]), output);
if (extension === '.dmg') {
  const mount = mkdtempSync(join(tmpdir(), 'chickreomte-dmg-'));
  execFileSync('hdiutil', ['attach', output, '-mountpoint', mount, '-nobrowse', '-readonly'], { input: 'Y\n', stdio: ['pipe', 'inherit', 'inherit'] });
  try {
    execFileSync('codesign', ['--verify', '--deep', '--strict', join(mount, 'ChickReomte.app')], { stdio: 'inherit' });
  } finally {
    execFileSync('hdiutil', ['detach', mount], { stdio: 'inherit' });
    rmdirSync(mount);
  }
  execFileSync('codesign', ['--force', '--sign', '-', output], { stdio: 'inherit' });
  execFileSync('codesign', ['--verify', '--verbose=2', output], { stdio: 'inherit' });
  execFileSync('hdiutil', ['verify', output], { stdio: 'inherit' });
}
const hash = createHash('sha256').update(readFileSync(output)).digest('hex');
writeFileSync(`${output}.sha256`, `${hash}  ${name}\n`);
console.log(output);
