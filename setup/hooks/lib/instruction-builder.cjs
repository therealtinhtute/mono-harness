#!/usr/bin/env node
/**
 * instruction-builder.cjs
 * Inject ephemeral session context: today's date and .kit/ path conventions.
 * Static rules live in ~/.claude/CLAUDE.md and rules/ — not injected here.
 */

const fs = require('fs');

/**
 * @returns {string} Ephemeral context string (~30 tokens)
 */
function buildInstructions() {
  const date = new Date().toISOString().split('T')[0];
  const hasTasks = fs.existsSync('.kit/tasks/active.md') || fs.existsSync('tasks/todo.md');
  let msg = `Date: ${date}. Output → \`.kit/\` (plans: \`.kit/plans/{date}-{slug}/\`, reports: \`.kit/reports/{skill}/\`).`;
  if (hasTasks) msg += ` Active tasks available.`;
  return msg;
}

module.exports = { buildInstructions };
