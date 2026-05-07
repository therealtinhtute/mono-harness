#!/usr/bin/env node
/**
 * question-validator.cjs
 * PostToolUse hook - Detect prose questions and log violations
 *
 * Exit codes:
 *   0 - Always (fail-open, never blocks operations)
 */

const { detectQuestion } = require('./lib/question-detector.cjs');
const { log, logViolation } = require('./lib/hook-logger.cjs');

const HOOK_NAME = 'question-validator';

function main() {
  const startTime = Date.now();

  try {
    let input = '';
    const chunks = [];

    process.stdin.on('data', chunk => chunks.push(chunk));
    process.stdin.on('end', () => {
      input = Buffer.concat(chunks).toString('utf8');
      processInput(input, startTime);
    });
  } catch (error) {
    const duration = Date.now() - startTime;
    log(HOOK_NAME, '', duration, 'error', 0, error.message);
    process.exit(0);
  }
}

function processInput(input, startTime) {
  try {
    const data = JSON.parse(input);
    const toolName = data.tool_name || '';
    const toolOutput = data.tool_output || '';

    const result = detectQuestion(toolOutput);

    if (result.isQuestion) {
      logViolation(toolName, toolOutput, result.pattern);

      const duration = Date.now() - startTime;
      log(HOOK_NAME, toolName, duration, 'violation', 0, '');

      console.error(`⚠️  Prose question detected (pattern: ${result.pattern})`);
      console.error('   Use AskUserQuestion tool instead.');
    } else {
      const duration = Date.now() - startTime;
      log(HOOK_NAME, toolName, duration, 'ok', 0, '');
    }

    process.exit(0);
  } catch (error) {
    const duration = Date.now() - startTime;
    log(HOOK_NAME, '', duration, 'error', 0, error.message);
    process.exit(0);
  }
}

main();
