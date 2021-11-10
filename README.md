# Terraform Provider Qovery

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-qovery
```

## Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Setup terraform for dev purposes by adding the following content to the `~/.terraformrc` file:

```
provider_installation {

  dev_overrides {
      "qovery.com/api/qovery" = "/XYZ/go/bin"
  }

  direct {}
  
}
```

where `/XYZ/go/bin` is a path to your `$GOPATH` bin folder

Run the following command to initialize the workspace and apply the sample configuration (adjust values like API token in the `main.tf` file before):

```shell
$ terraform apply
```
