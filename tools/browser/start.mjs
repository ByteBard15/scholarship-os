import { createServer } from 'node:http';
import { randomBytes } from 'node:crypto';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js';
import { mkdir, chmod, writeFile } from 'node:fs/promises';
import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const root = fileURLToPath(new URL('../../', import.meta.url));
const state = path.join(root, '.var', 'application-browser');
for (const directory of [state, path.join(state, 'profile'), path.join(state, 'workspace'), path.join(state, 'workspace', 'output')]) {
  await mkdir(directory, { recursive: true, mode: 0o700 });
  await chmod(directory, 0o700);
}
const args = [
  fileURLToPath(new URL('./node_modules/@playwright/mcp/cli.js', import.meta.url)),
  '--host', '127.0.0.1', '--port', '8931', '--browser', 'chrome',
  '--user-data-dir', path.join(state, 'profile'),
  '--output-dir', path.join(state, 'workspace', 'output'),
  '--shared-browser-context', '--ignore-roots', '--no-webmcp', '--codegen', 'none',
  '--redact-url-params', 'access_token,id_token,token,code,password,session,ticket',
];
if (process.env.SCHOLARSHIP_BROWSER_EXECUTABLE) args.push('--executable-path', process.env.SCHOLARSHIP_BROWSER_EXECUTABLE);
const child = spawn(process.execPath, args, { cwd: path.join(state, 'workspace'), stdio: 'inherit' });
for (const signal of ['SIGINT', 'SIGTERM']) process.on(signal, () => child.kill(signal));
child.on('error', (error) => { console.error(error.message); process.exitCode = 1; });
child.on('exit', (code) => { process.exitCode = code ?? 1; });

// Keep one MCP client connected so short-lived inspection commands share the tab.
const client = new Client({ name: 'scholarship-browser-bridge', version: '0.1.0' });
let connected = false;
for (let attempt = 0; attempt < 40; attempt++) {
  try { await client.connect(new StreamableHTTPClientTransport(new URL('http://127.0.0.1:8931/mcp'))); connected = true; break; }
  catch { await new Promise((resolve) => setTimeout(resolve, 250)); }
}
if (!connected) { child.kill(); throw new Error('Could not connect to local Playwright server'); }
const token = randomBytes(32).toString('hex');
await writeFile(path.join(state, 'bridge-token'), token, { mode: 0o600 });
const permitted = new Set(['browser_snapshot', 'browser_navigate', 'browser_tabs', 'browser_take_screenshot']);
const bridge = createServer(async (request, response) => {
  if (request.method !== 'POST' || request.url !== '/inspect' || request.headers.authorization !== `Bearer ${token}`) { response.writeHead(403).end(); return; }
  try {
    let body = '';
    for await (const chunk of request) { body += chunk; if (body.length > 20000) throw new Error('Request too large'); }
    const { tool, args = {} } = JSON.parse(body);
    if (tool !== 'list' && !permitted.has(tool)) throw new Error('Inspection tools only');
    if (tool === 'browser_tabs' && args.action && !['list', 'select'].includes(args.action)) throw new Error('Only list/select tabs');
    const result = tool === 'list' ? { tools: (await client.listTools()).tools.filter((entry) => permitted.has(entry.name)) } : await client.callTool({ name: tool, arguments: args });
    response.writeHead(200, { 'Content-Type': 'application/json' }).end(JSON.stringify(result));
  } catch (error) { response.writeHead(400, { 'Content-Type': 'application/json' }).end(JSON.stringify({ error: error.message })); }
});
bridge.listen(8932, '127.0.0.1', () => console.error('Inspection bridge ready on localhost:8932'));
child.on('exit', () => { bridge.close(); void client.close(); });
