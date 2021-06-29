Feature: Terrasec cli
  In order to use gopass as vault for state and secrets
  As a terraform user
  I need a comfortable tool to get that together

  Scenario: Command line interface prints version
    Given there is a terminal
    When I make a terrasec call with "--version"
    Then I should get output with pattern "terrasec.*v\d+\.\d+\.\d+"
  
  Scenario: Required terrasec config check
    Given there is a terminal
     And there is a new terraform project
    When I make a terrasec call with "init"
    Then I should get output with pattern "failed to read configuration"