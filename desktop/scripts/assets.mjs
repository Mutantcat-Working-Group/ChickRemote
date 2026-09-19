import { cpSync, lstatSync, readFileSync, readlinkSync, rmSync } from 'node:fs';
import { dirname, resolve } from 'node:path';

// Git identifies placeholders reliably; short text assets are not symlinks.
export function materializeLinks(root, paths) {
  for (const relative of paths) {
    const path = resolve(root, relative);
    const stat = lstatSync(path);
    if (stat.isDirectory()) continue;
    const target = stat.isSymbolicLink() ? readlinkSync(path) : readFileSync(path, 'utf8').trim();
    if (!stat.isSymbolicLink() && !/^(\.\.\/)+[^\r\n]+$/.test(target)) continue;
    const source = resolve(dirname(path), target);
    if (source === path) throw new Error(`Self-referencing asset: ${relative}`);
    lstatSync(source);
    rmSync(path);
    cpSync(source, path, { recursive: true, dereference: true });
  }
}
