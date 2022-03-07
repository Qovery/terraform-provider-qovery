package apierrors

type APIResource string

const (
	APIResourceAWSCredentials                 APIResource = "aws credentials"
	APIResourceApplication                    APIResource = "application"
	APIResourceApplicationEnvironmentVariable APIResource = "application environment variable"
	APIResourceApplicationStatus              APIResource = "application status"
	APIResourceDatabase                       APIResource = "database"
	APIResourceDatabaseStatus                 APIResource = "database status"
	APIResourceEnvironment                    APIResource = "environment"
	APIResourceEnvironmentEnvironmentVariable APIResource = "environment environment variable"
	APIResourceEnvironmentStatus              APIResource = "environment status"
	APIResourceOrganization                   APIResource = "organization"
	APIResourceProject                        APIResource = "project"
	APIResourceProjectEnvironmentVariable     APIResource = "project environment variable"
	APIResourceScalewayCredentials            APIResource = "scaleway credentials"
)
