const fs = require('fs');
const t = fs.readFileSync('backend/launcherupdate.go', 'utf8');
for (const m of t.matchAll(/^func\s+(\([^)]*\)\s*)?(\w+)/gm)) {
  console.log('m[1]=', JSON.stringify(m[1]), 'm[2]=', JSON.stringify(m[2]));
  let n = m[2];
  if (m[1]) {
    const r = m[1].replace(/^\(|\)$/g, '').trim();
    console.log('  r=', JSON.stringify(r));
    const parts = r.split(/\s+/);
    console.log('  parts=', JSON.stringify(parts));
    const typeName = parts[parts.length - 1].replace(/^\*/, '');
    console.log('  typeName=', JSON.stringify(typeName));
    n = typeName + '.' + n;
  }
  console.log('  n=', JSON.stringify(n));
}
