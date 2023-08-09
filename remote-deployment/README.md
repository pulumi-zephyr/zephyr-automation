# Standing up the Zephyr Online Store with Pulumi Remote Automation API

This Pulumi Remote Automation API program stands up all four stacks needed for a full deployment of the Zephyr online store leveraging Pulumi Deployments, a feature of Pulumi Cloud.

## Prerequisites

* The Pulumi CLI installed
* A valid set of AWS credentials configured on your system
* Go 1.19 or later installed
* NodeJS installed
* The `kubectl` CLI tool installed
* The existing Zephyr repositories (`zephyr-infra`, `zephyr-k8s`, `zephyr-data`, and `zephyr-app`) cloned onto your system
* The contents of this directory copied or cloned to your system

## Providing the Correct Configuration

TBD

## Running the Program

Once you've appropriately modified `config.yaml`, you can run the Pulumi Remote Automation API program with `go run main.go`. This will kick off a deployment on Pulumi Cloud, the progress of which you can track using the Pulumi Cloud console.

If you wish to destroy the stacks, run `go run main.go destroy`.

**NOTE:** Due to the number and types of resources involved, provisioning or destroying resources can take upwards of 15 minutes.

## Complete Example of Configuration File

TBD

## Accessing Kubernetes and the Online Store

TBD
