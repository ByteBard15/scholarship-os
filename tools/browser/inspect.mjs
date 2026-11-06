import { readFile } from 'node:fs/promises';
const [tool = 'browser_snapshot', json = '{}'] = process.argv.slice(2);
const token = await readFile(new URL('../../.var/application-browser/bridge-token', import.meta.url), 'utf8');
const response = await fetch('http://127.0.0.1:8932/inspect', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
  body: JSON.stringify({ tool, args: JSON.parse(json) }),
});
const result = await response.json();
if (!response.ok) throw new Error(result.error ?? 'Inspection failed');
if (result.tools) console.log(JSON.stringify(result.tools, null, 2));
for (const item of result.content ?? []) if (item.type === 'text') console.log(item.text);
if (result.isError) process.exitCode = 1;
