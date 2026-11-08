// Applicant-authorized draft interactions. No arbitrary code, submission or credentials.
import { readFile } from 'node:fs/promises';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js';
const [tool = 'list', json = '{}', tabIndex] = process.argv.slice(2);
const permitted = new Set(['browser_click', 'browser_fill_form', 'browser_select_option', 'browser_snapshot', 'browser_tabs', 'browser_handle_dialog', 'browser_file_upload']);
if (tool !== 'list' && !permitted.has(tool)) throw new Error('Unsupported draft interaction');
const args = JSON.parse(json.startsWith('@') ? await readFile(json.slice(1), 'utf8') : json);
if (tool === 'browser_tabs' && args.action && !['list', 'select', 'new', 'close'].includes(args.action)) throw new Error('Unsupported tab action');
if (tool === 'browser_fill_form' && (args.fields ?? []).some((field) => /password|verification code|one.time|otp/i.test(field.name ?? ''))) throw new Error('Applicant must enter credentials and verification codes directly');
const reviewNavigation = args.target === 'role=button[name="Review and Submit Application"]';
if (tool === 'browser_click' && !reviewNavigation && /\b(?:submit|pay|payment|declaration|certify|sign in|delete)\b/i.test(args.element ?? '')) throw new Error('This action requires an applicant handoff');
const client = new Client({ name: 'scholarship-draft-assistance', version: '0.1.0' });
try {
 await client.connect(new StreamableHTTPClientTransport(new URL('http://127.0.0.1:8931/mcp')));
 if (tabIndex !== undefined) {
  if (!/^\d+$/.test(tabIndex)) throw new Error('Tab index must be a nonnegative integer');
  await client.callTool({ name: 'browser_tabs', arguments: { action: 'list' } });
  const selected = await client.callTool({ name: 'browser_tabs', arguments: { action: 'select', index: Number(tabIndex) } });
  if (selected.isError && tool !== 'browser_handle_dialog') throw new Error(JSON.stringify(selected.content));
 }
 const result = tool === 'list' ? { tools: (await client.listTools()).tools.filter((entry) => permitted.has(entry.name)) } : await client.callTool({ name: tool, arguments: args });
 if (result.tools) console.log(JSON.stringify(result.tools, null, 2));
 for (const item of result.content ?? []) if (item.type === 'text') console.log(item.text);
 if (result.isError) process.exitCode = 1;
} finally { await client.close(); }
