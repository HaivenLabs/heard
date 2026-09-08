#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { HAIVEN_NODE_VERSION, HAIVEN_PNPM_VERSION, validateSupplyChainPolicy } from '../lib/supply-chain-policy.js';
import { validateDesignPolicy } from '../lib/design-policy.js';

const VERSION = 1;
const COMPASS_VERSION = '1.1.0';
const AGENTS = ['codex','claude-code','claude-desktop','cursor','copilot','antigravity'];
const DOC_FILES = ['HAIVEN_CONSTITUTION.md','PRODUCT_REGISTRY.md','SHARED_SERVICES.md','AGENT_INSTRUCTIONS.md','DESIGN_SYSTEM.md','docs/standards/api-contracts.md','docs/standards/testing.md','docs/standards/observability.md','docs/standards/security.md','docs/standards/local-development.md','docs/products/passage.md','docs/products/qurl.md','docs/products/heard.md'];
const CONSTITUTION_SKILL_PATH = '.haiven/skills/haiven-constitution/SKILL.md';
const CONSTITUTION_HOOK_PATH = '.haiven/hooks/haiven-constitution-check.py';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const bundledDocsRoot = path.resolve(__dirname, '..', 'assets', 'haiven-docs');
const exists = p => fs.existsSync(p);
const read = p => fs.readFileSync(p,'utf8');
const write = (p,c) => { fs.mkdirSync(path.dirname(p),{recursive:true}); fs.writeFileSync(p, c.endsWith('\n') ? c : c+'\n','utf8'); };
const rm = p => { if(exists(p)) fs.rmSync(p,{recursive:true,force:true}); };
const copy = (a,b) => { fs.mkdirSync(path.dirname(b),{recursive:true}); fs.copyFileSync(a,b); };
function walk(root,out=[],ignore=[]){ if(!exists(root)) return out; const n=root.split(path.sep).join('/'); if(ignore.some(x=>n.includes(x))) return out; const st=fs.statSync(root); if(st.isDirectory()) for(const i of fs.readdirSync(root)) walk(path.join(root,i),out,ignore); else if(st.isFile()) out.push(root); return out; }
function rel(p,root){ return path.relative(root,p).split(path.sep).join('/'); }
function parse(argv){ const cmd=argv[0]; const isAgent=cmd==='agent'; const out={cmd,sub:isAgent?argv[1]:undefined,flags:{},rest:[]}; const items=isAgent?argv.slice(2):argv.slice(1); for(let i=0;i<items.length;i++){ const item=items[i]; if(item.startsWith('--')){ const k=item.slice(2), next=items[i+1]; if(!next || next.startsWith('--')) out.flags[k]=true; else { out.flags[k]=next; i++; } } else out.rest.push(item); } return out; }
function findRepoRoot(start=process.cwd()){ let cur=path.resolve(start); while(true){ if(exists(path.join(cur,'.git')) || exists(path.join(cur,'haiven.config.json')) || exists(path.join(cur,'package.json'))) return cur; const p=path.dirname(cur); if(p===cur) return path.resolve(start); cur=p; } }
function resolveSourceInfo(repoRoot, explicit){
  const candidates=[];
  if(explicit) candidates.push({root:path.resolve(repoRoot, explicit), kind:'--source override'});
  else if(process.env.HAIVEN_SOURCE_DIR) candidates.push({root:path.resolve(process.env.HAIVEN_SOURCE_DIR), kind:'HAIVEN_SOURCE_DIR override'});
  if(!explicit && !process.env.HAIVEN_SOURCE_DIR && exists(path.join(repoRoot,'HAIVEN_CONSTITUTION.md'))) candidates.push({root:repoRoot, kind:'repo root docs'});
  candidates.push({root:bundledDocsRoot, kind:'bundled Compass docs'});
  for(const c of candidates) if(exists(path.join(c.root,'HAIVEN_CONSTITUTION.md'))) return c;
  throw new Error('Could not find Haiven source docs. Reinstall @haiven/compass or pass --source /path/to/haiven.');
}
function resolveSource(repoRoot, explicit){ return resolveSourceInfo(repoRoot, explicit).root; }
function inferProduct(repoRoot){
  const pkgPath=path.join(repoRoot,'package.json');
  if(exists(pkgPath)){
    try{
      const pkg=JSON.parse(read(pkgPath));
      if(pkg.name && typeof pkg.name==='string'){
        const name=pkg.name.split('/').pop().trim();
        if(name) return name;
      }
    }catch{}
  }
  const folder=path.basename(repoRoot).trim();
  if(folder) return folder;
  throw new Error('Could not infer product name. Pass --product <name>.');
}
function hashDir(root){ const files=walk(root,[],['node_modules','.git','dist']).sort(); const text=files.map(f=>rel(f,root)+'\n'+read(f)).join('\n---FILE---\n'); return crypto.createHash('sha256').update(text).digest('hex'); }
function readConfig(repoRoot){ const p=path.join(repoRoot,'haiven.config.json'); if(!exists(p)) throw new Error('Missing haiven.config.json. Run: haiven init'); return JSON.parse(read(p)); }
function writeConfig(repoRoot,cfg){ write(path.join(repoRoot,'haiven.config.json'), JSON.stringify(cfg,null,2)); }
function ensureEnvironmentIgnore(repoRoot){
  const file=path.join(repoRoot,'.gitignore');
  const current=exists(file) ? read(file) : '';
  const required=['.env','.env.*','!.env.example'];
  const existing=new Set(current.split(/\r?\n/).map(line=>line.trim()).filter(Boolean));
  const missing=required.filter(rule=>!existing.has(rule));
  if(!missing.length) return;
  const prefix=current && !current.endsWith('\n') ? '\n' : '';
  const heading=existing.has('# Environment files') ? '' : '# Environment files\n';
  write(file, `${current}${prefix}${heading}${missing.join('\n')}\n`);
}
function gitResult(repoRoot,args){ return spawnSync('git',args,{cwd:repoRoot,stdio:'ignore'}).status===0; }
function checkEnvironmentFile(repoRoot,relative){
  if(gitResult(repoRoot,['ls-files','--error-unmatch','--',relative])) return 'tracked';
  if(gitResult(repoRoot,['check-ignore','-q','--no-index','--',relative])) return 'ignored';
  return 'unignored';
}
function environmentFiles(root){
  return walk(root,[],['node_modules','.git','.cache','dist','.next','build','coverage','.haiven'])
    .filter(file=>{ const base=path.basename(file); return (base==='.env' || base.startsWith('.env.')) && base!=='.env.example'; });
}
function banner(){ return '<!--\nGenerated by Haiven Compass.\nDo not edit directly.\nSource: .haiven/HAIVEN_CONSTITUTION.md and .haiven/AGENT_INSTRUCTIONS.md\nRegenerate with: haiven sync\n-->\n\n'; }
function shBanner(){ return '# Generated by Haiven Compass.\n# Do not edit directly.\n# Regenerate with: haiven sync.\n\n'; }
function defaultOptions(){ return {hooks:{constitution:true},skills:{constitution:true},plugins:{}}; }
function ensureOptions(cfg){
  cfg.options=cfg.options||{};
  cfg.options.hooks=cfg.options.hooks||{};
  cfg.options.skills=cfg.options.skills||{};
  cfg.options.plugins=cfg.options.plugins||{};
  if(cfg.options.hooks.constitution===undefined) cfg.options.hooks.constitution=true;
  if(cfg.options.skills.constitution===undefined) cfg.options.skills.constitution=true;
  return cfg;
}
function applyRuntimeFlags(cfg,flags={}){
  ensureOptions(cfg);
  if(flags.hooks) cfg.options.hooks.constitution=true;
  if(flags.skills) cfg.options.skills.constitution=true;
  if(flags.plugins && cfg.options.plugins.constitution===undefined) cfg.options.plugins.constitution=true;
  return cfg;
}
function neutralAgent(cfg){ return `${banner()}# Haiven Agent Instructions\n\nThis repository is part of the Haiven ecosystem.\n\nProduct: \`${cfg.product}\`\n\nBefore making changes, read:\n\n- \`.haiven/HAIVEN_CONSTITUTION.md\`\n- \`.haiven/AGENT_INSTRUCTIONS.md\`\n- \`.haiven/SHARED_SERVICES.md\`\n- \`.haiven/PRODUCT_REGISTRY.md\`\n- \`${CONSTITUTION_SKILL_PATH}\`\n- product-specific docs and active implementation plans\n\nNon-negotiable rules:\n\n- API-first.\n- Contract-first.\n- Test-first.\n- Modular monolith first unless explicitly documented otherwise.\n- Build vertical slices, not disconnected scaffolding.\n- Use Passage for shared auth/identity.\n- Use qurl for QR generation.\n- Do not bypass tenant, role, permission, security, or software supply-chain boundaries.\n- Preserve pinned package managers and lockfiles, deny unreviewed install scripts, pin CI actions to full SHAs, and keep untrusted code away from credentials.\n- Do not ship developer-demo UI for production surfaces.\n- Update durable docs when product, architecture, UX, or quality rules change.\n\nBefore marking work complete, run:\n\n\`\`\`bash\nhaiven check\n\`\`\`\n\nIf \`haiven\` is not globally available, use the repo script:\n\n\`\`\`bash\npnpm haiven:check\n\`\`\`\n`; }
function claudeMd(cfg){ return `${banner()}@AGENTS.md\n\n## Claude Code\n\nUse the Haiven rules in this repository before planning, editing, testing, or opening a PR.\n\nProduct: \`${cfg.product}\`\n\nBefore finishing implementation work, run:\n\n\`\`\`bash\npnpm haiven:check\n\`\`\`\n`; }
function constitutionSkill(cfg){ return `${banner()}# Haiven Constitution Skill\n\nUse this skill for any implementation task in \`${cfg.product}\`.\n\n## 1) Classify the change first\n\nPick one or more:\n\n- \`domain logic\`\n- \`API contract\`\n- \`workflow\`\n- \`eventing\`\n- \`auth/permissions\`\n- \`tenant isolation\`\n- \`provider integration\`\n- \`UI/UX\`\n\n## 2) Build test-first, not test-later\n\nTests are required and should be planned before code.\n\nMinimum expected coverage depends on change scope:\n\n- Unit tests for deterministic logic and validation.\n- Integration tests for service, database, provider, and workflow behavior.\n- Contract tests when API/event contracts change.\n- Permission and tenant-isolation tests for protected behavior.\n- Failure-mode tests for critical workflows.\n\n## 3) Use BDD for critical workflows\n\nWhen a workflow is multi-step, permission-sensitive, user-facing, provider-backed, or event-driven, include scenario-style tests:\n\n- Given known system state\n- When user or system action occurs\n- Then business outcome is correct\n- And permissions, audit visibility, and event behavior are enforced\n\nDo not force heavyweight BDD for small deterministic utilities.\n\n## 4) Respect shared-service boundaries\n\n- Use Passage for shared identity/session/account behavior.\n- Use qurl for QR generation and rendering.\n- Temporary local adapters are only for documented fake/test/offline use.\n\n## 5) Done means done\n\nBefore completion, verify all applicable items:\n\n- Contracts updated first when behavior changes.\n- Tests added and passing.\n- Permissions and tenant boundaries enforced.\n- Failure handling and observability included.\n- Docs updated for durable decisions.\n- Supply-chain policy, dependency audit, immutable CI action, least-privilege credential, SBOM, and provenance checks pass.\n`; }
function constitutionHook(){ return `#!/usr/bin/env python3\n# Generated by Haiven Compass.\n# Do not edit directly.\n\nimport shutil\nimport subprocess\nimport sys\nfrom pathlib import Path\n\nCODE_EXTENSIONS = {\".ts\", \".tsx\", \".js\", \".jsx\", \".go\", \".py\", \".cs\", \".java\", \".rb\", \".php\", \".rs\"}\nTEST_MARKERS = (\"/test/\", \"/tests/\", \"__tests__\", \".test.\", \".spec.\")\nAPI_MARKERS = (\"/api/\", \"controller\", \"handler\", \"route\")\nCONTRACT_MARKERS = (\"openapi\", \"contract\", \"schema\", \"swagger\")\n\n\ndef run_and_capture(cmd, cwd):\n    try:\n        res = subprocess.run(cmd, cwd=cwd, check=False, text=True, capture_output=True)\n        return res.returncode, res.stdout\n    except Exception:\n        return 1, \"\"\n\n\ndef changed_files(root):\n    commands = [\n        [\"git\", \"diff\", \"--name-only\", \"--cached\"],\n        [\"git\", \"diff\", \"--name-only\"],\n    ]\n    files = []\n    seen = set()\n    for cmd in commands:\n        code, out = run_and_capture(cmd, root)\n        if code != 0:\n            continue\n        for line in out.splitlines():\n            item = line.strip().replace(\"\\\\\", \"/\")\n            if item and item not in seen:\n                seen.add(item)\n                files.append(item)\n    return files\n\n\ndef looks_like_test(path):\n    lower = path.lower()\n    if any(marker in lower for marker in TEST_MARKERS):\n        return True\n    base = Path(lower).name\n    return base.startswith(\"test_\")\n\n\ndef looks_like_code(path):\n    return Path(path).suffix.lower() in CODE_EXTENSIONS\n\n\ndef looks_like_api(path):\n    lower = path.lower()\n    return any(marker in lower for marker in API_MARKERS)\n\n\ndef looks_like_contract(path):\n    lower = path.lower()\n    return any(marker in lower for marker in CONTRACT_MARKERS)\n\n\ndef run_haiven_check(root):\n    if (root / \"package.json\").exists() and shutil.which(\"pnpm\"):\n        code, _ = run_and_capture([\"pnpm\", \"haiven:check\"], root)\n        if code == 0:\n            return 0\n    if shutil.which(\"haiven\"):\n        code, _ = run_and_capture([\"haiven\", \"check\"], root)\n        return code\n    local_compass = root / \"packages\" / \"compass\" / \"bin\" / \"haiven.js\"\n    if local_compass.exists() and shutil.which(\"node\"):\n        code, _ = run_and_capture([\"node\", str(local_compass), \"check\"], root)\n        return code\n    print(\"haiven check not found. Install @haiven/compass or add pnpm haiven:check.\")\n    return 0\n\n\ndef main():\n    root = Path.cwd()\n    files = changed_files(root)\n    code_files = [f for f in files if looks_like_code(f)]\n    test_files = [f for f in files if looks_like_test(f)]\n    api_files = [f for f in code_files if looks_like_api(f)]\n    contract_files = [f for f in files if looks_like_contract(f)]\n\n    print(\"[haiven] constitution hook\")\n    print(f\"[haiven] changed files detected: {len(files)}\")\n\n    failures = []\n    notices = []\n\n    if code_files and not test_files:\n        failures.append(\"Code changed without test changes. Haiven Constitution section 6 requires test-first delivery.\")\n\n    if api_files and not contract_files:\n        notices.append(\"API-like files changed without obvious contract file updates. Confirm section 4 (contract-first).\")\n\n    if notices:\n        print(\"[haiven] notices:\")\n        for item in notices:\n            print(f\"- {item}\")\n\n    if failures:\n        print(\"[haiven] failures:\")\n        for item in failures:\n            print(f\"- {item}\")\n        return 1\n\n    return run_haiven_check(root)\n\n\nif __name__ == \"__main__\":\n    sys.exit(main())\n`; }
function claudeHook(){ return `${shBanner()}set -euo pipefail\n\nif [ -f \"${CONSTITUTION_HOOK_PATH}\" ]; then\n  if command -v python3 >/dev/null 2>&1; then\n    python3 \"${CONSTITUTION_HOOK_PATH}\"\n    exit $?\n  elif command -v python >/dev/null 2>&1; then\n    python \"${CONSTITUTION_HOOK_PATH}\"\n    exit $?\n  fi\nfi\n\nif command -v pnpm >/dev/null 2>&1 && [ -f package.json ]; then\n  pnpm haiven:check\nelif command -v haiven >/dev/null 2>&1; then\n  haiven check\nelse\n  echo \"haiven check not found. Install @haiven/compass or add pnpm haiven:check.\"\nfi\n`; }
function codexHook(){ return `#!/usr/bin/env python3\n# Generated by Haiven Compass.\n# Do not edit directly.\n\nimport shutil\nimport subprocess\nimport sys\nfrom pathlib import Path\n\nroot = Path.cwd()\nshared_hook = root / \"${CONSTITUTION_HOOK_PATH}\"\nif shared_hook.exists() and shared_hook.resolve() != Path(__file__).resolve():\n    sys.exit(subprocess.run([sys.executable, str(shared_hook)], cwd=root).returncode)\n\nif (root / \"package.json\").exists() and shutil.which(\"pnpm\"):\n    subprocess.run([\"pnpm\", \"haiven:check\"], check=True)\nelif shutil.which(\"haiven\"):\n    subprocess.run([\"haiven\", \"check\"], check=True)\nelse:\n    print(\"haiven check not found. Install @haiven/compass or add pnpm haiven:check.\")\n`; }
function environmentPolicySkill(){ return `\n## 5) Protect environment files\n\n- Keep actual environment files out of Git with \`.env\`, \`.env.*\`, and \`!.env.example\` in \`.gitignore\`.\n- Commit only sanitized placeholders in \`.env.example\`.\n- If an actual environment file was tracked, remove it from Git tracking and rotate any exposed credentials.\n`; }
function browserOriginPolicySkill(){ return `\n## 6) Keep browser identity origins canonical\n\n- Configure one canonical public browser origin per environment, including local, integration, staging, and production.\n- Redirect aliases before issuing OAuth, session, CSRF, PKCE, or state cookies.\n- Derive or validate OAuth callbacks, CORS, and branded return URLs against that configured origin; never use an untrusted request host to select a callback.\n- Require HTTPS and explicit configuration outside local runtimes, and add regression tests for origin/callback drift.\n`; }
function writeConstitutionRuntime(repoRoot,cfg){
  ensureOptions(cfg);
  if(cfg.options.skills.constitution) write(path.join(repoRoot,CONSTITUTION_SKILL_PATH), constitutionSkill(cfg).replace('\n## 5) Done means done', `${environmentPolicySkill()}${browserOriginPolicySkill()}\n## 7) Done means done`));
  if(cfg.options.hooks.constitution) write(path.join(repoRoot,CONSTITUTION_HOOK_PATH), constitutionHook());
  copy(__filename,path.join(repoRoot,'.haiven/bin/haiven.js'));
  copy(path.resolve(__dirname,'..','lib','supply-chain-policy.js'),path.join(repoRoot,'.haiven/lib/supply-chain-policy.js'));
}
function adapterFiles(agent){ return ({'codex':['AGENTS.md','.codex/config.toml','.codex/rules/haiven.rules','.codex/hooks/haiven-check.py'],'claude-code':['AGENTS.md','CLAUDE.md','.claude/rules/haiven.md','.claude/rules/api-contracts.md','.claude/rules/testing.md','.claude/rules/security.md','.claude/hooks/haiven-check.sh'],'claude-desktop':['AGENTS.md','CLAUDE_DESKTOP.md','.haiven/adapters/claude-desktop/skill/SKILL.md'],'cursor':['AGENTS.md','.cursor/rules/haiven.mdc'],'copilot':['AGENTS.md','.github/copilot-instructions.md','.github/instructions/haiven.instructions.md','.github/instructions/api-contracts.instructions.md','.github/instructions/testing.instructions.md'],'antigravity':['AGENTS.md','.antigravity/rules/haiven.md','.antigravity/hooks/haiven-check.sh']})[agent] || []; }
function enableAdapter(repoRoot,cfg,agent){
  if(!AGENTS.includes(agent)) throw new Error(`Unsupported agent: ${agent}`);
  writeConstitutionRuntime(repoRoot,cfg);
  write(path.join(repoRoot,'AGENTS.md'), neutralAgent(cfg));
  if(agent==='codex'){ write(path.join(repoRoot,'.codex/config.toml'), `# Generated by Haiven Compass.\napproval_policy = "on-request"\nsandbox_mode = "workspace-write"\n\n[haiven]\nenabled = true\ncheck_command = "pnpm haiven:check"\n`); write(path.join(repoRoot,'.codex/rules/haiven.rules'), `${shBanner()}# Haiven Codex Rules\n\nProduct: ${cfg.product}\n\nRead .haiven/HAIVEN_CONSTITUTION.md, ${CONSTITUTION_SKILL_PATH}, and AGENTS.md before making changes. Do not bypass contracts. Use Passage for shared identity. Use qurl for QR generation. Build vertical slices. Run haiven check before completion.\n`); write(path.join(repoRoot,'.codex/hooks/haiven-check.py'), codexHook()); }
  if(agent==='claude-code'){ write(path.join(repoRoot,'CLAUDE.md'), claudeMd(cfg)); write(path.join(repoRoot,'.claude/rules/haiven.md'), `${banner()}# Haiven Rules\n\nFollow the Haiven Constitution and product-specific docs. Build vertical slices, preserve contracts, and run checks before completion.\n\nAlso apply: \`${CONSTITUTION_SKILL_PATH}\`\n`); write(path.join(repoRoot,'.claude/rules/api-contracts.md'), `${banner()}# API and Contract Rules\n\nUpdate OpenAPI and JSON Schema contracts before implementation changes. Do not add undocumented API behavior.\n`); write(path.join(repoRoot,'.claude/rules/testing.md'), `${banner()}# Testing Rules\n\nEvery meaningful change needs appropriate tests. Do not leave test work as a TODO.\n`); write(path.join(repoRoot,'.claude/rules/security.md'), `${banner()}# Security Rules\n\nUse Passage for shared auth. Enforce tenant, role, permission, and privacy boundaries. Never commit secrets.\n`); write(path.join(repoRoot,'.claude/hooks/haiven-check.sh'), claudeHook()); }
  if(agent==='claude-desktop'){ const skill=`${banner()}# Haiven Repository Skill\n\nUse this skill when working with the \`${cfg.product}\` repository or any Haiven product repository.\n\nRead .haiven/HAIVEN_CONSTITUTION.md, .haiven/AGENT_INSTRUCTIONS.md, .haiven/SHARED_SERVICES.md, AGENTS.md, and ${CONSTITUTION_SKILL_PATH}.\n\nRun haiven check before declaring completion.\n`; write(path.join(repoRoot,'CLAUDE_DESKTOP.md'), skill); write(path.join(repoRoot,'.haiven/adapters/claude-desktop/skill/SKILL.md'), skill); }
  if(agent==='cursor'){ write(path.join(repoRoot,'.cursor/rules/haiven.mdc'), `${banner()}---\ndescription: Haiven ecosystem rules for ${cfg.product}\nalwaysApply: true\n---\n\n# Haiven Cursor Rules\n\nRead .haiven/HAIVEN_CONSTITUTION.md, .haiven/AGENT_INSTRUCTIONS.md, .haiven/SHARED_SERVICES.md, ${CONSTITUTION_SKILL_PATH}, and AGENTS.md if present.\n\nRules: API-first. Contract-first. Test-first. Build vertical slices. Use Passage for identity. Use qurl for QR. Run haiven check before completion.\n`); }
  if(agent==='copilot'){ write(path.join(repoRoot,'.github/copilot-instructions.md'), `${banner()}# GitHub Copilot Instructions\n\nThis repository is part of Haiven. Product: \`${cfg.product}\`. Follow .haiven/HAIVEN_CONSTITUTION.md, ${CONSTITUTION_SKILL_PATH}, AGENTS.md, and product-specific docs. Use Passage for shared identity. Use qurl for QR generation. Build vertical slices.\n`); write(path.join(repoRoot,'.github/instructions/haiven.instructions.md'), `${banner()}---\napplyTo: \"**/*\"\n---\n\n# Haiven Instructions\n\nFollow the Haiven Constitution. Use Passage for shared auth and qurl for QR generation. Build vertical slices.\n`); write(path.join(repoRoot,'.github/instructions/api-contracts.instructions.md'), `${banner()}---\napplyTo: \"**/*.{ts,tsx,go,cs,py,yaml,json}\"\n---\n\n# API Contract Instructions\n\nWhen API behavior changes, update OpenAPI and JSON Schema contracts before implementation.\n`); write(path.join(repoRoot,'.github/instructions/testing.instructions.md'), `${banner()}---\napplyTo: \"**/*.{ts,tsx,go,cs,py}\"\n---\n\n# Testing Instructions\n\nEvery meaningful change needs relevant tests. Include permission, tenant isolation, provider, and failure-mode tests where applicable.\n`); }
  if(agent==='antigravity'){ write(path.join(repoRoot,'.antigravity/rules/haiven.md'), `${banner()}# Haiven Antigravity Rules\n\nProduct: \`${cfg.product}\`\n\nRead .haiven/HAIVEN_CONSTITUTION.md, ${CONSTITUTION_SKILL_PATH}, AGENTS.md, and product-specific docs. API-first. Contract-first. Test-first. Use Passage for identity. Use qurl for QR. Run haiven check.\n\nNote: Antigravity adapter paths may need adjustment as its local configuration conventions stabilize.\n`); write(path.join(repoRoot,'.antigravity/hooks/haiven-check.sh'), claudeHook()); }
}
function disableAdapter(repoRoot,agent){ for(const f of adapterFiles(agent)) rm(path.join(repoRoot,f)); }

