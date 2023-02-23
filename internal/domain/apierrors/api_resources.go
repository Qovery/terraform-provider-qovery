package apierrors

// ApiResource is an enum that contains every resource we handle using the api .
// This is used to generate a detailed error message displayed by terraform when the api return an error.
type ApiResource string

const (
	ApiResourceAWSCredentials                 ApiResource = "aws credentials"
	ApiResourceApplication                    ApiResource = "application"
	ApiResourceApplicationCustomDomain        ApiResource = "application custom domain"
	ApiResourceApplicationEnvironmentVariable ApiResource = "application environment variable"
	ApiResourceApplicationSecret              ApiResource = "application secret"
	ApiResourceApplicationStatus              ApiResource = "application status"
	ApiResourceCluster                        ApiResource = "cluster"
	ApiResourceClusterCloudProvider           ApiResource = "cluster cloud provider"
	ApiResourceClusterInstanceType            ApiResource = "cluster instance type"
	ApiResourceClusterRoutingTable            ApiResource = "cluster routing table"
	ApiResourceClusterStatus                  ApiResource = "cluster status"
	ApiResourceContainer                      ApiResource = "container"
	ApiResourceContainerEnvironmentVariable   ApiResource = "container environment variable"
	ApiResourceContainerRegistry              ApiResource = "container registry"
	ApiResourceContainerSecret                ApiResource = "container secret"
	ApiResourceContainerStatus                ApiResource = "container status"
	ApiResourceDatabase                       ApiResource = "database"
	ApiResourceDatabaseStatus                 ApiResource = "database status"
	ApiResourceEnvironment                    ApiResource = "environment"
	ApiResourceEnvironmentEnvironmentVariable ApiResource = "environment environment variable"
	ApiResourceEnvironmentSecret              ApiResource = "environment secret"
	ApiResourceEnvironmentStatus              ApiResource = "environment status"
	ApiResourceOrganization                   ApiResource = "organization"
	ApiResourceProject                        ApiResource = "project"
	ApiResourceProjectEnvironmentVariable     ApiResource = "project environment variable"
	ApiResourceProjectSecret                  ApiResource = "project secret"
	ApiResourceScalewayCredentials            ApiResource = "scaleway credentials"
	ApiResourceDeploymentStage                ApiResource = "deployment stage"
)
