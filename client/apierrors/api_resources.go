package apierrors

type APIResource string

const (
	APIResourceApplication                            APIResource = "application"
	APIResourceUpdateDeploymentStage                  APIResource = "deployment stage"
	APIResourceApplicationCustomDomain                APIResource = "application custom domain"
	APIResourceApplicationEnvironmentVariable         APIResource = "application environment variable"
	APIResourceApplicationEnvironmentAliasVariable    APIResource = "application environment variable alias"
	APIResourceApplicationEnvironmentOverrideVariable APIResource = "application environment variable override"
	APIResourceApplicationSecret                      APIResource = "application secret"
	APIResourceApplicationSecretAlias                 APIResource = "application secret alias"
	APIResourceApplicationSecretOverride              APIResource = "application secret override"
	APIResourceApplicationStatus                      APIResource = "application status"
	APIResourceCluster                                APIResource = "cluster"
	APIResourceClusterCloudProvider                   APIResource = "cluster cloud provider"
	APIResourceClusterInstanceType                    APIResource = "cluster instance type"
	APIResourceClusterRoutingTable                    APIResource = "cluster routing table"
	APIResourceClusterStatus                          APIResource = "cluster status"
	APIResourceDatabase                               APIResource = "database"
	APIResourceDatabaseStatus                         APIResource = "database status"
	APIResourceEnvironment                            APIResource = "environment"
	APIResourceEnvironmentEnvironmentVariable         APIResource = "environment environment variable"
	APIResourceEnvironmentSecret                      APIResource = "environment secret"
	APIResourceEnvironmentStatus                      APIResource = "environment status"
	APIResourceClusterAdvancedSettings                APIResource = "cluster advanced settings"
	APIResourceProjectEnvironmentVariable             APIResource = "project environment variable"
	APIResourceServiceDeploymentRestriction           APIResource = "service deployment restriction"
)
