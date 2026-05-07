#!/usr/bin/env node
/**
 * privacy-checker.cjs
 * Sensitive file detection patterns for privacy-guard hook.
 */

const SENSITIVE_PATTERNS = [
  /\.env$/i, /\.env\./i, /\.env-.*/i,
  /credentials/i, /secrets/i, /\.secret/i,
  /api[-_]?key/i, /apikey/i,
  /token/i, /\.token/i,
  /id_rsa/i, /id_dsa/i, /id_ecdsa/i, /id_ed25519/i,
  /\.pem$/i, /\.key$/i,
  /\.aws\/credentials/i, /\.gcp\/credentials/i, /gcloud\/credentials/i,
  /\.pgpass/i, /\.my\.cnf/i,
  /\.crt$/i, /\.cer$/i, /\.p12$/i, /\.pfx$/i
];

function isSensitive(filePath) {
  if (!filePath || typeof filePath !== 'string') return false;

  const normalized = filePath.replace(/\\/g, '/');

  for (const pattern of SENSITIVE_PATTERNS) {
    if (pattern.test(normalized)) return true;
  }

  return false;
}

function requiresApproval(toolInput) {
  // Always require approval — bash bypass is the intended approval flow
  return true;
}

module.exports = { isSensitive, requiresApproval };
