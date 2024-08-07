# qovery_database (Data Source)

Provides a Qovery database resource. This can be used to create and manage Qovery databases.
## Example Usage
```terraform
data "qovery_database" "my_database" {
  id = "<database_id>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Id of the database.

### Optional

- `accessibility` (String) Accessibility of the database.
	- Can be: `PRIVATE`, `PUBLIC`.
	- Default: `PUBLIC`.
- `annotations_group_ids` (Set of String) List of annotations group ids
- `cpu` (Number) CPU of the database in millicores (m) [1000m = 1 CPU].
	- Must be: `>= 250`.
	- Default: `250`.
- `deployment_stage_id` (String) Id of the deployment stage.
- `icon_uri` (String) Icon URI representing the database.
- `instance_type` (String) Instance type of the database.
- `labels_group_ids` (Set of String) List of labels group ids
- `memory` (Number) RAM of the database in MB [1024MB = 1GB].
	- Must be: `>= 100`.
	- Default: `256`.
- `storage` (Number) Storage of the database in GB [1024MB = 1GB] [NOTE: can't be updated after creation].
	- Must be: `>= 10`.
	- Default: `10`.

### Read-Only

- `environment_id` (String) Id of the environment.
- `external_host` (String) The database external FQDN host [NOTE: only if your container is using a publicly accessible port].
- `internal_host` (String) The database internal host (Recommended for your application)
- `login` (String) The login to connect to your database
- `mode` (String) Mode of the database [NOTE: can't be updated after creation].
	- Can be: `CONTAINER`, `MANAGED`.
- `name` (String) Name of the database.
- `password` (String) The password to connect to your database
- `port` (Number) The port to connect to your database
- `type` (String) Type of the database [NOTE: can't be updated after creation].
	- Can be: `MONGODB`, `MYSQL`, `POSTGRESQL`, `REDIS`.
- `version` (String) Version of the database

