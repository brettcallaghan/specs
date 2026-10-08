@FEAT-CHK-002 @frd-checkout
Feature: Refunds

  As a shopper
  I want a placed order refunded
  So that a mistake costs me nothing

  @AC-CHK-002.1 @manual
  Scenario: A placed order is refunded
    Given a placed order
    When the shopper asks for a refund
    Then the order is refunded
