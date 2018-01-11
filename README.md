# tuxedo

Config file configuration for Jenkins instances.

### Summary

Tuxedo is an [HCL](https://github.com/hashicorp/hcl) based configuration
language used to configure Jenkins instances.


### Build and Installation

Tuxedo is built using [Go](https://golang.org/), and requires [dep](https://github.com/golang/dep)
for installing dependencies.

With these installed, building the `tuxedo` binary is simply a matter of:

```
$ dep ensure
$ go build
```

### Usage

Tuxedo runs in `--dry-run` mode, which does not actually apply changes
but just prints out the changes it will make when it is run normally.

```
$ ./tuxedo --dry-run
```

Running `tuxedo` without ``-dry-run` actually applies changes to the
Jenkins server you are pointed to.

```
$ ./tuxedo
```

### Example configuration

Tuxedo requires there be a `jenkins.tux` file in the directory that you
are running the `tuxedo` binary in.

All configurations require an `ssh_settings` block.

```terraform
// jenkins.tux

ssh_settings {
  host_ip = "35.193.201.174"
  ssh_username = "sidshanker"
  path_to_key = "/Users/fin/.ssh/id_rsa"
}

security {
  disable_signup = true
  disable_remember_me = true
}

general {
  jenkins_dir = "/var/lib/jenkins"
  num_executors = 2
  workspace_dir = "${JENKINS_HOME}/workspace/${ITEM_FULLNAME}"
}
```

### Status

At the moment, Tuxedo is very much a work in progress. It currently only supports
making changes to basic security settings.

Future changes will include an easier framework for supporting new Jenkins features,
as well as support for configuring Jenkins plugins and pipelines.

### Inspiration

This is based on Hashicorp's [HCL](https://github.com/hashicorp/hcl), and is very much inspired by [terraform](https://terraform.io).
