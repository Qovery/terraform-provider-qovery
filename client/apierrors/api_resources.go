package apierrors

type APIResource string

const (
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
	APIResourceDatabase                       APIResource = "database"
	APIResourceDatabaseStatus                 APIResource = "database status"
	APIResourceEnvironment                    APIResource = "environment"
	APIResourceEnvironmentEnvironmentVariable APIResource = "environment environment variable"
	APIResourceEnvironmentSecret              APIResource = "environment secret"
	APIResourceEnvironmentStatus              APIResource = "environment status"
)
