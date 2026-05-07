#!/usr/bin/env node
/**
 * path-matcher.cjs
 * Gitignore-spec pattern matching for scout-block.
 * Reads ~/.claude/.orkitignore for block patterns.
 */

const fs = require('fs');
const path = require('path');

const ORKITIGNORE_PATH = path.join(process.env.HOME, '.claude', '.orkitignore');

let cachedPatterns = null;
let cacheTime = 0;
const CACHE_TTL = 5000;

function loadPatterns() {
  const now = Date.now();

  if (cachedPatterns && (now - cacheTime) < CACHE_TTL) {
    return cachedPatterns;
  }

  try {
    if (!fs.existsSync(ORKITIGNORE_PATH)) {
      cachedPatterns = [];
      cacheTime = now;
      return cachedPatterns;
    }

    const content = fs.readFileSync(ORKITIGNORE_PATH, 'utf8');
    const patterns = [];

    for (const line of content.split('\n')) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith('#')) continue;

      const negate = trimmed.startsWith('!');
      const pattern = negate ? trimmed.substring(1) : trimmed;
      patterns.push({ pattern, negate });
    }

    cachedPatterns = patterns;
    cacheTime = now;
    return patterns;
  } catch (error) {
    cachedPatterns = [];
    cacheTime = now;
    return cachedPatterns;
  }
}

function matchesPattern(targetPath, pattern) {
  const normalized = targetPath.replace(/\\/g, '/');

  if (normalized === pattern) return true;
  if (normalized.includes(`/${pattern}/`) || normalized.includes(`/${pattern}`)) return true;
  if (normalized.startsWith(pattern + '/')) return true;

  if (pattern.includes('*')) {
    const regex = new RegExp('^' + pattern.replace(/\*/g, '.*') + '$');
    return regex.test(normalized);
  }

  return false;
}

function shouldBlock(targetPath) {
  if (!targetPath || typeof targetPath !== 'string') return false;

  const patterns = loadPatterns();
  let blocked = false;

  for (const { pattern, negate } of patterns) {
    if (matchesPattern(targetPath, pattern)) {
      blocked = !negate;
    }
  }

  return blocked;
}

module.exports = { shouldBlock, loadPatterns };
