#!/usr/bin/env node
/**
 * i18n Smoke Test — 国际化冒烟测试
 *
 * Checks:
 * 1. Translation key parity: en.json and zh.json must have identical key sets
 * 2. Hardcoded string detection: .tsx files should not contain user-facing English/Chinese literals
 * 3. Missing translation usage: all page.tsx and component .tsx files must import useTranslations or getTranslations
 * 4. Language switcher exists and is wired into the sidebar
 * 5. next-intl is installed and configured
 *
 * Usage:
 *   node scripts/i18n-smoke-test.mjs          # run all checks
 *   node scripts/i18n-smoke-test.mjs --strict  # fail on any hardcoded string (not just warnings)
 */

import {readFileSync, readdirSync, statSync, existsSync} from 'fs';
import {join, relative} from 'path';
import {fileURLToPath} from 'url';

const __dirname = fileURLToPath(new URL('.', import.meta.url));
const WEB = join(__dirname, '..');

const STRICT = process.argv.includes('--strict');
const EXIT_CODE = {pass: 0, warn: 1, fail: 2};

let warnings = 0;
let failures = 0;
let passes = 0;

// ── Helpers ────────────────────────────────────────────────────────────────

function pass(msg) {
  console.log(`  ✅ ${msg}`);
  passes++;
}

function warn(msg) {
  console.log(`  ⚠️  ${msg}`);
  warnings++;
}

function fail(msg) {
  console.log(`  ❌ ${msg}`);
  failures++;
}

function section(title) {
  console.log(`\n${'─'.repeat(60)}`);
  console.log(`  ${title}`);
  console.log('─'.repeat(60));
}

// ── Check 1: Translation key parity ────────────────────────────────────────

function checkKeyParity() {
  section('1. Translation Key Parity (en.json ↔ zh.json)');

  const enPath = join(WEB, 'messages', 'en.json');
  const zhPath = join(WEB, 'messages', 'zh.json');

  if (!existsSync(enPath) || !existsSync(zhPath)) {
    fail('Translation files missing');
    return;
  }

  const en = JSON.parse(readFileSync(enPath, 'utf-8'));
  const zh = JSON.parse(readFileSync(zhPath, 'utf-8'));

  const enKeys = new Set(flattenKeys(en));
  const zhKeys = new Set(flattenKeys(zh));

  const missingInZh = [...enKeys].filter(k => !zhKeys.has(k));
  const missingInEn = [...zhKeys].filter(k => !enKeys.has(k));

  if (missingInZh.length === 0 && missingInEn.length === 0) {
    pass(`Key parity OK — ${enKeys.size} keys match in both locales`);
  } else {
    if (missingInZh.length > 0) {
      fail(`Missing in zh.json (${missingInZh.length}): ${missingInZh.slice(0, 5).join(', ')}${missingInZh.length > 5 ? '...' : ''}`);
    }
    if (missingInEn.length > 0) {
      fail(`Missing in en.json (${missingInEn.length}): ${missingInEn.slice(0, 5).join(', ')}${missingInEn.length > 5 ? '...' : ''}`);
    }
  }
}

