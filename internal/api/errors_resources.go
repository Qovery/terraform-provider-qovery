package api

type apiResource string

const (
	apiResourceAWSCredentials                 apiResource = "aws credentials"
	apiResourceApplication                    apiResource = "application"
	apiResourceApplicationCustomDomain        apiResource = "application custom domain"
	apiResourceApplicationEnvironmentVariable apiResource = "application environment variable"
	apiResourceApplicationSecret              apiResource = "application secret"
	apiResourceApplicationStatus              apiResource = "application status"
	apiResourceCluster                        apiResource = "cluster"
	apiResourceClusterCloudProvider           apiResource = "cluster cloud provider"
	apiResourceClusterInstanceType            apiResource = "cluster instance type"
	apiResourceClusterRoutingTable            apiResource = "cluster routing table"
	apiResourceClusterStatus                  apiResource = "cluster status"
	apiResourceDatabase                       apiResource = "database"
	apiResourceDatabaseStatus                 apiResource = "database status"
	apiResourceEnvironment                    apiResource = "environment"
	apiResourceEnvironmentEnvironmentVariable apiResource = "environment environment variable"
	apiResourceEnvironmentSecret              apiResource = "environment secret"
	apiResourceEnvironmentStatus              apiResource = "environment status"
	apiResourceOrganization                   apiResource = "organization"
	apiResourceProject                        apiResource = "project"
	apiResourceProjectEnvironmentVariable     apiResource = "project environment variable"
	apiResourceProjectSecret                  apiResource = "project secret"
	apiResourceScalewayCredentials            apiResource = "scaleway credentials"
)
