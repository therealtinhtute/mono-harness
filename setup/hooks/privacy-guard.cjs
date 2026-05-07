#!/usr/bin/env node
/**
 * privacy-guard.cjs
 * PreToolUse hook - Block sensitive files with approval flow
 *
 * Exit codes:
 *   0 - Allow operation
 *   2 - Block operation (requires approval)
 *   Other - Error (fail-open)
 */

const { isSensitive, requiresApproval } = require('./lib/privacy-checker.cjs');
const { log } = require('./lib/hook-logger.cjs');

const HOOK_NAME = 'privacy-guard';

const TOOLS_TO_CHECK = ['Read', 'Edit', 'Write', 'Grep'];

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

    if (toolName === 'Bash') {
      const duration = Date.now() - startTime;
      log(HOOK_NAME, toolName, duration, 'bypass', 0, '');
      process.exit(0);
      return;
    }

    let targetPath = '';
    if (toolName === 'Read' || toolName === 'Edit' || toolName === 'Write') {
      targetPath = toolInput.file_path || '';
    } else if (toolName === 'Grep') {
      targetPath = toolInput.path || '';
    }

    if (!targetPath) {
      const duration = Date.now() - startTime;
      log(HOOK_NAME, toolName, duration, 'ok', 0, '');
      process.exit(0);
      return;
    }

    if (isSensitive(targetPath)) {
      if (requiresApproval(toolInput)) {
        const duration = Date.now() - startTime;
        log(HOOK_NAME, toolName, duration, 'block', 2, '');

        console.error('@@PRIVACY_PROMPT_START@@');
        console.error(JSON.stringify({
          file: targetPath,
          reason: 'Sensitive file detected',
          suggestion: 'Use AskUserQuestion to request approval, then retry with bash'
        }));
        console.error('@@PRIVACY_PROMPT_END@@');

        console.error(`❌ Blocked: ${targetPath}`);
        console.error('   Reason: Sensitive file (credentials, secrets, API keys)');
        console.error('   Action: Request user approval, then use bash to access');

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