function detectAgents(repoRoot){
  const detections=[];
  const add=(agent, reason)=>{ if(!detections.some(d=>d.agent===agent)) detections.push({agent, reason}); };
  if(exists(path.join(repoRoot,'.cursor')) || exists(path.join(repoRoot,'.cursorrules'))) add('cursor','Cursor repo files detected');
  if(exists(path.join(repoRoot,'.codex'))) add('codex','Codex repo files detected');
  if(exists(path.join(repoRoot,'.claude')) || exists(path.join(repoRoot,'CLAUDE.md'))) add('claude-code','Claude Code repo files detected');
  if(exists(path.join(repoRoot,'CLAUDE_DESKTOP.md')) || exists(path.join(repoRoot,'.haiven/adapters/claude-desktop'))) add('claude-desktop','Claude Desktop repo files detected');
  if(exists(path.join(repoRoot,'.github/copilot-instructions.md')) || exists(path.join(repoRoot,'.github/instructions'))) add('copilot','GitHub Copilot repo files detected');
  if(exists(path.join(repoRoot,'.antigravity'))) add('antigravity','Antigravity repo files detected');
  return detections;
}
function parseAgentList(value){
  if(!value || value===true) return [];
  if(value==='auto') return ['auto'];
  return String(value).split(',').map(x=>x.trim()).filter(Boolean);
}
function agentsRequested(flags){
  const agents=[];
  for(const key of ['agent','agents']) for(const a of parseAgentList(flags[key])) agents.push(a);
  return [...new Set(agents)];
}
function enableAdapters(repoRoot,cfg,agents){
  const enabled=[];
  cfg.enabledAgents=cfg.enabledAgents||{};
  for(const agent of agents){
    if(!AGENTS.includes(agent)) throw new Error(`Unsupported agent: ${agent}. Use one of: ${AGENTS.join(', ')}, or auto.`);
    cfg.enabledAgents[agent]=true;
    enableAdapter(repoRoot,cfg,agent);
    enabled.push(agent);
  }
  return enabled;
}
function enableDetectedAgents(repoRoot,cfg){
  const detections=detectAgents(repoRoot);
  const toEnable=detections.map(d=>d.agent).filter(a=>!cfg.enabledAgents?.[a]);
  const enabled=enableAdapters(repoRoot,cfg,toEnable);
  return {detections, enabled};
}
function printDetected(detections){
  if(!detections.length){ console.log('Detected repo tools: none'); return; }
  console.log('Detected repo tools:');
  for(const d of detections) console.log(`- ${d.agent}: ${d.reason}`);
}
function copyDocs(source,repoRoot){ const local=path.join(repoRoot,'.haiven'); fs.mkdirSync(local,{recursive:true}); for(const f of DOC_FILES){ const from=path.join(source,f); if(exists(from)) copy(from,path.join(local,f)); } write(path.join(local,'manifest.json'), JSON.stringify({generatedBy:'haiven-compass',sourceRoot:source,docsHash:hashDir(local),generatedAt:new Date().toISOString()},null,2)); }
function upsertScripts(repoRoot){ const p=path.join(repoRoot,'package.json'); if(!exists(p)) return; const pkg=JSON.parse(read(p)); pkg.packageManager=`pnpm@${HAIVEN_PNPM_VERSION}`; pkg.engines=pkg.engines||{}; pkg.engines.node=`>=${HAIVEN_NODE_VERSION} <23`; pkg.scripts=pkg.scripts||{}; pkg.scripts['haiven:sync']=pkg.scripts['haiven:sync']||'haiven sync'; pkg.scripts['haiven:check']='node .haiven/bin/haiven.js check'; pkg.scripts['haiven:doctor']=pkg.scripts['haiven:doctor']||'haiven doctor'; pkg.scripts['audit:dependencies']=pkg.scripts['audit:dependencies']||'pnpm audit --audit-level=high'; write(p,JSON.stringify(pkg,null,2)); write(path.join(repoRoot,'.node-version'),`${HAIVEN_NODE_VERSION}\n`); }
function ensurePackagePolicy(repoRoot){ const p=path.join(repoRoot,'pnpm-workspace.yaml'); if(exists(p)) return; write(p, `packages:\n  - apps/*\n  - packages/*\n\nminimumReleaseAge: 1440\nminimumReleaseAgeStrict: true\nminimumReleaseAgeIgnoreMissingTime: false\ntrustPolicy: no-downgrade\ntrustPolicyIgnoreAfter: 525600\ntrustLockfile: false\nblockExoticSubdeps: true\nstrictDepBuilds: true\nallowBuilds: {}\n`); }
function ensureWorkflow(repoRoot){
  const workflowRoot=path.join(repoRoot,'.github/workflows');
  const p=path.join(workflowRoot,'haiven-standards.yml');
  if(!exists(p)) write(p, `name: Haiven Standards\n\non:\n  pull_request:\n  push:\n    branches: [main]\n\npermissions:\n  contents: read\n\njobs:\n  quality:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n    steps:\n      - uses: actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803 # v6\n        with:\n          persist-credentials: false\n      - uses: pnpm/action-setup@f40ffcd9367d9f12939873eb1018b921a783ffaa # v4\n        with:\n          run_install: false\n      - uses: actions/setup-node@249970729cb0ef3589644e2896645e5dc5ba9c38 # v6\n        with:\n          node-version: 22.23.2\n          cache: pnpm\n      - run: pnpm install --frozen-lockfile\n      - run: pnpm audit --audit-level=high\n      - run: pnpm haiven:check\n      - uses: anchore/sbom-action@e22c389904149dbc22b58101806040fa8d37a610 # v0\n        with:\n          path: .\n          format: cyclonedx-json\n          output-file: haiven-sbom.cdx.json\n          upload-artifact: false\n      - uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02 # v4\n        with:\n          name: haiven-sbom\n          path: haiven-sbom.cdx.json\n          if-no-files-found: error\n\n  dependency-review:\n    if: github.event_name == 'pull_request' && (github.event.repository.private == false || vars.ENABLE_GITHUB_ADVANCED_SECURITY == 'true')\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803 # v6\n        with:\n          persist-credentials: false\n      - uses: actions/dependency-review-action@1e69f48acb82d1966a394da916b4c1698aa569d6 # v4\n        with:\n          fail-on-severity: moderate\n`);
  const codeql=path.join(workflowRoot,'codeql.yml');
  if(!exists(codeql)) write(codeql, `name: CodeQL\n\non:\n  pull_request:\n  push:\n    branches: [main]\n  schedule:\n    - cron: '17 4 * * 1'\n\npermissions:\n  contents: read\n\njobs:\n  analyze:\n    if: github.event.repository.private == false || vars.ENABLE_GITHUB_ADVANCED_SECURITY == 'true'\n    runs-on: ubuntu-latest\n    permissions:\n      contents: read\n      packages: read\n      security-events: write\n    steps:\n      - uses: actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803 # v6\n        with:\n          persist-credentials: false\n      - uses: github/codeql-action/init@988661ebb5e81487b3fb31b2185d2856c0a10679 # v4\n        with:\n          languages: javascript-typescript\n      - uses: github/codeql-action/analyze@988661ebb5e81487b3fb31b2185d2856c0a10679 # v4\n`);
  const dependabot=path.join(repoRoot,'.github/dependabot.yml');
  if(!exists(dependabot)) write(dependabot, `version: 2\nupdates:\n  - package-ecosystem: npm\n    directory: /\n    schedule:\n      interval: weekly\n  - package-ecosystem: github-actions\n    directory: /\n    schedule:\n      interval: weekly\n`);
}

