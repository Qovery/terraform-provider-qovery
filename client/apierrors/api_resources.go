package apierrors

type APIResource string

const (
	APIResourceAWSCredentials                 APIResource = "aws credentials"
	APIResourceScalewayCredentials            APIResource = "scaleway credentials"
	APIResourceApplication                    APIResource = "application"
	APIResourceApplicationEnvironmentVariable APIResource = "application environment variable"
	APIResourceApplicationStatus              APIResource = "application status"
	APIResourceEnvironment                    APIResource = "environment"
	APIResourceEnvironmentEnvironmentVariable APIResource = "environment environment variable"
	APIResourceOrganization                   APIResource = "organization"
	APIResourceProject                        APIResource = "project"
	APIResourceProjectEnvironmentVariable     APIResource = "project environment variable"
)
