Feature: Backend server

  In order to be prepared for remote backend calls
  As terrasec application
  I need an http server for terraform's api

  Scenario: Init a new terraform backend
    Given there is a terminal
    And there is a new terraform project
    And there is a terrasec config with "file" repository
    When I make a terrasec call with "init"
    Then terrasec should start an http server
    And the command should run properly
    And at the end the server should be stopped

  Scenario: Work through an initialized terraform project
    Given there is a terminal
    And there is a new terraform project
    And there is a terrasec config with "file" repository
    When I make a terrasec call with "init"
    Then the command should run properly
    When I make a terrasec call with "init"
    Then I should get output with pattern "Terraform has been successfully initialized!"
    Then the command should run properly
    And at the end the server should be stopped
    When I make a terrasec call with "plan"
    Then I should get output with pattern "1 to add, 0 to change, 0 to destroy."
    Then the command should run properly
    When I make a terrasec call with "apply --auto-approve"
    Then I should get output with pattern "Apply complete! Resources: 1 added, 0 changed, 0 destroyed."
    Then the command should run properly
    When I make a terrasec call with "destroy --auto-approve"
    Then I should get output with pattern "Destroy complete! Resources: 1 destroyed."
    Then the command should run properly

  Scenario: A currupted state must stop terraform run
    Given there is a terminal
    And there is an existing terrasec project
    And the saved state is broken in terms of content
    When I make a terrasec call with "plan"
    Then the command should exit with error
    Then I should get error output with pattern "Error loading state:"