function init(repoRoot,flags){
  const product=(flags.product && flags.product!==true) ? flags.product : inferProduct(repoRoot);
  const sourceInfo=resolveSourceInfo(repoRoot,flags.source);
  copyDocs(sourceInfo.root,repoRoot);
  ensureEnvironmentIgnore(repoRoot);
  const cfgPath=path.join(repoRoot,'haiven.config.json');
  let cfg;
  if(!exists(cfgPath)){
    cfg={product,haivenVersion:VERSION,constitutionVersion:VERSION,docsSource:sourceInfo.kind,source:flags.source,enabledAgents:{},options:defaultOptions()};
  }else{
    cfg=readConfig(repoRoot);
    cfg.product=cfg.product || product;
    cfg.docsSource=sourceInfo.kind;
  }
  applyRuntimeFlags(cfg,flags);
  writeConstitutionRuntime(repoRoot,cfg);
  const requested=agentsRequested(flags);
  const explicitAgents=requested.includes('auto') ? [] : requested;
  const explicitlyEnabled=enableAdapters(repoRoot,cfg,explicitAgents);
  const autoResult=requested.includes('auto') || requested.length===0 ? enableDetectedAgents(repoRoot,cfg) : {detections:detectAgents(repoRoot), enabled:[]};
  writeConfig(repoRoot,cfg);
  upsertScripts(repoRoot);
  ensurePackagePolicy(repoRoot);
  if(!flags['no-workflow']) ensureWorkflow(repoRoot);
  console.log('Initialized Haiven repo');
  console.log(`Product inferred: ${product}`);
  console.log(`Docs source: ${sourceInfo.kind}` + (sourceInfo.kind.includes('override') ? ` (${sourceInfo.root})` : ''));
  printDetected(autoResult.detections);
  const enabled=[...new Set([...explicitlyEnabled,...autoResult.enabled])];
  console.log(`Enabled adapters: ${enabled.length ? enabled.join(', ') : 'none'}`);
}
function sync(repoRoot,flags){
  const cfg=readConfig(repoRoot);
  const sourceInfo=resolveSourceInfo(repoRoot,flags.source || cfg.source);
  copyDocs(sourceInfo.root,repoRoot);
  ensureEnvironmentIgnore(repoRoot);
  applyRuntimeFlags(cfg,flags);
  writeConstitutionRuntime(repoRoot,cfg);
  cfg.docsSource=sourceInfo.kind;
  for(const [a,on] of Object.entries(cfg.enabledAgents||{})) if(on) enableAdapter(repoRoot,cfg,a);
  writeConfig(repoRoot,cfg);
  console.log('Haiven docs synced.');
  console.log(`Docs source: ${sourceInfo.kind}` + (sourceInfo.kind.includes('override') ? ` (${sourceInfo.root})` : ''));
}
function agent(repoRoot,sub,rest,flags){
  const cfg=readConfig(repoRoot);
  ensureOptions(cfg);
  if(sub==='list'){
    for(const a of AGENTS) console.log(`- ${a}: ${cfg.enabledAgents?.[a]?'enabled':'disabled'}`);
    console.log('');
    console.log(`- constitution skill: ${cfg.options?.skills?.constitution ? 'enabled' : 'disabled'}`);
    console.log(`- constitution hook: ${cfg.options?.hooks?.constitution ? 'enabled' : 'disabled'}`);
    return;
  }
  const name=rest[0];
  if(!AGENTS.includes(name)) throw new Error(`Unsupported agent. Use one of: ${AGENTS.join(', ')}`);
  cfg.enabledAgents=cfg.enabledAgents||{};
  applyRuntimeFlags(cfg,flags);
  if(sub==='enable'){
    cfg.enabledAgents[name]=true;
    writeConstitutionRuntime(repoRoot,cfg);
    enableAdapter(repoRoot,cfg,name);
    writeConfig(repoRoot,cfg);
    console.log(`Enabled Haiven adapter: ${name}`);
    return;
  }
  if(sub==='disable'){
    cfg.enabledAgents[name]=false;
    disableAdapter(repoRoot,name);
    writeConstitutionRuntime(repoRoot,cfg);
    for(const [a,on] of Object.entries(cfg.enabledAgents||{})) if(on) enableAdapter(repoRoot,cfg,a);
    writeConfig(repoRoot,cfg);
    console.log(`Disabled Haiven adapter: ${name}`);
    return;
  }
  throw new Error('Use: haiven agent <enable|disable|list> [agent]');
}
function check(repoRoot){
  const errors=[], warnings=[];
  let cfg;
  try{ cfg=readConfig(repoRoot); }catch(e){ console.error(e.message); process.exitCode=1; return; }
  ensureOptions(cfg);
  for(const f of ['.haiven/HAIVEN_CONSTITUTION.md','.haiven/AGENT_INSTRUCTIONS.md','.haiven/SHARED_SERVICES.md','.haiven/PRODUCT_REGISTRY.md']) if(!exists(path.join(repoRoot,f))) errors.push(`Missing required Haiven doc: ${f}`);
  if(cfg.options?.skills?.constitution && !exists(path.join(repoRoot,CONSTITUTION_SKILL_PATH))) errors.push(`Missing required constitution skill file: ${CONSTITUTION_SKILL_PATH}`);
  if(cfg.options?.hooks?.constitution && !exists(path.join(repoRoot,CONSTITUTION_HOOK_PATH))) errors.push(`Missing required constitution hook file: ${CONSTITUTION_HOOK_PATH}`);
  for(const [a,on] of Object.entries(cfg.enabledAgents||{})) if(on) for(const f of adapterFiles(a)) if(!exists(path.join(repoRoot,f))) errors.push(`Enabled agent '${a}' is missing generated file: ${f}`);
  for(const file of environmentFiles(repoRoot)){
    const relative=rel(file,repoRoot);
    const state=checkEnvironmentFile(repoRoot,relative);
    if(state==='tracked') errors.push(`Environment file is tracked by Git: ${relative}`);
    else if(state==='unignored') errors.push(`Environment file is not ignored by Git: ${relative}`);
  }
  const pkgPath=path.join(repoRoot,'package.json');
  if(exists(pkgPath)){
    const pkg=JSON.parse(read(pkgPath));
    const deps=Object.assign({},pkg.dependencies,pkg.devDependencies,pkg.peerDependencies,pkg.optionalDependencies);
    if(!pkg.scripts?.['haiven:check']) warnings.push('package.json is missing script: haiven:check');
    if(cfg.product!=='qurl') for(const d of ['qrcode','qr-code-styling','react-qr-code','qr-image','awesome-qr']) if(deps[d]) errors.push(`Non-qurl product depends on QR package '${d}'. Use qurl or a documented fake/test provider.`);
    if(cfg.product!=='passage') for(const d of ['next-auth','@auth/core','passport','lucia','better-auth','auth0','@auth0/nextjs-auth0']) if(deps[d]) warnings.push(`Non-Passage product depends on auth package '${d}'. Confirm this is only a Passage adapter or temporary fake.`);
  }
  const supplyChain = validateSupplyChainPolicy(repoRoot);
  warnings.push(...supplyChain.warnings);
  errors.push(...supplyChain.errors);
  const files=walk(repoRoot,[],['node_modules','.git','.cache','dist','.next','build','coverage','.haiven']); for(const f of files){ const r=rel(f,repoRoot), base=path.basename(f); if(/migrations?\//i.test(r) && !/^(\d{3,}|\d{4}[_-]\d{2}[_-]\d{2}|v\d+)[_-].+\.(sql|ts|js|go)$/i.test(base)) warnings.push(`Migration file may not be versioned: ${r}`); if(/\.(ts|tsx|js|jsx|go|py|cs|java)$/i.test(f)){ const c=read(f); if(/console\.log\([^)]*(password|secret|token|apiKey|api_key)/i.test(c)) errors.push(`Possible sensitive logging found: ${r}`); if(cfg.product!=='qurl' && /(from\s+["']qrcode["']|require\(["']qrcode["']\)|QRCodeCanvas|QRCodeSVG)/.test(c)) errors.push(`Possible local QR rendering in non-qurl product: ${r}`); if(cfg.product!=='passage' && /(createSession|setSession|sessionCookie|passport\.use|NextAuth\()/i.test(c)) warnings.push(`Possible local auth/session implementation outside Passage: ${r}`); } }
  const design=validateDesignPolicy(repoRoot,files);
  warnings.push(...design.warnings);
  errors.push(...design.errors);
  if(warnings.length){ console.log('Haiven warnings:'); warnings.forEach(w=>console.log('- '+w)); } if(errors.length){ console.error('Haiven errors:'); errors.forEach(e=>console.error('- '+e)); process.exitCode=1; } else console.log('Haiven check passed.'); }

function doctor(repoRoot,flags={}){
  let cfg;
  let hasConfig=true;
  try{ cfg=readConfig(repoRoot); ensureOptions(cfg); }catch{ hasConfig=false; cfg={product:inferProduct(repoRoot),enabledAgents:{},options:defaultOptions()}; }
  const detections=detectAgents(repoRoot);
  if(flags.fix){
    if(!hasConfig){
      const sourceInfo=resolveSourceInfo(repoRoot,flags.source);
      copyDocs(sourceInfo.root,repoRoot);
      cfg={product:cfg.product,haivenVersion:VERSION,constitutionVersion:VERSION,docsSource:sourceInfo.kind,source:flags.source,enabledAgents:{},options:defaultOptions()};
      applyRuntimeFlags(cfg,flags);
      writeConstitutionRuntime(repoRoot,cfg);
      upsertScripts(repoRoot);
      ensurePackagePolicy(repoRoot);
      ensureWorkflow(repoRoot);
    }else{
      sync(repoRoot,flags);
      cfg=readConfig(repoRoot);
    }
    ensureOptions(cfg);
    const result=enableDetectedAgents(repoRoot,cfg);
    writeConstitutionRuntime(repoRoot,cfg);
    writeConfig(repoRoot,cfg);
    console.log('Haiven doctor fixed repo state.');
    printDetected(result.detections);
    console.log(`Enabled adapters: ${result.enabled.length ? result.enabled.join(', ') : 'none newly enabled'}`);
    check(repoRoot);
    return;
  }
  console.log('Haiven Doctor');
  console.log('');
  console.log(`Repo: ${repoRoot}`);
  console.log(`Product: ${cfg.product}`);
  console.log(`Config: ${hasConfig?'ok':'missing'}`);
  console.log(`Docs: ${exists(path.join(repoRoot,'.haiven/HAIVEN_CONSTITUTION.md'))?'ok':'missing'}`);
  console.log('');
  printDetected(detections);
  const missing=detections.map(d=>d.agent).filter(a=>!cfg.enabledAgents?.[a]);
  if(missing.length){
    console.log('');
    console.log(`Fixable: run haiven doctor --fix to enable missing adapters: ${missing.join(', ')}`);
  }
  console.log('');
  console.log('Agents:');
  for(const a of AGENTS) console.log(`- ${a}: ${cfg.enabledAgents?.[a]?'enabled':'disabled'}`);
}

function listFleetRepos(root, recursive=false){
  const start=path.resolve(root || process.cwd());
  const repos=[];
  const ignore=new Set(['node_modules','.git','dist','.next','build','coverage']);
  if(!exists(start)) throw new Error(`Fleet root does not exist: ${start}`);
  function scan(dir, depth){
    if(exists(path.join(dir,'haiven.config.json'))){ repos.push(dir); if(!recursive) return; }
    if(!recursive && depth>=1) return;
    let entries=[];
    try{ entries=fs.readdirSync(dir,{withFileTypes:true}); }catch{return;}
    for(const entry of entries){
      if(!entry.isDirectory()) continue;
      if(ignore.has(entry.name)) continue;
      scan(path.join(dir,entry.name), depth+1);
    }
  }
  scan(start,0);
  return [...new Set(repos)].sort();
}
function fleet(repoRoot,rest,flags){
  const action=rest[0] || 'doctor';
  const root=flags.root && flags.root!==true ? flags.root : path.dirname(repoRoot);
  const recursive=!!flags.recursive;
  const repos=listFleetRepos(root, recursive);
  if(action==='list'){
    console.log(`Haiven fleet root: ${path.resolve(root)}`);
    console.log(`Repos: ${repos.length}`);
    repos.forEach(r=>console.log(`- ${r}`));
    return;
  }
  if(!['doctor','check','sync'].includes(action)) throw new Error('Use: haiven fleet <list|doctor|check|sync> [--fix] [--root <path>] [--recursive]');
  console.log(`Haiven fleet root: ${path.resolve(root)}`);
  console.log(`Repos: ${repos.length}`);
  let failures=0;
  for(const repo of repos){
    console.log(`\n=== ${action}${flags.fix?' --fix':''}: ${repo} ===`);
    const before=process.exitCode;
    process.exitCode=0;
    try{
      if(action==='doctor') doctor(repo, flags);
      else if(action==='check') check(repo);
      else if(action==='sync') sync(repo, flags);
    }catch(e){
      failures++;
      console.error(e.message || String(e));
      continue;
    }
    if(process.exitCode && process.exitCode!==0) failures++;
    process.exitCode=before;
  }
  if(failures){
    console.error(`\nFleet completed with ${failures} failing repo(s).`);
    process.exitCode=1;
  }else{
    console.log('\nFleet completed successfully.');
  }
}
function version(){
  console.log(`Haiven Compass ${COMPASS_VERSION}`);
  console.log(`Constitution version ${VERSION}`);
  console.log(`Bundled docs: ${bundledDocsRoot}`);
}

function help(){ console.log(`haiven

Commands:
  init [--product <name>] [--agent <agent>] [--agents <a,b|auto>] [--source <path>] [--hooks] [--skills] [--no-workflow]
  sync [--source <path>] [--hooks] [--skills]
  check
  doctor [--fix]
  version
  agent list
  agent enable <agent> [--hooks] [--skills] [--plugins]
  agent disable <agent>
  fleet list [--root <path>] [--recursive]
  fleet doctor [--fix] [--root <path>] [--recursive]
  fleet check [--root <path>] [--recursive]
  fleet sync [--root <path>] [--recursive]

Examples:
  haiven init
  haiven init --agent cursor
  haiven init --agents cursor,codex
  haiven init --agents auto
  haiven doctor --fix
  haiven version
  haiven fleet doctor --fix --root D:\\dev\\HaivenLabs
`); }
try{ const args=parse(process.argv.slice(2)); const repoRoot=findRepoRoot(); if(!args.cmd || args.cmd==='help' || args.cmd==='--help') help(); else if(args.cmd==='init') init(repoRoot,args.flags); else if(args.cmd==='sync') sync(repoRoot,args.flags); else if(args.cmd==='check') check(repoRoot); else if(args.cmd==='doctor') doctor(repoRoot,args.flags); else if(args.cmd==='version' || args.cmd==='--version' || args.cmd==='-v') version(); else if(args.cmd==='fleet') fleet(repoRoot,args.rest,args.flags); else if(args.cmd==='agent') agent(repoRoot,args.sub,args.rest,args.flags); else throw new Error(`Unknown command: ${args.cmd}`); }catch(e){ console.error(e.message || String(e)); process.exit(1); }
