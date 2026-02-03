import { chromium, devices } from '@playwright/test';
import { copyFile, mkdir, rename } from 'node:fs/promises';
import { basename, dirname, extname, resolve } from 'node:path';
import process from 'node:process';

const argv = process.argv.slice(2);

const usage = `Usage: npm run record -- --url <url> [options]

Options:
  --url <url>         Target URL to record.
  --auto              Try common local URLs if --url is not provided.
  --device <name>     Playwright device name (default: Pixel 7).
  --click <label>     Click a button by aria-label (repeatable).
  --step <action>     Custom step: click:<selector> or wait:<ms> (repeatable).
  --wait <ms>         Wait after each click (default: 400).
  --tail <ms>         Wait at the end before finishing (default: 400).
  --timeout <ms>      Action timeout (default: 10000).
  --out <path|dir>    Output file or directory (default: tmp/recordings).
  --name <file>       Output filename (default: auto timestamp).
  --headful           Run browser with a visible window.
  --help              Show this help message.

Env:
  UI_RECORD_URL, UI_RECORD_DEVICE
`;

const opts = {
  url: process.env.UI_RECORD_URL || '',
  auto: false,
  device: process.env.UI_RECORD_DEVICE || 'Pixel 7',
  wait: 400,
  tail: 400,
  timeout: 10000,
  clicks: [],
  steps: [],
  out: '',
  name: '',
  headless: true,
};

function die(message) {
  console.error(message);
  console.error(usage);
  process.exit(1);
}

function shiftValue(flag, index) {
  if (index + 1 >= argv.length) {
    die(`Missing value for ${flag}.`);
  }
  return argv[index + 1];
}

function parseIntFlag(value, flag) {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed)) {
    die(`Invalid number for ${flag}: ${value}`);
  }
  return parsed;
}

function parseStep(raw) {
  if (raw.startsWith('click:')) {
    return { type: 'click', selector: raw.slice('click:'.length) };
  }
  if (raw.startsWith('wait:')) {
    return { type: 'wait', ms: parseIntFlag(raw.slice('wait:'.length), '--step wait') };
  }
  die(`Invalid --step value: ${raw}`);
  return null;
}

for (let i = 0; i < argv.length; i += 1) {
  const arg = argv[i];
  if (arg === '--help') {
    console.log(usage);
    process.exit(0);
  }
  if (arg === '--auto') {
    opts.auto = true;
    continue;
  }
  if (arg === '--headful') {
    opts.headless = false;
    continue;
  }
  if (arg === '--url') {
    opts.url = shiftValue(arg, i);
    i += 1;
    continue;
  }
  if (arg === '--device') {
    opts.device = shiftValue(arg, i);
    i += 1;
    continue;
  }
  if (arg === '--click') {
    opts.clicks.push(shiftValue(arg, i));
    i += 1;
    continue;
  }
  if (arg === '--step') {
    opts.steps.push(parseStep(shiftValue(arg, i)));
    i += 1;
    continue;
  }
  if (arg === '--wait') {
    opts.wait = parseIntFlag(shiftValue(arg, i), arg);
    i += 1;
    continue;
  }
  if (arg === '--tail') {
    opts.tail = parseIntFlag(shiftValue(arg, i), arg);
    i += 1;
    continue;
  }
  if (arg === '--timeout') {
    opts.timeout = parseIntFlag(shiftValue(arg, i), arg);
    i += 1;
    continue;
  }
  if (arg === '--out') {
    opts.out = shiftValue(arg, i);
    i += 1;
    continue;
  }
  if (arg === '--name') {
    opts.name = shiftValue(arg, i);
    i += 1;
    continue;
  }
  if (!arg.startsWith('-') && !opts.url) {
    opts.url = arg;
    continue;
  }
  die(`Unknown argument: ${arg}`);
}

async function detectUrl() {
  const candidates = [
    'http://localhost:5173/dashboard',
    'http://localhost:8710/dashboard',
    'http://localhost:5173/',
    'http://localhost:8710/',
  ];
  for (const candidate of candidates) {
    try {
      await fetch(candidate, { redirect: 'manual' });
      return candidate;
    } catch {
      // ignore
    }
  }
  return '';
}

if (!opts.url && opts.auto) {
  opts.url = await detectUrl();
}

if (!opts.url) {
  die('Missing --url. Provide a target URL or use --auto.');
}

const availableDevices = Object.keys(devices);
const fallbackDevice =
  devices['Pixel 7'] ||
  devices['Pixel 5'] ||
  devices['iPhone 13'] ||
  devices[availableDevices[0]];
const deviceDescriptor = devices[opts.device] || fallbackDevice;
if (!devices[opts.device]) {
  const fallbackName = availableDevices.find((name) => devices[name] === deviceDescriptor);
  console.warn(`Device "${opts.device}" not found. Using "${fallbackName}".`);
}

const viewport = deviceDescriptor?.viewport || { width: 1280, height: 720 };

const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
const defaultName = `ui-recording-${timestamp}.webm`;
const outValue = opts.out || 'tmp/recordings';
const outIsFile = extname(outValue) === '.webm';
const outDir = resolve(process.cwd(), outIsFile ? dirname(outValue) : outValue);
const outName = opts.name || (outIsFile ? basename(outValue) : defaultName);
const outPath = resolve(outDir, outName);

await mkdir(outDir, { recursive: true });

const browser = await chromium.launch({ headless: opts.headless });
const context = await browser.newContext({
  ...deviceDescriptor,
  recordVideo: {
    dir: outDir,
    size: { width: viewport.width, height: viewport.height },
  },
});
const page = await context.newPage();
page.setDefaultTimeout(opts.timeout);

const response = await page.goto(opts.url, { waitUntil: 'domcontentloaded' });
if (!response) {
  console.warn('No response received from the target URL.');
}
await page.waitForLoadState('networkidle', { timeout: 5000 }).catch(() => {});
await page.waitForTimeout(200);

const steps = [];
if (opts.steps.length > 0) {
  steps.push(...opts.steps);
}
if (opts.clicks.length > 0) {
  steps.push(...opts.clicks.map((label) => ({ type: 'label', label })));
}
if (steps.length === 0) {
  steps.push(
    { type: 'label', label: 'Inbox' },
    { type: 'label', label: 'Sessions' },
    { type: 'label', label: 'Search' },
    { type: 'label', label: 'Layers' },
  );
}

const video = page.video();

for (const step of steps) {
  if (step.type === 'wait') {
    await page.waitForTimeout(step.ms);
    continue;
  }
  if (step.type === 'click') {
    await page.locator(step.selector).first().click({ timeout: opts.timeout });
    await page.waitForTimeout(opts.wait);
    continue;
  }
  if (step.type === 'label') {
    await page.getByRole('button', { name: step.label }).first().click({ timeout: opts.timeout });
    await page.waitForTimeout(opts.wait);
    continue;
  }
  die(`Unknown step type: ${step.type}`);
}

await page.waitForTimeout(opts.tail);
await context.close();
await browser.close();

const recordedPath = await video?.path();
if (!recordedPath) {
  die('Video was not recorded.');
}

if (recordedPath !== outPath) {
  try {
    await rename(recordedPath, outPath);
  } catch {
    await copyFile(recordedPath, outPath);
  }
}

console.log(`Recorded video: ${outPath}`);
