'use strict';
// Local synthetic feasibility only. There is deliberately no live URL mode.
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');

async function main() {
  const args = process.argv.slice(2);
  if (args.length !== 6 || args[0] !== '--playwright-root' || args[2] !== '--browser' || args[4] !== '--output') throw new Error('arguments');
  const repo = path.resolve(__dirname, '..');
  const output = path.resolve(args[5]);
  if (!output.startsWith(path.join(repo, 'build') + path.sep) || fs.existsSync(output)) throw new Error('output');
  fs.mkdirSync(output, { recursive: true });
  const temp = path.join(output, 'tmp');
  fs.mkdirSync(temp);
  process.env.TEMP = temp; process.env.TMP = temp; process.env.TMPDIR = temp;
  const pwRoot = path.resolve(args[1]);
  const { chromium } = require(pwRoot);
  const report = {
    kind: 'browser-comparison-offline', schema_version: 1, synthetic_only: true,
    real_campus_mode_available: false, node: process.version,
    playwright: JSON.parse(fs.readFileSync(path.join(pwRoot, 'package.json'), 'utf8')).version,
    script_sha256: crypto.createHash('sha256').update(fs.readFileSync(__filename)).digest('hex'),
    checks: [], passed: false, browser_closed: false,
    limits: { wall_seconds: 90, case_seconds: 8 },
    scope: 'loopback_fixtures_only_not_os_traffic_proof',
  };
  const save = () => fs.writeFileSync(path.join(output, 'result.json'), JSON.stringify(report, null, 2) + '\n');
  let browser;
  const hard = setTimeout(() => { report.stop = 'deadline'; save(); process.exit(124); }, 90000);
  hard.unref();
  const requests = [];
  let deniedProxy = 0, socketArrivals = 0;
  const fixtures = new Map();
  const server = http.createServer((req, res) => {
    // Never read headers, cookies or bodies. Test-only closed route labels.
    const name = new URL(req.url, 'http://127.0.0.1').pathname;
    requests.push({ method: req.method, name: fixtures.has(name) ? name : 'sink' });
    const item = fixtures.get(name);
    if (!item || req.method !== 'GET') { res.writeHead(409); res.end(); return; }
    res.writeHead(item.status || 200, { 'Content-Type': item.type || 'text/html', 'Cache-Control': 'no-store', ...(item.headers || {}) });
    res.end(item.body || '');
  });
  server.on('upgrade', (req, socket) => { socketArrivals++; socket.destroy(); });
  const proxy = http.createServer((req, res) => { deniedProxy++; res.writeHead(502); res.end(); });
  proxy.on('connect', (req, socket) => { deniedProxy++; socket.on('error',()=>{}); socket.end('HTTP/1.1 502 Rejected\r\n\r\n'); });
  for (const listener of [server,proxy]) {
    listener.on('connection', socket=>socket.on('error',()=>{}));
    listener.on('clientError', (_error,socket)=>socket.destroy());
  }
  const listen = s => new Promise(resolve => s.listen(0, '127.0.0.1', resolve));
  await listen(server); await listen(proxy);
  const origin = `http://127.0.0.1:${server.address().port}`;
  const proxyURL = `http://127.0.0.1:${proxy.address().port}`;
  fixtures.set('/page', { body: '<!doctype html><title>Synthetic comparison</title><div id="ready">ready</div>' });
  fixtures.set('/frame', { body: '<!doctype html><title>Synthetic frame</title>' });
  fixtures.set('/data', { type: 'application/json', body: '{"synthetic":true}' });
  fixtures.set('/redirect', { status: 302, headers: { Location: origin + '/sink' } });
  fixtures.set('/worker.js', { type: 'application/javascript', body: `fetch('${origin}/sink',{method:'POST',body:'synthetic'}).then(()=>postMessage('done'),()=>postMessage('denied'));` });
  fixtures.set('/worker-redirect.js', { type:'application/javascript', body:`fetch('${origin}/redirect').then(()=>postMessage('done'),()=>postMessage('denied'));` });
  fixtures.set('/sw.js', { type: 'application/javascript', body: 'self.addEventListener("fetch",()=>{});' });
  fixtures.set('/sw-page', { body: '<!doctype html><title>Worker fixture</title>' });
  const rsa = () => crypto.generateKeyPairSync('rsa', { modulusLength: 1024 }).publicKey.export({ type: 'spki', format: 'der' }).toString('base64');
  const decoy = rsa(), active = rsa();
  const script = [
    `// publicKeyStr='${decoy}';`,
    `const activeKey='${active}';`,
    'function consumeKey(key) {',
    '  const marker = key.length;',
    '  return marker;',
    '}',
    'function prepareSynthetic() { return consumeKey(activeKey); }',
    'document.querySelector("button").addEventListener("click", prepareSynthetic);',
  ].join('\n');
  fixtures.set('/logic.js', { type: 'application/javascript', body: script });
  fixtures.set('/call', { body: '<!doctype html><button>Prepare synthetic</button><script src="/logic.js"></script>' });
  const allowed = new Set(['/page', '/frame', '/data', '/redirect', '/worker.js', '/worker-redirect.js', '/sw-page', '/sw.js', '/call', '/logic.js']);
  async function context(budget = 32, cdpGate = false) {
    const ctx = await browser.newContext({ serviceWorkers: 'block', acceptDownloads: false });
    const counts = { attempts: 0, allowed: 0, blocked: 0, websocket_blocked: 0, budget_stopped: false };
    await ctx.routeWebSocket('**/*', socket => { counts.websocket_blocked++; socket.close(); });
    await ctx.route('**/*', async route => {
      const req = route.request();
      const u = new URL(req.url());
      const withinBudget = ++counts.attempts <= budget;
      if (!withinBudget) counts.budget_stopped = true;
      if (withinBudget && req.method() === 'GET' && u.origin === origin && !u.search && !u.hash && allowed.has(u.pathname)) {
        counts.allowed++;
        await route.continue();
      } else {
        counts.blocked++;
        await route.abort('blockedbyclient');
      }
    });
    ctx.setDefaultTimeout(8000);
    ctx.setDefaultNavigationTimeout(8000);
    async function newPage() {
      const page = await ctx.newPage();
      if (cdpGate) {
        const session = await ctx.newCDPSession(page);
        counts.cdp_attempts = 0; counts.cdp_blocked = 0;
        session.on('Fetch.requestPaused', async event => {
          const u = new URL(event.request.url);
          const permit = ++counts.cdp_attempts <= budget && event.request.method === 'GET' && u.origin === origin && !u.search && allowed.has(u.pathname);
          if (!permit) counts.cdp_blocked++;
          try { await session.send(permit ? 'Fetch.continueRequest' : 'Fetch.failRequest', { requestId: event.requestId, ...(permit ? {} : {errorReason:'BlockedByClient'}) }); }
          catch (_) { counts.cdp_control_failed = true; }
        });
        await session.send('Fetch.enable', { patterns:[{urlPattern:'*',requestStage:'Request'}] });
      }
      return page;
    }
    return { ctx, counts, newPage };
  }
  let caseFacts = {};
  async function check(name, work) {
    const start = requests.length, sockets = socketArrivals;
    caseFacts = {};
    let timer, expired = false;
    try {
      const facts = await Promise.race([work(), new Promise((_,reject)=>{timer=setTimeout(()=>{expired=true;reject(new Error('case_deadline'));},8000);})]);
      assert.equal(requests.slice(start).filter(r => r.name === 'sink' || r.method !== 'GET').length, 0);
      assert.equal(socketArrivals, sockets);
      report.checks.push({ name, passed: true, ...facts });
    } catch (error) {
      // Fixed classification only; do not serialize browser/CDP errors or data.
      report.checks.push({ name, passed: false, reason: expired ? 'case_deadline' : error instanceof assert.AssertionError ? 'assertion_failed' : 'operation_failed', facts:caseFacts, sink_arrivals:requests.slice(start).filter(r=>r.name==='sink'||r.method!=='GET').length });
    }
    clearTimeout(timer);
    save();
    if(expired) throw new Error('case_deadline');
  }
  try {
    browser = await chromium.launch({
      executablePath: path.resolve(args[3]), headless: true, chromiumSandbox: true,
      proxy: { server: proxyURL, bypass: '127.0.0.1' },
      args: ['--disable-quic', '--host-resolver-rules=MAP * ~NOTFOUND, EXCLUDE 127.0.0.1'],
    });
    report.browser = browser.version();
    await check('fresh_context_and_storage_separation', async () => {
      const a = await context(), b = await context();
      try {
        assert.equal((await a.ctx.cookies()).length, 0);
        assert.equal((await b.ctx.cookies()).length, 0);
        const pa = await a.ctx.newPage(), pb = await b.ctx.newPage();
        await pa.goto(origin + '/page'); await pb.goto(origin + '/page');
        await pa.evaluate(() => { document.cookie = 'fixture=synthetic'; localStorage.setItem('fixture', 'synthetic'); });
        const isolated = await pb.evaluate(() => document.cookie === '' && localStorage.getItem('fixture') === null);
        assert.ok(isolated);
        return { separate: true, initially_empty: true };
      } finally { await a.ctx.close(); await b.ctx.close(); }
    });
    await check('fetch_xhr_beacon_forms', async () => {
      const {ctx, counts} = await context();
      try {
        const page = await ctx.newPage(); await page.goto(origin + '/page');
        await page.evaluate(async () => {
          const done = [fetch('/sink',{method:'POST',body:'synthetic'}).catch(()=>null)];
          done.push(new Promise(resolve => { const x=new XMLHttpRequest(); x.open('POST','/sink'); x.onloadend=resolve; x.send('synthetic'); }));
          navigator.sendBeacon('/sink','synthetic');
          const frame=document.createElement('iframe'); frame.name='form-target'; document.body.append(frame);
          const form=document.createElement('form'); form.method='POST'; form.action='/sink'; form.target='form-target'; document.body.append(form); form.submit();
          await Promise.all(done);
        });
        await page.waitForFunction(() => document.readyState === 'complete');
        await new Promise(resolve => setTimeout(resolve, 200));
        assert.ok(counts.blocked >= 4);
        return { counts };
      } finally { await ctx.close(); }
    });
    await check('redirect_unknown_query_and_cross_origin', async () => {
      const {ctx, counts} = await context();
      try {
        const page = await ctx.newPage(); await page.goto(origin + '/page');
        await page.evaluate(async () => {
          await Promise.all(['/redirect','/data?extra=synthetic','http://example.invalid/synthetic'].map(u=>fetch(u).catch(()=>null)));
        });
        caseFacts = {counts};
        assert.ok(counts.blocked >= 3);
        return { counts };
      } finally { await ctx.close(); }
    });
    await check('cdp_per_hop_redirect_guard', async () => {
      const {ctx, counts, newPage} = await context(32,true);
      try {
        const page = await newPage(); await page.goto(origin+'/page');
        await page.evaluate(async()=>fetch('/redirect').catch(()=>null));
        caseFacts = {counts};
        assert.ok(counts.cdp_blocked>=1); assert.ok(!counts.cdp_control_failed);
        return {counts, scope:'attached_page_only'};
      } finally { await ctx.close(); }
    });
    await check('popup_iframe_and_websocket', async () => {
      const {ctx, counts} = await context();
      try {
        const page = await ctx.newPage(); await page.goto(origin + '/page');
        await page.evaluate(async () => {
          window.open('/sink','synthetic-popup');
          const frame=document.createElement('iframe'); frame.src='/sink'; document.body.append(frame);
          await new Promise(resolve => { const w=new WebSocket(location.origin.replace('http:', 'ws:')+'/sink'); w.onerror=resolve; w.onclose=resolve; setTimeout(resolve,1000); });
        });
        await new Promise(resolve => setTimeout(resolve, 200));
        assert.ok(counts.blocked >= 2); assert.equal(counts.websocket_blocked, 1);
        return { counts };
      } finally { await ctx.close(); }
    });
    await check('worker_requests_and_serviceworker_block', async () => {
      const {ctx, counts} = await context();
      try {
        const page = await ctx.newPage(); await page.goto(origin + '/sw-page');
        const facts = await page.evaluate(async () => {
          const registration = await navigator.serviceWorker.register('/sw.js').then(()=>true,()=>false);
          const workerDenied = await new Promise(resolve => {
            const w=new Worker('/worker.js');
            const timer=setTimeout(()=>{w.terminate();resolve(false);},2000);
            w.onmessage=e=>{clearTimeout(timer);w.terminate();resolve(e.data==='denied');};
          });
          return {registration,workerDenied};
        });
        caseFacts = { ...facts, serviceworker_targets:ctx.serviceWorkers().length, serviceworker_script_requested:requests.some(r=>r.name==='/sw.js'), counts };
        assert.equal(ctx.serviceWorkers().length,0);
        assert.equal(requests.some(r=>r.name==='/sw.js'),false);
        assert.equal(facts.workerDenied,true); assert.ok(counts.blocked>=1);
        return { serviceworker_blocked:true, worker_post_blocked:true, registration_promise_resolved: facts.registration, counts };
      } finally { await ctx.close(); }
    });
    await check('page_cdp_popup_redirect_coverage', async () => {
      const {ctx,counts,newPage} = await context(32,true);
      try {
        const page=await newPage(); await page.goto(origin+'/page');
        const popupPromise=page.waitForEvent('popup');
        await page.evaluate(()=>window.open('/redirect','synthetic-redirect'));
        const popup=await popupPromise;
        await popup.waitForLoadState('domcontentloaded').catch(()=>null);
        caseFacts={counts,scope:'popup_without_dedicated_cdp'};
        return caseFacts;
      } finally { await ctx.close(); }
    });
    await check('page_cdp_worker_redirect_coverage', async () => {
      const {ctx,counts,newPage} = await context(32,true);
      try {
        const page=await newPage(); await page.goto(origin+'/page');
        const completed=await page.evaluate(()=>new Promise(resolve=>{
          const w=new Worker('/worker-redirect.js');
          const timer=setTimeout(()=>{w.terminate();resolve(false);},2000);
          w.onmessage=()=>{clearTimeout(timer);w.terminate();resolve(true);};
        }));
        caseFacts={counts,worker_finished:completed,scope:'worker_without_dedicated_cdp'};
        assert.ok(completed);
        return caseFacts;
      } finally { await ctx.close(); }
    });
    await check('synchronous_request_budget', async () => {
      const {ctx, counts} = await context(4);
      try {
        const page = await ctx.newPage(); await page.goto(origin + '/page');
        await page.evaluate(async()=>Promise.all(Array.from({length:12},()=>fetch('/data',{cache:'no-store'}).catch(()=>null))));
        assert.equal(counts.allowed,4); assert.equal(counts.blocked,9); assert.ok(counts.budget_stopped);
        return {counts};
      } finally { await ctx.close(); }
    });
    await check('cdp_source_and_actual_argument_relationship', async () => {
      const {ctx} = await context();
      try {
        const page = await ctx.newPage();
        const cdp = await ctx.newCDPSession(page);
        await cdp.send('Debugger.enable');
        let scriptId;
        cdp.on('Debugger.scriptParsed', e => { if(e.url === origin+'/logic.js') scriptId=e.scriptId; });
        await cdp.send('Debugger.setBreakpointByUrl', { url:origin+'/logic.js',lineNumber:3 });
        await page.goto(origin+'/call');
        assert.ok(scriptId);
        const observed = await cdp.send('Debugger.getScriptSource', {scriptId});
        assert.equal(observed.scriptSource,script);
        const match = /(?:publicKeyStr|publicKey)\s*=\s*["']([A-Za-z0-9+/=]+)["']/.exec(observed.scriptSource);
        assert.equal(match[1],decoy);
        const pause = new Promise((resolve,reject)=>{
          const timer=setTimeout(()=>reject(new Error('pause_timeout')),5000);
          cdp.once('Debugger.paused', e=>{clearTimeout(timer);resolve(e);});
        });
        const click = page.getByRole('button',{name:'Prepare synthetic'}).click();
        const event=await pause;
        assert.equal(event.callFrames[0].functionName,'consumeKey');
        assert.ok(event.callFrames.some(f=>f.functionName==='prepareSynthetic'));
        const evaluation=await cdp.send('Debugger.evaluateOnCallFrame',{
          callFrameId:event.callFrames[0].callFrameId,
          expression:`({active:key===${JSON.stringify(active)},decoy:key===${JSON.stringify(decoy)}})`,
          returnByValue:true, silent:true, throwOnSideEffect:true,
        });
        assert.ok(!evaluation.exceptionDetails);
        assert.deepEqual(evaluation.result.value,{active:true,decoy:false});
        await cdp.send('Debugger.resume'); await click;
        await cdp.send('Debugger.disable');
        return { source_match:true, text_matches_comment:true, actual_argument_matches_active:true, actual_argument_matches_comment:false, synthetic_handler_observed:true };
      } finally { await ctx.close(); }
    });
    report.passed=report.checks.length===10 && report.checks.every(c=>c.passed);
  } catch (_) { report.stop='browser_or_harness_unavailable'; }
  finally {
    if(browser) { await browser.close(); report.browser_closed=true; }
    server.closeAllConnections(); proxy.closeAllConnections();
    await Promise.all([new Promise(resolve=>server.close(resolve)),new Promise(resolve=>proxy.close(resolve))]);
    clearTimeout(hard);
    report.sink_arrivals=requests.filter(r=>r.name==='sink'||r.method!=='GET').length;
    report.websocket_server_arrivals=socketArrivals;
    report.rejected_proxy_requests=deniedProxy;
    report.fixture_get_arrivals=requests.filter(r=>r.name!=='sink'&&r.method==='GET').length;
    report.passed=report.passed&&report.sink_arrivals===0&&socketArrivals===0&&report.browser_closed;
    save();
  }
  console.log(JSON.stringify(report,null,2));
  if(!report.passed) process.exitCode=1;
}
main().catch(()=>{console.error('offline_harness_failed');process.exitCode=2;});
