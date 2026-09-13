const status = document.querySelector("#status"), file = document.querySelector("#file"), button = document.querySelector("#download");
const params = new URLSearchParams(location.hash.slice(1)), addr = params.get("tc");
let conn, hello;
async function read() { const b = await conn.read(); if (!b) throw Error("closed"); return b; }
function frame(m) { const b = new TextEncoder().encode(JSON.stringify(m)), o = new Uint8Array(4+b.length); new DataView(o.buffer).setUint32(0,b.length); o.set(b,4); return o; }
async function connect() { if (params.get("v") !== "1" || !addr) { status.textContent="Invalid link"; return; } try { conn=await tailcatDial({addr}); const h=await read(), n=new DataView(h.buffer).getUint32(0); const b=await read(); hello=JSON.parse(new TextDecoder().decode(b.slice(0,n))); file.textContent=`${hello.name}\n${(hello.size/1e6).toFixed(1)} MB`; button.hidden=false; status.textContent="Ready to download"; } catch { status.textContent="Could not connect — reload to retry"; } }
button.onclick=async()=>{ button.disabled=true; status.textContent="Downloading…"; try { await conn.write(frame({type:"get"})); let a=[],n=0; while(n<hello.size){const b=await read();a.push(b);n+=b.length;} const blob=new Blob(a,{type:hello.mime||"application/octet-stream"}), link=document.createElement("a"); link.href=URL.createObjectURL(blob); link.download=hello.name; link.click(); status.textContent="Downloaded"; conn.close(); } catch { status.textContent="Download failed"; button.disabled=false; } };
connect();
