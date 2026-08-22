import fs from 'node:fs';
import path from 'node:path';

export const HAIVEN_PNPM_VERSION = '11.22.0';
export const HAIVEN_NODE_VERSION = '22.23.2';

const exists = (file) => fs.existsSync(file);
const read = (file) => fs.readFileSync(file, 'utf8');

function filesUnder(root, predicate) {
  if (!exists(root)) return [];
  const output = [];
  const visit = (current) => {
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      if (['.git', '.cache', 'node_modules', '.next', 'dist', 'build', 'coverage'].includes(entry.name)) continue;
      const absolute = path.join(current, entry.name);
      if (entry.isDirectory()) visit(absolute);
      else if (entry.isFile() && predicate(absolute)) output.push(absolute);
    }
  };
  visit(root);
  return output;
}

function requireSetting(text, key, expected, errors) {
  const escaped = String(expected).replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  if (!new RegExp(`^${key}:\\s*${escaped}\\s*$`, 'm').test(text)) {
    errors.push(`pnpm-workspace.yaml must set ${key}: ${expected}.`);
  }
}

export function validateSupplyChainPolicy(repoRoot) {
  const errors = [];
  const warnings = [];
  const packagePath = path.join(repoRoot, 'package.json');
  if (!exists(packagePath)) return { errors, warnings };

  let pkg;
  try {
    pkg = JSON.parse(read(packagePath));
  } catch {
    return { errors: ['package.json is not valid JSON.'], warnings };
  }

  if (pkg.packageManager !== `pnpm@${HAIVEN_PNPM_VERSION}`) errors.push(`package.json must pin packageManager to pnpm@${HAIVEN_PNPM_VERSION}.`);
  if (pkg.engines?.node !== `>=${HAIVEN_NODE_VERSION} <23`) errors.push(`package.json must constrain Node to >=${HAIVEN_NODE_VERSION} <23.`);
  const nodeVersionPath = path.join(repoRoot, '.node-version');
  if (!exists(nodeVersionPath) || read(nodeVersionPath).trim() !== HAIVEN_NODE_VERSION) errors.push(`.node-version must pin Node ${HAIVEN_NODE_VERSION}.`);
  if (!exists(path.join(repoRoot, 'pnpm-lock.yaml'))) errors.push('Missing committed pnpm-lock.yaml.');

  const workspacePath = path.join(repoRoot, 'pnpm-workspace.yaml');
  if (!exists(workspacePath)) {
    errors.push('Missing pnpm-workspace.yaml supply-chain policy.');
  } else {
    const workspace = read(workspacePath);
    requireSetting(workspace, 'minimumReleaseAge', 1440, errors);
    requireSetting(workspace, 'minimumReleaseAgeStrict', true, errors);
    requireSetting(workspace, 'minimumReleaseAgeIgnoreMissingTime', false, errors);
    requireSetting(workspace, 'trustPolicy', 'no-downgrade', errors);
    requireSetting(workspace, 'trustLockfile', false, errors);
    requireSetting(workspace, 'blockExoticSubdeps', true, errors);
    requireSetting(workspace, 'strictDepBuilds', true, errors);
    if (!/^allowBuilds:\s*(?:\{\})?\s*$/m.test(workspace)) errors.push('pnpm-workspace.yaml must define an explicit allowBuilds policy.');
  }

  const competingLocks = filesUnder(repoRoot, (file) => ['package-lock.json', 'npm-shrinkwrap.json', 'yarn.lock'].includes(path.basename(file)));
  for (const lock of competingLocks) errors.push(`Competing package-manager lockfile is prohibited: ${path.relative(repoRoot, lock).replaceAll('\\', '/')}`);

  const dockerfiles = filesUnder(repoRoot, (file) => /^Dockerfile(?:\..+)?$/i.test(path.basename(file)));
  for (const dockerfile of dockerfiles) {
    const relative = path.relative(repoRoot, dockerfile).replaceAll('\\', '/');
    const aliases = new Set();
    for (const base of read(dockerfile).matchAll(/^FROM\s+(?:--platform=\S+\s+)?(\S+)(?:\s+AS\s+(\S+))?/gmi)) {
      if (!aliases.has(base[1]) && !base[1].includes('@sha256:')) errors.push(`${relative} uses an unpinned container base: ${base[1]}`);
      if (base[2]) aliases.add(base[2]);
    }
  }

  const workflowRoot = path.join(repoRoot, '.github', 'workflows');
  const workflows = filesUnder(workflowRoot, (file) => /\.ya?ml$/i.test(file));
  if (!workflows.length) {
    errors.push('Missing GitHub Actions workflow security enforcement.');
  } else {
    const combined = workflows.map(read).join('\n');
    if (!/actions\/setup-node@/.test(combined)) errors.push('CI must install the pinned Node runtime with actions/setup-node.');
    for (const workflow of workflows) {
      const text = read(workflow);
      const relative = path.relative(repoRoot, workflow).replaceAll('\\', '/');
      if (!/^permissions:\s*\n(?:[ \t]+.*\n)*?[ \t]+contents:\s*read\s*$/m.test(text)) errors.push(`${relative} must declare top-level contents: read permissions.`);
      for (const match of text.matchAll(/^\s*-?\s*uses:\s*([^\s#]+).*$/gm)) {
        const target = match[1];
        if (target.startsWith('./') || target.startsWith('docker://')) continue;
        if (!/^[0-9a-f]{40}$/i.test(target.split('@').at(-1))) errors.push(`${relative} uses mutable action reference: ${target}`);
      }
      const checkoutCount = (text.match(/uses:\s*actions\/checkout@/g) || []).length;
      const credentialFreeCount = (text.match(/persist-credentials:\s*false/g) || []).length;
      if (credentialFreeCount < checkoutCount) errors.push(`${relative} must disable persisted credentials for every checkout.`);
      if (/\bpull_request_target\s*:/.test(text)) errors.push(`${relative} may not use pull_request_target.`);
      if (/runs-on:\s*(?:\[[^\]]*self-hosted|self-hosted)/i.test(text)) errors.push(`${relative} may not run CI on self-hosted runners.`);
      if (/\b(?:NPM_TOKEN|NODE_AUTH_TOKEN)\b/.test(text)) errors.push(`${relative} contains a reusable npm publication token.`);
      for (const install of text.matchAll(/run:\s*pnpm\s+install([^\n]*)/g)) if (!install[1].includes('--frozen-lockfile')) errors.push(`${relative} has a non-frozen pnpm install.`);
      for (const version of text.matchAll(/node-version:\s*['"]?([^\s'"#]+)['"]?/g)) if (version[1] !== HAIVEN_NODE_VERSION) errors.push(`${relative} must pin setup-node to ${HAIVEN_NODE_VERSION}.`);
    }
    if (!/pnpm\s+(?:run\s+)?audit[^\n]*--audit-level=(?:high|critical)/.test(combined)) errors.push('CI must audit the complete pnpm dependency graph at high severity or stricter.');
    if (!/actions\/dependency-review-action@/.test(combined)) errors.push('CI must perform dependency review.');
    if (!/github\/codeql-action\/(?:init|analyze)@/.test(combined)) errors.push('CI must run CodeQL or an approved equivalent.');
    if (!/(?:anchore\/sbom-action@|cyclonedx)/i.test(combined)) errors.push('CI must generate a CycloneDX-compatible SBOM.');
  }

  const dependabotPath = path.join(repoRoot, '.github', 'dependabot.yml');
  if (!exists(dependabotPath)) {
    errors.push('Missing .github/dependabot.yml.');
  } else {
    const dependabot = read(dependabotPath);
    if (!/package-ecosystem:\s*["']?npm["']?/.test(dependabot)) errors.push('Dependabot must monitor npm dependencies.');
    if (!/package-ecosystem:\s*["']?github-actions["']?/.test(dependabot)) errors.push('Dependabot must monitor GitHub Actions.');
  }

  return { errors, warnings };
}
