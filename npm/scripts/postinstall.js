#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const https = require("https");
const http = require("http");
const { execSync } = require("child_process");

// 支持环境变量跳过下载 (离线/CI)
if (process.env.CRABCLAW_SKILL_BINARY_PATH) {
  console.log(
    `[crabclaw-skill] Using binary from CRABCLAW_SKILL_BINARY_PATH: ${process.env.CRABCLAW_SKILL_BINARY_PATH}`
  );
  process.exit(0);
}

const pkg = require("../package.json");
const version = pkg.version;

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const platform = PLATFORM_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

if (!platform || !arch) {
  console.error(
    `[crabclaw-skill] Unsupported platform: ${process.platform}/${process.arch}`
  );
  process.exit(1);
}

const ext = platform === "windows" ? ".exe" : "";
const binaryName = `crabclaw-skill-${platform}-${arch}${ext}`;
const downloadUrl = `https://github.com/acosmi/acosmi-sdk-go/releases/download/v${version}/${binaryName}`;
const binDir = path.join(__dirname, "..", "bin");
const destPath = path.join(binDir, `crabclaw-skill${ext}`);

// 确保 bin 目录存在
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

console.log(`[crabclaw-skill] Downloading ${binaryName} from GitHub Releases...`);

function download(url, dest, redirects) {
  if (redirects === undefined) redirects = 5;
  if (redirects <= 0) {
    console.error("[crabclaw-skill] Too many redirects");
    process.exit(1);
  }

  const client = url.startsWith("https") ? https : http;
  client
    .get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        download(res.headers.location, dest, redirects - 1);
        return;
      }

      if (res.statusCode !== 200) {
        console.error(
          `[crabclaw-skill] Download failed: HTTP ${res.statusCode}`
        );
        console.error(
          `[crabclaw-skill] You can manually download from: ${downloadUrl}`
        );
        console.error(
          `[crabclaw-skill] Or set CRABCLAW_SKILL_BINARY_PATH to an existing binary`
        );
        process.exit(1);
      }

      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on("finish", () => {
        file.close();
        // 设置可执行权限 (non-Windows)
        if (process.platform !== "win32") {
          fs.chmodSync(dest, 0o755);
        }
        console.log(`[crabclaw-skill] Binary installed to ${dest}`);
      });
    })
    .on("error", (err) => {
      console.error(`[crabclaw-skill] Download error: ${err.message}`);
      console.error(
        `[crabclaw-skill] Set CRABCLAW_SKILL_BINARY_PATH to skip download`
      );
      process.exit(1);
    });
}

download(downloadUrl, destPath);
