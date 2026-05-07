#!/usr/bin/env node
/**
 * mandatory-instructions.cjs
 * UserPromptSubmit hook - Inject ephemeral session context on every user message.
 * Static rules live in ~/.claude/CLAUDE.md — not duplicated here.
 *
 * Exit codes:
 *   0 - Success
 *   Other - Error (fail-open, no instructions injected)
 */

const { buildInstructions } = require('./lib/instruction-builder.cjs');
const { log } = require('./lib/hook-logger.cjs');

const HOOK_NAME = 'mandatory-instructions';

function main() {
  const startTime = Date.now();

  try {
    const instructions = buildInstructions();

    const output = {
      hookSpecificOutput: {
        hookEventName: 'UserPromptSubmit',
        additionalContext: instructions
      }
    };

    console.log(JSON.stringify(output));

    const duration = Date.now() - startTime;
    log(HOOK_NAME, '', duration, 'ok', 0, '');

    process.exit(0);
  } catch (error) {
    const duration = Date.now() - startTime;
    log(HOOK_NAME, '', duration, 'error', 0, error.message);
    process.exit(0);
  }
}

main();
