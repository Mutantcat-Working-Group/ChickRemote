export function windowsVersion(version) {
  const match = /^(\d+)\.(\d+)\.(\d+)$/.exec(version);
  if (!match) throw new Error(`Invalid release version: ${version}`);
  const [, major, minor, patch] = match.map(Number);
  if (major > 65535 || minor > 65535) throw new Error('Windows version component exceeds 65535');
  if (patch <= 65535) return version;
  const date = String(patch);
  const year = Number(date.slice(0, 4)), month = Number(date.slice(4, 6)), day = Number(date.slice(6));
  const parsed = new Date(Date.UTC(year, month - 1, day));
  if (date.length !== 8 || year < 2000 || parsed.getUTCFullYear() !== year || parsed.getUTCMonth() !== month - 1 || parsed.getUTCDate() !== day) {
    throw new Error('Date version must contain a valid YYYYMMDD');
  }
  return `${major}.${minor}.${year}+${month * 100 + day}`;
}

export function checkTag(tag, version) {
  if (tag !== `v${version}`) throw new Error(`Tag ${tag} does not match v${version}`);
}
