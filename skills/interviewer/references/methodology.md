# Interview Methodology

## Framework Steps

### 1. Initialize

Read input, identify gaps, list topics to cover:

```
🎤 Interview: [Topic]
   Progress: ░░░░░░░░ 0/N topics

Topics:
  ○ Core Requirements
  ○ User Flows  
  ○ Data Model
  ○ Error Handling
  ○ Edge Cases
  ○ Security
  ○ Performance
```

### 2. Grill Loop

For each topic until confidence ≥ 80%:

**Question Depth:**
- Vague answer → Probe deeper (max 5 levels)
- Chain: "Why?" → "Why that?" → "What if not?" → "Who else affected?" → "What breaks?"

**Detect Weak Answers:**
| Signal | Response |
|--------|----------|
| Contains "probably", "maybe", "I think" | Push back: "I need specifics" |
| < 20 words | "Elaborate. What exactly..." |
| Hand-wavy | "Be specific. Give me concrete..." |

**Challenge Patterns:**
- "What if [assumption] is wrong?"
- "Why not [alternative approach]?"
- "What happens when [edge case]?"
- "Who else is affected by this?"

### 3. Track Progress

Display progress after each topic:

```
🎤 Progress: ████░░░░ 4/8 topics

Covered:
  ● Core Requirements (95%)
  ● User Flows (85%)
  ◑ Data Model (60%) ← need more
  ○ Error Handling (0%)
```

**Symbols:**
- ● Complete (≥80%)
- ◑ Partial (40-79%)
- ○ Not started (<40%)

### 4. Validate

Before generating output:

1. **Playback Summary** - repeat key points back
2. **Confirm Understanding** - "Did I get this right?"
3. **Catch Contradictions** - flag conflicts in answers
4. **Force MoSCoW Ranking**:
   - **M**ust have
   - **S**hould have
   - **C**ould have
   - **W**on't have (this time)

### 5. Exit Conditions

Stop interviewing when:
- All topics covered (confidence ≥ 80%)
- User explicitly says "enough" or "done"
- No more meaningful questions remain

## Anti-Patterns

**DON'T:**
- Ask multiple questions at once
- Accept vague answers without pushback
- Skip topics because they seem obvious
- Implement anything - only document
- Go over 15 minutes without progress check
