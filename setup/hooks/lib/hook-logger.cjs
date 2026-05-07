#!/usr/bin/env node
/**
 * hook-logger.cjs
 * Zero-dependency JSONL logger with auto-rotation
 */

const fs = require('fs');
const path = require('path');

const LOG_DIR = path.join(process.env.HOME, '.claude', 'hooks', '.logs');
const LOG_FILE = path.join(LOG_DIR, 'hook-log.jsonl');
const MAX_LINES = 1000;
const TRIM_TO_LINES = 500;

function ensureLogDir() {
  try {
    if (!fs.existsSync(LOG_DIR)) {
      fs.mkdirSync(LOG_DIR, { recursive: true });
    }
  } catch (error) {
    // Fail-silent
  }
}

function rotateIfNeeded() {
  try {
    if (!fs.existsSync(LOG_FILE)) return;

    const content = fs.readFileSync(LOG_FILE, 'utf8');
    const lines = content.split('\n').filter(line => line.trim());

    if (lines.length >= MAX_LINES) {
      const trimmed = lines.slice(-TRIM_TO_LINES).join('\n') + '\n';
      fs.writeFileSync(LOG_FILE, trimmed, 'utf8');
    }
  } catch (error) {
    // Fail-silent
  }
}

function log(hook, tool, dur, status, exit, error) {
  try {
    ensureLogDir();
    rotateIfNeeded();

    const entry = { ts: new Date().toISOString(), hook, tool, dur, status, exit, error };
    fs.appendFileSync(LOG_FILE, JSON.stringify(entry) + '\n', 'utf8');
  } catch (err) {
    // Fail-silent
  }
}

function logViolation(tool, output, pattern) {
  try {
    ensureLogDir();

    const violationFile = path.join(LOG_DIR, 'violations.jsonl');
    const entry = {
      ts: new Date().toISOString(),
      tool,
      pattern,
      output: output.substring(0, 200)
    };

    fs.appendFileSync(violationFile, JSON.stringify(entry) + '\n', 'utf8');
  } catch (err) {
    // Fail-silent
  }
}

module.exports = { log, logViolation };
