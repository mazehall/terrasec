repository "gopass" {
  state = "aws.private/terrasec.eks-cluster-ci"
  secret = {
    access-key = "aws.private/root-account/access-key"
    secret-key = "aws.private/root-account/secret-key"
  }
}