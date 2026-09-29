# IAM role that Qovery assumes (STS). For static IAM access keys, set access_key_id and secret_access_key instead.
resource "qovery_eks_anywhere_vsphere_credentials" "my_eks_anywhere_vsphere_credentials" {
  organization_id  = qovery_organization.my_organization.id
  name             = "my-eks-anywhere-vsphere-credentials"
  vsphere_user     = var.vsphere_user
  vsphere_password = var.vsphere_password
  role_arn         = var.aws_role_arn
}