function flattenKeys(obj, prefix = '') {
  const keys = [];
  for (const [k, v] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${k}` : k;
    if (typeof v === 'object' && v !== null && !Array.isArray(v)) {
      keys.push(...flattenKeys(v, fullKey));
    } else {
      keys.push(fullKey);
    }
  }
  return keys;
}

// ── Check 2: Infrastructure setup ──────────────────────────────────────────

function checkInfrastructure() {
  section('2. Infrastructure Setup');

  const pkgPath = join(WEB, 'package.json');
  const pkg = JSON.parse(readFileSync(pkgPath, 'utf-8'));

  if (pkg.dependencies?.['next-intl']) {
    pass(`next-intl installed (${pkg.dependencies['next-intl']})`);
  } else {
    fail('next-intl not found in package.json dependencies');
  }

  const files = ['i18n/request.ts', 'middleware.ts', 'app/providers.tsx'];
  for (const f of files) {
    if (existsSync(join(WEB, f))) {
      pass(`${f} exists`);
    } else {
      fail(`${f} missing`);
    }
  }

  const enMsg = existsSync(join(WEB, 'messages', 'en.json'));
  const zhMsg = existsSync(join(WEB, 'messages', 'zh.json'));
  if (enMsg && zhMsg) {
    pass('Translation files (en.json, zh.json) exist');
  } else {
    fail('Translation files missing');
  }
}

// ── Check 3: Language Switcher ─────────────────────────────────────────────

function checkLanguageSwitcher() {
  section('3. Language Switcher');

  const swPath = join(WEB, 'components', 'LanguageSwitcher.tsx');
  if (existsSync(swPath)) {
    pass('LanguageSwitcher.tsx exists');
    const content = readFileSync(swPath, 'utf-8');
    if (content.includes('NEXT_LOCALE')) {
      pass('LanguageSwitcher sets NEXT_LOCALE cookie');
    } else {
      fail('LanguageSwitcher does not set NEXT_LOCALE cookie');
    }
    if (content.includes('window.location.reload') || content.includes('router.refresh')) {
      pass('LanguageSwitcher triggers page reload after locale change');
    } else {
      warn('LanguageSwitcher may not trigger page reload after locale change');
    }
  } else {
    fail('LanguageSwitcher.tsx not found');
  }

  const sidebarPath = join(WEB, 'components', 'Sidebar.tsx');
  if (existsSync(sidebarPath)) {
    const sidebar = readFileSync(sidebarPath, 'utf-8');
    if (sidebar.includes('LanguageSwitcher')) {
      pass('LanguageSwitcher is imported in Sidebar');
    } else {
      fail('LanguageSwitcher is NOT imported in Sidebar');
    }
  }
}

// ── Check 4: Hardcoded string detection ────────────────────────────────────

function checkHardcodedStrings() {
  section('4. Hardcoded String Detection');

  const dirs = [
    join(WEB, 'app'),
    join(WEB, 'components'),
  ];

  const tsxFiles = [];
  for (const dir of dirs) {
    if (existsSync(dir)) {
      tsxFiles.push(...findTsxFiles(dir));
    }
  }

  // Patterns that indicate hardcoded user-facing strings
  const patterns = [
    // JSX text content: >Some Text< or >Some Text</
    {regex: />\s*[A-Z][a-z]{2,}.*[a-z]\s*</g, desc: 'JSX text content'},
    // String literals in JSX attributes that look like labels
    {regex: /placeholder="[^"]{4,}"/g, desc: 'placeholder attributes'},
    // Button text patterns
    {regex: />(Create|Edit|Delete|Cancel|Save|Submit|Search|Filter|Loading|Approve|Reject|Run|Compare|Close|Back|Login|Logout|Register|Trigger|Build|Retry|Sign in|Sign Out)\s*[\.<]/g, desc: 'button/action text'},
  ];

  // Files to skip (no user-facing text or are test/config files)
  const skipFiles = [
    'loading.tsx', 'error.tsx', 'layout.tsx', 'providers.tsx',
    'RadarChart.tsx', 'BarChart.tsx', 'LineChart.tsx',
    'SkeletonCard.tsx', 'SkeletonPanel.tsx',
  ];

  let totalIssues = 0;
  const filesWithIssues = [];

  for (const file of tsxFiles) {
    const relPath = relative(WEB, file);
    if (skipFiles.some(skip => relPath.endsWith(skip))) continue;

    const content = readFileSync(file, 'utf-8');
    const fileIssues = [];

    for (const {regex, desc} of patterns) {
      const matches = content.match(regex);
      if (matches) {
        // Filter out strings that are already using t() calls
        const realIssues = matches.filter(m => {
          const line = getLineForMatch(content, m);
          // Skip if line already uses t() translation
          if (line.includes("t('") || line.includes('t("')) return false;
          // Skip if it's inside a translation file
          if (file.includes('messages/')) return false;
          return true;
        });
        if (realIssues.length > 0) {
          fileIssues.push(...realIssues.map(m => ({desc, match: m.trim().slice(0, 60)})));
        }
      }
    }

    if (fileIssues.length > 0) {
      filesWithIssues.push({file: relPath, issues: fileIssues});
      totalIssues += fileIssues.length;
    }
  }

  if (totalIssues === 0) {
    pass('No hardcoded user-facing strings detected');
  } else {
    const level = STRICT ? 'fail' : 'warn';
    const action = level === 'fail' ? fail : warn;
    action(`${totalIssues} potential hardcoded strings found in ${filesWithIssues.length} file(s)`);

    for (const {file, issues} of filesWithIssues.slice(0, 10)) {
      console.log(`    📄 ${file}:`);
      for (const issue of issues.slice(0, 3)) {
        console.log(`      - ${issue.desc}: "${issue.match}"`);
      }
      if (issues.length > 3) {
        console.log(`      ... and ${issues.length - 3} more`);
      }
    }

    if (filesWithIssues.length > 10) {
      console.log(`    ... and ${filesWithIssues.length - 10} more files`);
    }

    if (STRICT) {
      console.log(`\n    💡 Run without --strict to see these as warnings`);
    }
  }
}

function getLineForMatch(content, match) {
  const idx = content.indexOf(match);
  if (idx === -1) return '';
  const start = content.lastIndexOf('\n', idx - 1) + 1;
  const end = content.indexOf('\n', idx + match.length);
  return content.slice(start, end === -1 ? undefined : end);
}

function findTsxFiles(dir) {
  const files = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      files.push(...findTsxFiles(full));
    } else if (entry.endsWith('.tsx')) {
      files.push(full);
    }
  }
  return files;
}

// ── Check 5: Translation usage coverage ────────────────────────────────────

function checkTranslationUsage() {
  section('5. Translation Usage Coverage');

  const pageFiles = findTsxFiles(join(WEB, 'app')).filter(f =>
    f.endsWith('page.tsx') || f.includes('/components/')
  );

  let covered = 0;
  let notCovered = [];

  for (const file of pageFiles) {
    const relPath = relative(WEB, file);
    // Skip loading, error, layout files
    if (relPath.endsWith('loading.tsx') || relPath.endsWith('error.tsx')) continue;
    // Skip the messages directory
    if (relPath.includes('messages/')) continue;

    const content = readFileSync(file, 'utf-8');
    const hasUseTranslations = content.includes("useTranslations") || content.includes("getTranslations");
    const hasNoText = !hasUserFacingText(content);

    if (hasUseTranslations || hasNoText) {
      covered++;
    } else {
      notCovered.push(relPath);
    }
  }

  if (notCovered.length === 0) {
    pass(`All ${covered} page/component files use translations`);
  } else {
    fail(`${notCovered.length} file(s) missing translation usage:`);
    for (const f of notCovered.slice(0, 10)) {
      console.log(`    - ${f}`);
    }
    if (notCovered.length > 10) {
      console.log(`    ... and ${notCovered.length - 10} more`);
    }
  }
}

function hasUserFacingText(content) {
  // Check if file contains JSX with English text content
  const hasEnglishJSX = />\s*[A-Z][a-z]{3,}.*[a-z]\s*</.test(content);
  const hasPlaceholders = /placeholder="[A-Z][a-zA-Z\s]{4,}"/.test(content);
  const hasButtonTitles = />(Create|Edit|Delete|Cancel|Save|Submit|Search|Register|Trigger|Build)\s*[\.<]/.test(content);
  return hasEnglishJSX || hasPlaceholders || hasButtonTitles;
}

// ── Check 6: CSS font support ──────────────────────────────────────────────

function checkChineseFont() {
  section('6. Chinese Font Support');

  const cssPath = join(WEB, 'app', 'globals.css');
  if (!existsSync(cssPath)) {
    fail('globals.css not found');
    return;
  }

  const css = readFileSync(cssPath, 'utf-8');

  if (css.includes('Noto Sans SC')) {
    pass('Noto Sans SC font is included');
  } else {
    fail('Noto Sans SC font is NOT included');
  }

  if (css.includes('fonts.googleapis.com') || css.includes('@font-face')) {
    pass('Font loading mechanism found');
  } else {
    warn('No explicit font loading found (may rely on system fonts)');
  }
}

// ── Main ───────────────────────────────────────────────────────────────────

function main() {
  console.log('\n🌐 i18n Smoke Test');
  console.log('═'.repeat(60));

  checkKeyParity();
  checkInfrastructure();
  checkLanguageSwitcher();
  checkHardcodedStrings();
  checkTranslationUsage();
  checkChineseFont();

  console.log(`\n${'═'.repeat(60)}`);
  console.log(`  Results: ✅ ${passes} passed  ⚠️  ${warnings} warnings  ❌ ${failures} failed`);
  console.log('═'.repeat(60));

  if (failures > 0) {
    console.log(`\n🚨 ${failures} check(s) FAILED. Fix before deploying.`);
    process.exit(EXIT_CODE.fail);
  } else if (warnings > 0) {
    console.log(`\n⚠️  ${warnings} warning(s). Review recommended.`);
    process.exit(EXIT_CODE.warn);
  } else {
    console.log(`\n✅ All i18n checks passed!`);
    process.exit(EXIT_CODE.pass);
  }
}

main();
