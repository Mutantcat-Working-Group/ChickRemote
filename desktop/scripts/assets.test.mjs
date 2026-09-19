import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, writeFileSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { materializeLinks } from './assets.mjs';

test('materializes Windows file and directory placeholders without changing text assets', () => {
  const root = mkdtempSync(join(tmpdir(), 'chickreomte-assets-'));
  try {
    mkdirSync(join(root, 'library'));
    mkdirSync(join(root, 'view'));
    writeFileSync(join(root, 'library/main.js'), 'window.ready = true;');
    writeFileSync(join(root, 'view/library'), '../library');
    writeFileSync(join(root, 'view/main.js'), '../library/main.js');
    writeFileSync(join(root, 'view/text.txt'), '../library/main.js');
    materializeLinks(root, ['view/library', 'view/main.js']);
    materializeLinks(root, ['view/library', 'view/main.js']);
    assert.equal(readFileSync(join(root, 'view/library/main.js'), 'utf8'), 'window.ready = true;');
    assert.equal(readFileSync(join(root, 'view/main.js'), 'utf8'), 'window.ready = true;');
    assert.equal(readFileSync(join(root, 'view/text.txt'), 'utf8'), '../library/main.js');
  } finally { rmSync(root, { recursive: true, force: true }); }
});
