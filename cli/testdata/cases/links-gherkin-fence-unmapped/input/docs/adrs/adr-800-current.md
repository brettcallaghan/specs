---
title: "ADR-800: Current Decision"
description: A live decision citing in Gherkin a document the map omits.
slug: adr-800-current
type: adr
status: approved
updated_at: 2026-08-01T00:00:00Z
---

# ADR-800: Current Decision

This decision refines [[adr-900-old-approach]], which has since been
replaced.

```gherkin
Feature: Refunds follow [[adr-901-new-approach]]

  Scenario: A refund is issued
    Given a paid order
    When the customer asks for a refund
    Then the refund is issued
```
