#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

// 优先使用环境变量指定的二进制
let binaryPath = process.env.CRABCLAW_SKILL_BINARY_PATH;

if (!binaryPath) {
  const ext = process.platform === "win32" ? ".exe" : "";
  binaryPath = path.join(__dirname, `crabclaw-skill${ext}`);
}

if (!fs.existsSync(binaryPath)) {
  console.error(
    `[crabclaw-skill] Binary not found: ${binaryPath}`
  );
  console.error(
    `[crabclaw-skill] Run 'npm rebuild @acosmi/crabclaw-skill' or set CRABCLAW_SKILL_BINARY_PATH`
  );
  process.exit(1);
}

try {
  execFileSync(binaryPath, process.argv.slice(2), { stdio: "inherit" });
} catch (err) {
  if (err.status !== undefined) {
    process.exit(err.status);
  }
  throw err;
}
