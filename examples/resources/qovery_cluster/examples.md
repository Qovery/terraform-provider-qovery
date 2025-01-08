* [Deploy an Application and Database within 3 environments](https://github.com/Qovery/terraform-examples/tree/main/examples/deploy-an-application-within-3-environments)

## Example for Karpenter requirements
You can select the `InstanceSize`, `InstanceFamily` and `Arch` to be used by Qovery node pools.
The default value generated when creating a Karpenter cluster from the console is the following:
```terraform
karpenter = {
  spot_enabled                 = false
  disk_size_in_gib             = 40
  default_service_architecture = "AMD64"
  qovery_node_pools            = {
    requirements = [
      {
        key      = "InstanceSize"
        operator = "In"
        values   = ["small", "medium", "large", "xlarge", "2xlarge", "3xlarge", "4xlarge", "6xlarge", "8xlarge", "9xlarge", "12xlarge", "16xlarge", "18xlarge", "24xlarge", "32xlarge"]
      },
      {
        key      = "InstanceFamily"
        operator = "In"
        values   = ["c5", "c5a", "c5d", "c5n", "c6g", "c6gd", "c6gn", "c6i", "c6in", "c7g", "c7i", "c7i-flex", "d2", "d3", "i3", "i3en", "i4i", "im4gn", "inf2", "is4gen", "m5", "m5a", "m5ad", "m5d", "m6g", "m6gd", "m6i", "m7g", "m7gd", "m7i", "m7i-flex", "r4", "r5", "r5a", "r5ad", "r5d", "r5dn", "r5n", "r6g", "r6gd", "r6i", "r7i", "t2", "t3", "t3a", "t4g", "x2iedn"]
      },
      {
        key      = "Arch"
        operator = "In"
        values   = ["ARM64", "AMD64"]
      }
    ]
  }
}
``