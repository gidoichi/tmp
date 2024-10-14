variable "GITHUB_TOKEN" {
  type = string
}

terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = ">= 6.2.3"
    }
  }
}

provider "github" {
  token = var.GITHUB_TOKEN
  owner = "gidoichi"
}

resource "github_repository" "this" {
  name                        = "secrets-store-csi-driver-provider-infisical"
  allow_merge_commit          = false
  allow_rebase_merge          = false
  delete_branch_on_merge      = true
  description                 = "Unofficial Infisical provider for the Secret Store CSI Driver."
  has_issues                  = true
  squash_merge_commit_message = "BLANK"
  squash_merge_commit_title   = "PR_TITLE"
}

resource "github_branch_protection" "default" {
  repository_id = github_repository.this.node_id
  pattern       = "main"
  required_status_checks {
    strict = true
    contexts = [
      "build-container",
      "mount / e2e-test",
      "sync / e2e-test",
      "multiple / e2e-test",
      "namespaced / e2e-test",
      "namespaced-neg / e2e-test",
      "go-test",
      "terraform-plan",
      "updated-helm-chart-index",
      "validate-deployment",
      "validate-helm-chart",
    ]
  }
}

resource "github_repository_environment" "dockerhub" {
  environment = "dockerhub"
  repository  = github_repository.this.name
}
