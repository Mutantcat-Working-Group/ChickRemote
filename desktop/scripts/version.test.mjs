import test from 'node:test';
import assert from 'node:assert/strict';
import { windowsVersion, checkTag } from './version.mjs';

test('maps date release into four valid Windows version components', () => {
  assert.equal(windowsVersion('1.0.20260919'), '1.0.2026+919');
  assert.equal(windowsVersion('1.0.20260920'), '1.0.2026+920');
  assert.equal(windowsVersion('1.0.20260101'), '1.0.2026+101');
  assert.equal(windowsVersion('1.2.3'), '1.2.3');
});
test('rejects malformed dates, oversized components and mismatched tags', () => {
  for (const version of ['1.0.20261319', '1.0.20260230', '65536.0.1', '1.0.999999999']) {
    assert.throws(() => windowsVersion(version));
  }
  assert.throws(() => checkTag('v1.0.20260918', '1.0.20260919'));
  assert.doesNotThrow(() => checkTag('v1.0.20260919', '1.0.20260919'));
  assert.throws(() => checkTag('v1.0.20260919', '1.0.20260920'));
  assert.doesNotThrow(() => checkTag('v1.0.20260920', '1.0.20260920'));
});
