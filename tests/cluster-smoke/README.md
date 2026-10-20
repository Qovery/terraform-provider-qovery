# Cluster smoke tests

Scheduled integration check that creates a fresh Qovery cluster on each selected cloud provider, verifies it reaches `DEPLOYED`, asserts no Terraform drift, then destroys it.

## What runs and when

- **Workflow:** `.github/workflows/cluster-tests.yml`
- **Schedule:** AWS and GCP run Mon–Fri at 03:00 UTC; Azure runs only on Mondays at 03:00 UTC.
- **Clouds:** AWS (EKS+Karpenter), Azure (AKS), GCP (GKE Autopilot)
- **Manual trigger:** `gh workflow run cluster-tests.yml -f cloud=<cloud>` where `<cloud>` is one of `aws-karpenter`, `azure-aks`, `gcp-gke`, `all`.

## Reading the Slack message

Every weekday morning a single Slack message lands in the existing test-results channel. Azure appears only on Mondays or when selected for a manual run:

```
Scheduled Cluster Smoke Tests
:white_check_mark: aws-karpenter — success (47 min)
:x: azure-aks — failure (12 min)
:white_check_mark: gcp-gke — success (23 min)
```
