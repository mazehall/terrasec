# Terrasec

Secure your secrets and the [Terraform by HashiCorp](https://www.terraform.io/) state file
through encryption with your own keys and even within teams.

## Requirements

- [_Terraform_](https://learn.hashicorp.com/tutorials/terraform/install-cli?in=terraform/aws-get-started)
- [_Gopass_](https://github.com/gopasspw/gopass/blob/master/docs/setup.md)

Terrasec is only interesting for users who understand gopass and the underlying GnuPG concepts.

You don't get security ready-made. You have to get used to it and keep learning and updating. It's worth it for that.

## Installation

Pre-compiled packages will be offered in the future if required.

To compile the Terrasec binary from source, clone the Terrasec repository.

```bash
git clone https://github.com/mazehall/terrasec.git
```

Navigate to the new directory.

```bash
cd terrasec
```

Then, compile the binary. This command will compile the binary and usally store it in $HOME/go/bin/terrasec.

```bash
go install
```

Finally, make sure that the terrasec binary is available on your PATH. This process will differ depending on your operating system.

## Usage

Since terrasec only brings a few unique commands of its own, each terraform call can be replaced by terrasec. It is mainly a wrapper that connects to gopass under the hood.

### Terraform project

You have to configure an empty http backend in your terraform project first.

```hcl
terraform {
  backend "http" {}
}
```

### Gopass mapping

One information is required: the state-file location.

Create a configuration file `terrasec.hcl` in the same directory where your terraform project starts.

```hcl
repository "gopass" {
  state = "aws.private/terrasec.eks-cluster-ci"
}
```

Then leave the further initialization to terrasec.

```bash
terrasec init
```

#### Planned feature

Optionally, you could also define secrets in the file. These are then passed on to Terraform.

```hcl
repository "gopass" {
  state = "aws.private/terrasec.eks-cluster-ci"
  secret = {
    access-key: "aws.private/root-account/access-key"
    secret-key: "aws.private/root-account/secret-key"
  }
}
```

#### Example gopass structure

```txt
gopass
├── apps/
│   └── ...
|
├── aws.private (/<path-to>/gopass/stores/aws.private)
│   ├── root-account/
│   │   ├── access-key
│   │   └── secret-key
│   └── terrasec.eks-cluster-ci
│ 
├── team.corp.com (/<path-to>/gopass/stores/team.corp.com)
│   ├── terrasec.project-big-thing/
│   │   ├── state
│   │   └── secrets/
│   │       ├── ...

```

## Development

### BDD tests

The godog Cucumber interpreter and test runner is integrated in the `go test` configuration.
