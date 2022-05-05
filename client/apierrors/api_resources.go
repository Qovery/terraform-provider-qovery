package apierrors

type APIResource string

const (
	APIResourceAWSCredentials                 APIResource = "aws credentials"
	APIResourceApplication                    APIResource = "application"
	APIResourceApplicationEnvironmentVariable APIResource = "application environment variable"
	APIResourceApplicationSecret              APIResource = "application secret"
	APIResourceApplicationStatus              APIResource = "application status"
	APIResourceCluster                        APIResource = "cluster"
	APIResourceClusterCloudProvider           APIResource = "cluster cloud provider"
	APIResourceClusterInstanceType            APIResource = "cluster instance type"
	APIResourceClusterStatus                  APIResource = "cluster status"
	APIResourceDatabase                       APIResource = "database"
	APIResourceDatabaseStatus                 APIResource = "database status"
	APIResourceEnvironment                    APIResource = "environment"
	APIResourceEnvironmentEnvironmentVariable APIResource = "environment environment variable"
	APIResourceEnvironmentSecret              APIResource = "environment secret"
	APIResourceEnvironmentStatus              APIResource = "environment status"
	APIResourceOrganization                   APIResource = "organization"
	APIResourceProject                        APIResource = "project"
	APIResourceProjectEnvironmentVariable     APIResource = "project environment variable"
	APIResourceProjectSecret                  APIResource = "project secret"
	APIResourceScalewayCredentials            APIResource = "scaleway credentials"
)
