import { readFile } from "node:fs/promises";

const contractPath = new URL("../../docs/openapi.slice1.yaml", import.meta.url);
const contract = await readFile(contractPath, "utf8");

function fail(message) {
  console.error(`OpenAPI contract validation failed: ${message}`);
  process.exitCode = 1;
}

if (!/^openapi:\s*3\.1\.0\s*$/m.test(contract)) fail("expected OpenAPI 3.1.0");
if (!/^info:\s*$/m.test(contract)) fail("missing info section");
if (!/^paths:\s*$/m.test(contract)) fail("missing paths section");

const componentNames = new Set(
  [...contract.matchAll(/^\s{4}([A-Za-z][A-Za-z0-9_-]*):\s*$/gm)].map((match) => match[1]),
);
for (const match of contract.matchAll(/\$ref:\s*['"]?#\/components\/[^/]+\/([^'"\s}\]]+)/g)) {
  if (!componentNames.has(match[1])) fail(`unresolved component reference ${match[0]}`);
}

const paths = [...contract.matchAll(/^\s{2}(\/[^:\s]+):\s*$/gm)].map((match) => match[1]);
if (paths.length === 0) fail("contract has no API paths");
if (process.exitCode) process.exit(process.exitCode);
console.log(`OpenAPI contract validated (${paths.length} paths).`);
