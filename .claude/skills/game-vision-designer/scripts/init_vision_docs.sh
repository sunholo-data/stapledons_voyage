#!/bin/bash
# Initialize vision docs directory structure

set -e

DOCS_DIR="docs/vision"

echo "Initializing vision docs..."

mkdir -p "$DOCS_DIR"

# Create core-pillars.md if it doesn't exist
if [ ! -f "$DOCS_DIR/core-pillars.md" ]; then
    cat > "$DOCS_DIR/core-pillars.md" << 'EOF'
# Core Pillars

These are non-negotiable design constraints. Every feature must serve at least one.

## 1. [Pillar Name]

[One sentence description]

**This means:** [Concrete implications for design]

**This excludes:** [What this rules out]

---

*Pillars extracted from vision interviews. Last updated: [DATE]*
EOF
    echo "✓ Created $DOCS_DIR/core-pillars.md"
else
    echo "• $DOCS_DIR/core-pillars.md already exists"
fi

# Create design-decisions.md if it doesn't exist
if [ ! -f "$DOCS_DIR/design-decisions.md" ]; then
    cat > "$DOCS_DIR/design-decisions.md" << 'EOF'
# Design Decisions

Log of design decisions with context and rationale.

---

<!-- Template for new decisions:

## [YYYY-MM-DD] [Decision Title]

**Context:** [Why this came up]

**Decision:** [What was decided]

**Rationale:** [How this serves the pillars]

**Alternatives rejected:** [What else was considered]

**Implications:** [What this means for other features]

-->
EOF
    echo "✓ Created $DOCS_DIR/design-decisions.md"
else
    echo "• $DOCS_DIR/design-decisions.md already exists"
fi

# Create open-questions.md if it doesn't exist
if [ ! -f "$DOCS_DIR/open-questions.md" ]; then
    cat > "$DOCS_DIR/open-questions.md" << 'EOF'
# Open Questions

Unresolved design questions that need exploration.

---

<!-- Template for new questions:

## [Question]

**Why it matters:** [Impact on design]

**Current thinking:** [Where we're leaning]

**Needs:** [What would help decide]

-->
EOF
    echo "✓ Created $DOCS_DIR/open-questions.md"
else
    echo "• $DOCS_DIR/open-questions.md already exists"
fi

# Create interview-log.md if it doesn't exist
if [ ! -f "$DOCS_DIR/interview-log.md" ]; then
    cat > "$DOCS_DIR/interview-log.md" << 'EOF'
# Interview Log

Q&A session history for game vision development.

---

<!-- Template for new sessions:

## [YYYY-MM-DD] Session: [Topic]

### Q: [Question asked]

**A:** [Answer given]

### Insights

- [Key insight extracted]

### Actions

- [ ] [Follow-up action]

-->
EOF
    echo "✓ Created $DOCS_DIR/interview-log.md"
else
    echo "• $DOCS_DIR/interview-log.md already exists"
fi

echo ""
echo "Vision docs initialized at $DOCS_DIR/"
echo ""
echo "Next: Run 'interview me' to start defining core pillars"
