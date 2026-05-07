#!/usr/bin/env node
/**
 * question-detector.cjs
 * Detect prose questions in tool output (for question-validator hook).
 */

const QUESTION_PATTERNS = [
  { regex: /\bshould I\b/i, name: 'should-i' },
  { regex: /\bwould you like\b/i, name: 'would-you-like' },
  { regex: /\bdo you want\b/i, name: 'do-you-want' },
  { regex: /\bwhich (one|option|approach)\b/i, name: 'which-option' },
  { regex: /\bhow (should|would you like)\b/i, name: 'how-should' },
  { regex: /\bwhat (should|would you like)\b/i, name: 'what-should' },
  { regex: /\b(option|choice) [A-D]:/i, name: 'option-list' },
  { regex: /\b\d+\.\s+\w+\s+\?\s*$/m, name: 'numbered-question' },
  { regex: /\bplease (confirm|let me know|tell me)\b/i, name: 'please-confirm' },
  { regex: /\bcan you (confirm|clarify|specify)\b/i, name: 'can-you-confirm' },
  { regex: /\?\s*$/m, name: 'question-mark' }
];

function detectQuestion(text) {
  if (!text || typeof text !== 'string') return { isQuestion: false, pattern: '' };
  if (text.trim().length < 10) return { isQuestion: false, pattern: '' };

  for (const { regex, name } of QUESTION_PATTERNS) {
    if (regex.test(text)) return { isQuestion: true, pattern: name };
  }

  return { isQuestion: false, pattern: '' };
}

module.exports = { detectQuestion };
