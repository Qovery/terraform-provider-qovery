# Qovery Terraform Provider 
Test

- Documentation: https://registry.terraform.io/providers/qovery/qovery/latest

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19 (to build the provider)
- [Task](https://taskfile.dev) v3 (to run Taskfile commands)
- [jq](https://stedolan.github.io/jq/download/) (to parse json from curl api calls)

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the task `build` command:

```shell
task build
```

## Developing The Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run the task `build` command:

```shell
task build
```

This will build the provider and put the provider binary in the repository `/bin` folder.

To be able to use the compiled provider binary, you will need to create a [development override](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers) for this provider in your `~/.terraformrc` pointing to the repository `/bin` folder. 
To create a development override, run the task `install-dev-override` command:

```shell
task install-dev-override
```

When you are finished using the compiled version of the provider, you can safely remove the `~/.terraformrc` file.
To remove a development override, run the task `uninstall-dev-override` command:

```shell
task uninstall-dev-override
```

## Testing The Provider

In order to run the full suite of Acceptance tests, run task `testacc` command:

```shell
task testacc
```

*Note:* Acceptance tests create real resources, and often cost money to run.

The acceptance tests require a `QOVERY_API_TOKEN` environment variable to be set. 
It corresponds to your JWT token on the Qovery's console and can be acquired using the [qovery-cli](https://github.com/Qovery/qovery-cli) with the following command (needs [jq](https://stedolan.github.io/jq/download/)): 

```shell
qovery auth ; cat ~/.qovery/context.json | jq -r .access_token | sed 's/.*Authorization: Bearer \(*\)/\1/' | tr -d '\n'
```

This JWT needs to be put inside a `.env` file at the root of the repository. 
You can use the `.env.example` file as a base for you file.  

```dotenv
QOVERY_API_TOKEN=<qovery-api-token>
```

*Note:* API tokens can be generated via the [qovery-cli](https://github.com/Qovery/qovery-cli) command `qovery token`.

In order to run the tests with extra debugging context, prefix with `TF_LOG` (see the [terraform documentation](https://www.terraform.io/docs/internals/debugging.html) for details).

```sh
TF_LOG=trace task testacc
```

To run a specific set of tests, use the `-run` flag and specify a regex pattern matching the test names.

```sh
task testacc -- -run 'TestAcc_Organization*'
```

### Using Intellij IDEA Debugger

To be able to add breakpoints in you code and use the debugger provided by Intellij IDEA, you'll need to add `--debug` in `Program arguments` field in Idea configuration.

Once you run the project in debug mod, you'll find a line in terminal looking like `TF_REATTACH_PROVIDERS='{"registry.terraform.io/Qovery/qovery":{"Protocol":"grpc","ProtocolVersion":6,"Pid":591078,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin1516819738"}}}'`

Copy and export this environment variable in the terminal you use to run Terraform Commands. Use any Terraform command in this terminal and Idea will listen to it and stop on breakpoints you set.

## Generating The Provider Documentation

The documentation is autogenerated from Description fields within the provider, and the `examples` directory.
Generating the documentation creates markdown in the `docs` folder, ready for deployment to Hashicorp.
To generate the documentation, run the task `docs` command:

*NOTE:* To generate the documentation you need to have provided the environment variable `QOVERY_API_TOKEN` as it uses the api to fetch some data present in the doc.

```sh
task docs
```

You can preview the generated documentation by copying `/docs` Markdown file content into this [preview tool](https://registry.terraform.io/tools/doc-preview).
