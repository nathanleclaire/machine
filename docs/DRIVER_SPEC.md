<!--[metadata]>
+++
draft=true
title = "Docker Machine"
description = "machine"
keywords = ["machine, orchestration, install, installation, docker, documentation"]
[menu.main]
parent="mn_install"
+++
<![end-metadata]-->

# Machine Driver Specification V1

This document details the specification that all drivers, core and plugin, are
expected to conform to.

# State of Created Machines

Machine is designed to be opinionated about the role that each of the plugin
method calls should play and the state of the rendered machine should be in,
say, a `Create` call.

### Networking

### Provisioning



### Base Operating System

In the case of local virtualization providers (e.g. VirtualBox), the default
operating system of the created instance must be
[boot2docker](https://github.com/boot2docker/boot2docker).

In the case of cloud virtualization or bare metal servers, the default operating
system must be Ubuntu Linux (currently 15.10 is the recommended version).

## API Access

We prefer accessing the provider service via HTTP APIs and strongly recommend
using those over shelling out to external executables.  For example, directly
accessing the AWS API is favored over `aws-cli` (and, indeed, this is how the
core `amazonec2` plugin is implemented.  If in doubt, contact a project
maintainer.

## SSH

The provider _must_ offer SSH access to control the instance and perform
provisioning.  This does not have to be public, but must be offered as Machine
relies on SSH for system level maintenance.

# Methods

In order to create a driver, the `Driver` interface must be fulfilled by the
plugin:

```golang
type Driver interface {
	Create() error
	DriverName() string
	GetCreateFlags() []mcnflag.Flag
	GetIP() (string, error)
	GetMachineName() string
	GetSSHHostname() (string, error)
	GetSSHKeyPath() string
	GetSSHPort() (int, error)
	GetSSHUsername() string
	GetURL() (string, error)
	GetState() (state.State, error)
	Kill() error
	PreCreateCheck() error
	Remove() error
	Restart() error
	SetConfigFromFlags(opts DriverOptions) error
	Start() error
	Stop() error
}
```

This specification defines a narrow scope for what each of these methods is
intended to accomplish.  If the driver implements additional 

## Create

`Create` should:

- 

`Create` will launch a new instance and make sure it is ready for provisioning.
This includes setting up the instance with the proper SSH keys and making sure
SSH is available including any access control (firewall).  This should return an
error on failure.

## DriverName

This method returns the name of the driver, e.g. `virtualbox`.

## Remove

`Remove` will remove the instance from the provider.  This should remove the
instance and any associated services or artifacts that were created as part
of the instance including keys and access groups.  This should return an
error on failure.

## Start

`Start` will start a stopped instance.  This should ensure the instance is
ready for operations such as SSH and Docker.  This should return an error on
failure.

## Stop

`Stop` will stop a running instance.  This should ensure the instance is
stopped and return an error on failure.

## Kill

`Kill` will forcibly stop a running instance.  This should ensure the instance
is stopped and return an error on failure.

## Restart

`Restart` will restart a running instance.  This should ensure the instance
is ready for operations such as SSH and Docker.  This should return an error
on failure.

## Status

`Status` will return the state of the instance.  This should return the
current state of the instance (running, stopped, error, etc).  This should
return an error on failure.

# Testing

Testing is strongly recommended for drivers.  Unit tests are preferred as well
as inclusion into the [integration tests](https://github.com/docker/machine#integration-tests).

# Maintaining

Driver plugin maintainers are encouraged to host their own repo and distribute
the driver plugins as executables.

# Implementation

The following describes what is needed to create a Machine Driver.  The driver
interface has methods that must be implemented for all drivers.  These include
operations such as `Create`, `Remove`, `Start`, `Stop` etc.

For details see the [Driver Interface](https://github.com/docker/machine/blob/master/drivers/drivers.go#L24).

To provide this functionality, you should embed the `drivers.BaseDriver` struct, similar to the following:

    type Driver struct {
        *drivers.BaseDriver
        DriverSpecificField string
    }

Each driver must then use an `init` func to "register" the driver:

    func init() {
        drivers.Register("drivername", &drivers.RegisteredDriver{
            New:            NewDriver,
            GetCreateFlags: GetCreateFlags,
        })
    }

## Flags

Driver flags are used for provider specific customizations.  To add flags, use
a `GetCreateFlags` func.  For example:

    func GetCreateFlags() []cli.Flag {
        return []cli.Flag{
            cli.StringFlag{
                EnvVar: "DRIVERNAME_TOKEN",
                Name:   "drivername-token",
                Usage:  "Provider access token",

            },
            cli.StringFlag{
                EnvVar: "DRIVERNAME_IMAGE",
                Name:   "drivername-image",
                Usage:  "Provider Image",
                Value:  "ubuntu-14-04-x64",
            },
        }
    }

## Examples

You can reference the existing [Drivers](https://github.com/docker/machine/tree/master/drivers)
as well.
