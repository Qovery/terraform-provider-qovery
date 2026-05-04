data "qovery_eks_anywhere_vsphere_credentials" "my_eks_anywhere_vsphere_creds" {
  id              = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  organization_id = qovery_organization.my_organization.id
}
