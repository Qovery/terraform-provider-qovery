package apierrors

// APIResource is an enum that contains every resource we handle using the api .
// This is used to generate a detailed error message displayed by terraform when the api return an error.
type APIResource string

const (
	APIResourceAWSCredentials                 APIResource = "aws credentials"
	APIResourceApplication                    APIResource = "application"
	APIResourceApplicationCustomDomain        APIResource = "application custom domain"
	APIResourceApplicationEnvironmentVariable APIResource = "application environment variable"
	APIResourceApplicationSecret              APIResource = "application secret"
	APIResourceApplicationStatus              APIResource = "application status"
	APIResourceCluster                        APIResource = "cluster"
	APIResourceClusterCloudProvider           APIResource = "cluster cloud provider"
	APIResourceClusterInstanceType            APIResource = "cluster instance type"
	APIResourceClusterRoutingTable            APIResource = "cluster routing table"
	APIResourceClusterStatus                  APIResource = "cluster status"
	APIResourceContainer                      APIResource = "container"
	APIResourceContainerCustomDomain          APIResource = "container custom domain"
	APIResourceContainerEnvironmentVariable   APIResource = "container environment variable"
	APIResourceContainerRegistry              APIResource = "container registry"
	APIResourceContainerSecret                APIResource = "container secret"
	APIResourceContainerStatus                APIResource = "container status"
	APIResourceJob                            APIResource = "job"
	APIResourceJobEnvironmentVariable         APIResource = "job environment variable"
	APIResourceJobSecret                      APIResource = "job secret"
	APIResourceJobStatus                      APIResource = "job status"
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
	APIResourceDeploymentStage                APIResource = "deployment stage"
	APIResourceDeployment                     APIResource = "deployment"
	APIGitToken                               APIResource = "git token"
)
