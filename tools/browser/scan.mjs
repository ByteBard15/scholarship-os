import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js';
import { readFile, writeFile, mkdir } from 'node:fs/promises';
const client = new Client({ name: 'scholarship-section-audit', version: '0.1.0' });
const root = new URL('../../.var/application-browser/workspace/', import.meta.url);
const out = new URL('../../tmp/calgary-portal-prefill/', import.meta.url);
await mkdir(out, { recursive: true, mode: 0o700 });
try {
 await client.connect(new StreamableHTTPClientTransport(new URL('http://127.0.0.1:8931/mcp')));
 for (const name of ['Program-Specific Questions','Contact Information','Citizenship','Post-Secondary Education','Test Scores','Referees','Personal Information','Addresses','Program Information']) {
  const clicked=await client.callTool({name:'browser_click',arguments:{target:`role=button[name="${name}"]`,element:`${name} section for inspection`}});
  if(clicked.isError) throw new Error(JSON.stringify(clicked.content));
  const filename='audit-'+name.toLowerCase().replaceAll(' ','-')+'.yml';
  const snapshot=await client.callTool({name:'browser_snapshot',arguments:{filename}});
  if(snapshot.isError) throw new Error(JSON.stringify(snapshot.content));
  const text=await readFile(new URL(filename,root),'utf8');
  await writeFile(new URL(filename,out),text,{mode:0o600});
  console.log(`\nSECTION: ${name}\n`+text.split('\n').filter(line=>!line.includes('option ')&&!/button \"(Read This First|Personal Information|Addresses|Contact Information|Citizenship|Post-Secondary Education|Program Information|Program-Specific Questions|Test Scores|Referees|Review and Submit Application)\"/.test(line)).join('\n'));
 }
} finally { await client.close(); }
