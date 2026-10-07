---
title: "container-alpha"
description: A fixture for the c4 command.
slug: container-alpha
type: container-blueprint
status: approved
updated_at: 2026-10-07T00:00:00Z
entity_name: Alpha
system: system-one
modules:
  - "module-gamma"
---

# Alpha

## Context

```mermaid
C4Container
    title Context — Alpha
    Container(a, "Alpha", "Go", "The first container")
    Container(b, "Beta", "Go", "The second container")
    Rel(a, b, "Calls", "HTTP")
    BiRel(a, b, "Talks with")
```

## Components

### Parser

```component
entity_name: Parser
slug: component-parser
description: Parses
```

### Writer

```component
entity_name: Writer
slug: component-writer
description: Writes
```
