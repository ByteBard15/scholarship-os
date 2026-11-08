import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js';
import { readFile, writeFile, mkdir } from 'node:fs/promises';
const client=new Client({name:'toronto-draft-audit',version:'0.1.0'});
const root=new URL('../../.var/application-browser/workspace/',import.meta.url);
const out=new URL('../../tmp/toronto-portal-prefill/',import.meta.url);
await mkdir(out,{recursive:true,mode:0o700});
try {
 await client.connect(new StreamableHTTPClientTransport(new URL('http://127.0.0.1:8931/mcp')));
 for(const name of ['English Language Proficiency Requirement Check','Additional Information','Faculty Questions: BME','Portfolio Upload','Recommendations','Surveys','Review']) {
  const clicked=await client.callTool({name:'browser_click',arguments:{target:`role=link[name="${name}"s]`,element:`Read ${name} section without accepting declarations or sending anything`}});
  if(clicked.isError)throw new Error(JSON.stringify(clicked.content));
  const filename='toronto-audit-'+name.toLowerCase().replaceAll(/[^a-z0-9]+/g,'-')+'.yml';
  const snap=await client.callTool({name:'browser_snapshot',arguments:{filename}});
  if(snap.isError)throw new Error(JSON.stringify(snap.content));
  const text=await readFile(new URL(filename,root),'utf8');await writeFile(new URL(filename,out),text,{mode:0o600});
  const start=text.search(/heading .*\[level=1\]/);const end=text.indexOf('  - contentinfo');
  console.log('\nSECTION '+name+'\n'+text.slice(start,end<0?undefined:end).split('\n').filter(line=>!line.includes('option ')).join('\n'));
 }
}finally{await client.close()}
