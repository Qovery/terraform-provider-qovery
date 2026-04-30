# EKS Anywhere vSphere credentials using IAM access keys
resource "qovery_eks_anywhere_vsphere_credentials" "my_eks_anywhere_vsphere_creds" {
  organization_id   = qovery_organization.my_organization.id
  name              = "my-eks-anywhere-vsphere-credentials"
  vsphere_user      = var.vsphere_user
  vsphere_password  = var.vsphere_password
  access_key_id     = var.aws_access_key_id
  secret_access_key = var.aws_secret_access_key
}

# EKS Anywhere vSphere credentials using IAM role (cross-account access)
resource "qovery_eks_anywhere_vsphere_credentials" "my_eks_anywhere_vsphere_role_creds" {
  organization_id  = qovery_organization.my_organization.id
  name             = "my-eks-anywhere-vsphere-role-credentials"
  vsphere_user     = var.vsphere_user
  vsphere_password = var.vsphere_password
  role_arn         = var.aws_role_arn
}
