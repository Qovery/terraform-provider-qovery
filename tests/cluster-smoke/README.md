# Cluster smoke tests

Daily integration check that creates a fresh Qovery cluster on each cloud provider, verifies it reaches `DEPLOYED`, asserts no Terraform drift, then destroys it.

## What runs and when

- **Workflow:** `.github/workflows/cluster-tests.yml`
- **Schedule:** Mon–Fri 03:00 UTC (enabled in a later commit; until then, `workflow_dispatch` only)
- **Clouds:** AWS (EKS+Karpenter), Azure (AKS), GCP (GKE Autopilot), Scaleway (Kapsule)
- **Manual trigger:** `gh workflow run cluster-tests.yml -f cloud=<cloud>` where `<cloud>` is one of `aws-karpenter`, `azure-aks`, `gcp-gke`, `scaleway-kapsule`, `all`.

## Reading the Slack message

Every weekday morning a single Slack message lands in the existing test-results channel:

```
Daily Cluster Smoke Tests
:white_check_mark: aws-karpenter — success (47 min)
:x: azure-aks — failure (12 min)
:white_check_mark: gcp-gke — success (23 min)
:white_check_mark: scaleway-kapsule — success (18 min)
```

Click "View Run" to see logs. Per-cloud duration is the heartbeat for speed regressions — eyeball it over a week.

## Orphan cluster cleanup

There is no automated reaper. If a workflow run is hard-cancelled mid-apply, a real cluster may be left running on the cloud provider.

**Owner:** TODO (fill in before merge)

**Weekly check:**

1. Open the Qovery console for the test organization.
2. List clusters with name prefix `tf-smoke-`.
3. For any cluster older than 24h: delete via console or Qovery API.

## Failure triage

- **Apply timeout (>120 min):** likely Qovery infrastructure issue — check cluster status in Qovery console; the cluster may eventually finish. Re-run `terraform destroy` from a fresh local checkout if needed.
- **`plan -detailed-exitcode` returns 2 (drift):** record which fields changed. Stable user-visible field drift → file a bug. Computed/volatile field drift (timestamps, IDs) → provider needs to mark them appropriately.
- **Destroy retried twice and failed:** manually delete via Qovery console. Investigate destroy error logs for a follow-up ticket.

## Costs (rough)

Per scheduled run: ~$1–3 in cloud compute across all four clouds (cluster control plane minutes + minimal nodes). Mon–Fri only = ~$25–65/month.
