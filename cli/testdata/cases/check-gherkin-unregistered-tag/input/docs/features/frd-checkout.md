---
title: "Checkout"
description: A feature requirements document used by a snapshot case.
slug: frd-checkout
type: frd
status: approved
updated_at: 2026-08-01T00:00:00Z
---

# Checkout

The feature below is written in a fenced block, so it is owned by this
document and carries no owning-document tag.

```gherkin
@FEAT-CHK-001
Feature: Checkout

  As a shopper
  I want to pay for a basket
  So that the goods are mine

  @AC-CHK-001.1 @level-unit
  Scenario: A basket with something in it can be paid for
    Given a basket holding one item
    When the shopper pays
    Then the order is placed

  @AC-CHK-001.2 @negatve @wip
  Scenario: An empty basket cannot be paid for
    Given an empty basket
    When the shopper pays
    Then payment is refused
```
