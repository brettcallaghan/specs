---
title: "Container Blueprint: Spec CLI"
description: A fixture for the c4 node checks.
slug: container-spec-cli
type: container-blueprint
status: approved
updated_at: 2026-10-07T00:00:00Z
entity_name: Spec CLI
---

# Spec CLI

## Context

```mermaid
C4Container
    title Context
    Person(author, "Spec Author", "Writes documents")
    System_Ext(ci, "CI Pipeline", "Runs checks")
    Container(cli, "Spec CLI", "Go binary", "Indexes a corpus (and checks it), quickly")
    ContainerDb(corpus, "Spec Corpus", "Markdown files", "The documents")
    Container(twin, "Twin", "Go", "One of two containers named Twin")
    Container(ghost, "Ghost Service", "Go", "Names no element")
    Frobnicate(x, "Not a macro")
    Rel(author, cli, "Runs")
```

## Components

```mermaid
C4Component
    title Components
    Component(parser, "Parser", "Go package (planned), internal", "Reads input")
```

### Parser

```component
entity_name: Parser
slug: component-cli-parser
description: Reads input
```

## Sequence

```mermaid
C4Dynamic
    title A dynamic view
```
