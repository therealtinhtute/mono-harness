#!/usr/bin/env node
/**
 * scout-block.cjs
 * PreToolUse hook - Block heavy directories for token reduction.
 * Reads patterns from ~/.claude/.orkitignore (gitignore format).
 *
 * Exit codes:
 *   0 - Allow operation
 *   2 - Block operation
 *   Other - Error (fail-open)
 */

const { shouldBlock } = require('./lib/path-matcher.cjs');
const { log } = require('./lib/hook-logger.cjs');

const HOOK_NAME = 'scout-block';

const TOOLS_TO_CHECK = ['Read', 'Glob', 'Grep', 'Edit', 'Write'];

const BUILD_COMMANDS = [
  'npm', 'yarn', 'pnpm', 'bun',
  'go build', 'cargo build', 'mvn', 'gradle',
  'make', 'cmake', 'python setup.py', 'pip install'
];

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
    const toolInput = data.tool_input || {};

    if (!TOOLS_TO_CHECK.includes(toolName)) {
      const duration = Date.now() - startTime;
      log(HOOK_NAME, toolName, duration, 'skip', 0, '');
      process.exit(0);
      return;
    }

    let targetPath = '';
    if (toolName === 'Read' || toolName === 'Edit' || toolName === 'Write') {
      targetPath = toolInput.file_path || '';
    } else if (toolName === 'Glob') {
      targetPath = toolInput.pattern || '';
    } else if (toolName === 'Grep') {
      targetPath = toolInput.path || '';
    }

    if (toolName === 'Bash') {
      const command = toolInput.command || '';

      if (BUILD_COMMANDS.some(cmd => command.includes(cmd))) {
        const duration = Date.now() - startTime;
        log(HOOK_NAME, toolName, duration, 'allow-build', 0, '');
        process.exit(0);
        return;
      }

      const blocked = shouldBlock(command);
      if (blocked) {
        const duration = Date.now() - startTime;
        log(HOOK_NAME, toolName, duration, 'block', 2, '');
        console.error(`❌ Blocked: ${command}`);
        console.error('   Reason: Accessing heavy directory (see ~/.claude/.orkitignore)');
        process.exit(2);
        return;
      }
    }

    if (targetPath) {
      const blocked = shouldBlock(targetPath);
      if (blocked) {
        const duration = Date.now() - startTime;
        log(HOOK_NAME, toolName, duration, 'block', 2, '');
        console.error(`❌ Blocked: ${targetPath}`);
        console.error('   Reason: Heavy directory (see ~/.claude/.orkitignore)');
        process.exit(2);
        return;
      }
    }

    const duration = Date.now() - startTime;
    log(HOOK_NAME, toolName, duration, 'ok', 0, '');
    process.exit(0);
  } catch (error) {
    const duration = Date.now() - startTime;
    log(HOOK_NAME, '', duration, 'error', 0, error.message);
    process.exit(0);
  }
}

main();
