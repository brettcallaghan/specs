@FEAT-CHK-003 @frd-checkout
Feature: Checkout

  As a shopper
  I want to pay for a basket
  So that the goods are mine

  @AC-CHK-003.1
  Scenario: A basket total is calculated silently

  @AC-CHK-003.2
  Scenario: An empty basket cannot be paid for

    The system SHALL refuse payment when the basket is empty.

    Given an empty basket
    When the shopper pays
    Then payment is refused

  Scenario Outline: Latency budgets against a corpus
    Given a corpus of documents
    When the CLI performs "<operation>"
    Then the CLI completes within "<budget>"

    @AC-CHK-003.3
    Examples: an agent asking one question
      | operation      | budget          |
      | a corpus query | 50 milliseconds |

    Examples: a write-time hook checking one file
      | operation               | budget           |
      | a single-document check | 100 milliseconds |
