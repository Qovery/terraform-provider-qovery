* [Deploy an Application and Database within 3 environments](https://github.com/Qovery/terraform-examples/tree/main/examples/deploy-an-application-within-3-environments)

### Select a value from a JSON external secret

The optional `external_secrets.json_target_key` selects one JSON value from the secret identified
by `reference`. For a secret containing `{"accounts":[{"password":"example"}]}`, use
`json_target_key = "accounts.0.password"` to expose that value as `DATABASE_PASSWORD`.

Use dots between object members and numeric indexes for arrays. To select a literal dot in a
member name, escape it in the selector; HCL quoted strings also escape the backslash:
`json_target_key = "database\\.password"` selects the member `database.password`.
Other query-like characters select literal member names; this field does not evaluate arbitrary
GJSON queries. Omitting or removing `json_target_key` reads the whole secret again.

The same field is available on external secrets in containers, jobs, Helm services, Terraform
services and environments. It is not supported on `external_secret_files`.
