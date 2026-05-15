# Handoff Context Capture Guidelines

## Step 3: Capture Context

### Technical Context
- What is being built/fixed
- Current implementation approach
- Key decisions made
- Technologies/frameworks involved

### Progress Context
- What was completed this session
- What is in progress
- What is blocked
- What is next
- Which phase / run / gate verdict the next session should resume from

### Environment Context
- Dependencies installed/updated
- Configuration changes
- Environment variables needed
- External services involved

## Step 4: Identify Blockers

### Common Blocker Types
- Missing information or requirements
- External dependencies (APIs, services)
- Technical challenges or unknowns
- Failing tests or build errors
- Merge conflicts
- Waiting for review/approval

### For Each Blocker
- Describe the issue
- State what's needed to unblock
- Suggest next steps
- Preserve blocker taxonomy when known (`BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, `BLOCKED_CONTRACT_DRIFT`)

## Step 5: Document Next Steps

### Prioritized Action Items
1. Immediate next action (what to do first)
2. Follow-up actions (what comes after)
3. Future considerations (what to think about later)

### For Each Action
- Clear, actionable description
- Context needed to execute
- Expected outcome

## Step 7: Verify Handoff Quality

### Completeness Check
- [ ] Current state clearly documented
- [ ] Progress tracked (completed/in-progress/pending)
- [ ] Blockers identified with unblock criteria
- [ ] Next steps are actionable
- [ ] Technical context sufficient for continuation
- [ ] Continuity anchors captured (phase, latest cook run, latest check verdict) when available
- [ ] No sensitive data exposed
