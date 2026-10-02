const fs = require("node:fs");
const { execFileSync } = require("node:child_process");

const profile = process.argv[2] || "coverage.out";
const minimum = Number(process.argv[3] || 50);

if (!fs.existsSync(profile)) {
  console.error(`No existe el perfil de cobertura: ${profile}`);
  process.exit(1);
}

const output = execFileSync("go", ["tool", "cover", `-func=${profile}`], {
  encoding: "utf8",
});
const match = output.match(/total:\s+\(statements\)\s+([0-9.]+)%/);
if (!match) {
  console.error("No fue posible interpretar la cobertura total de Go.");
  process.exit(1);
}

const coverage = Number(match[1]);
console.log(`Cobertura total: ${coverage.toFixed(1)}% (mínimo ${minimum.toFixed(1)}%)`);
if (coverage < minimum) process.exit(1);
