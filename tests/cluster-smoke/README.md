# Cluster smoke tests

Daily integration check that creates a fresh Qovery cluster on each cloud provider, verifies it reaches `DEPLOYED`, asserts no Terraform drift, then destroys it.

## What runs and when

- **Workflow:** `.github/workflows/cluster-tests.yml`
- **Schedule:** Mon–Fri 03:00 UTC
- **Clouds:** AWS (EKS+Karpenter), Azure (AKS), GCP (GKE Autopilot)
- **Manual trigger:** `gh workflow run cluster-tests.yml -f cloud=<cloud>` where `<cloud>` is one of `aws-karpenter`, `azure-aks`, `gcp-gke`, `all`.

## Reading the Slack message

Every weekday morning a single Slack message lands in the existing test-results channel:

```
Daily Cluster Smoke Tests
:white_check_mark: aws-karpenter — success (47 min)
:x: azure-aks — failure (12 min)
:white_check_mark: gcp-gke — success (23 min)
```
