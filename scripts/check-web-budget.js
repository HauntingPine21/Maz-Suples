const fs = require("node:fs");
const path = require("node:path");
const zlib = require("node:zlib");

const root = path.resolve("web", "js");
const files = fs.readdirSync(root)
  .filter((name) => name.endsWith(".js"))
  .sort();
const source = Buffer.concat(files.map((name) => fs.readFileSync(path.join(root, name))));
const compressed = zlib.gzipSync(source, { level: 9 });
const limit = 170 * 1024;

console.log(`JavaScript inicial comprimido: ${compressed.length} bytes; presupuesto: ${limit} bytes`);
if (compressed.length > limit) process.exit(1);
