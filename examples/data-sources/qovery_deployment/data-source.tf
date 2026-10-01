# Reads nothing from Qovery: the data source only echoes the arguments it is given.
data "qovery_deployment" "my_deployment" {
  id = "<deployment_id>"
}
