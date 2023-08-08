# Standing up the Zephyr Online Store with Pulumi Automation API

This Pulumi Automation API program stands up all four stacks needed for a full deployment of the Zephyr online store. Note that this Pulumi Automation API program assumes that the Pulumi programs for the four stacks exist on the same system where this Automation API program is being executed (this is noted in the Prerequisites below).

## Prerequisites

* The Pulumi CLI installed
* A valid set of AWS credentials configured on your system
* Go 1.19 or later installed
* NodeJS installed
* The `kubectl` CLI tool installed
* The existing Zephyr repositories (`zephyr-infra`, `zephyr-k8s`, `zephyr-data`, and `zephyr-app`) cloned onto your system
* The contents of this directory copied or cloned to your system

## Providing the Correct Configuration

Before running the Pulumi Automation API program, you must first supply the necessary configuration. The configuration is stored in a file named `config.yaml` in the same directory as the Pulumi Automation API program itself.

You need to make the following changes to this file:

1. Set your AWS region using the `region` key.
2. Set the organization in which all your stacks should be created using the `organization` key.
3. Set the desired name of the stacks you want to create in each project using the `stackName` key.
4. Provide the correct relative paths to each of the four repositories that are used to stand up the Zephyr online store. For example:

    ```yaml
    baseProject:
      location: ../../Zephyr/zephyr-infra
      name: zephyr-infra
      nickname: base
    ```

    You'll need to do this for all four projects (using the `baseProject`, `platformProject`, `dataProject`, and `appProject` sections).
5. Leave the `name` field unchanged for each project. (This name should correspond to the project name provided in `Pulumi.yaml` in the specified location.)
6. Feel free to modify the `nickname` field for each project; it is only used in messages being sent to the screen while the program is running.

## Running the Program

Once you've appropriately modified `config.yaml`, you can run the Pulumi Automation API program with `go run main.go`. This will refresh/update the stacks.

If you wish to destroy the stacks, run `go run main.go destroy`.

**NOTE:** Due to the number and types of resources involved, provisioning or destroying resources can take upwards of 15 minutes.

## Complete Example of Configuration File

```yaml
---
region: us-east-2
organization: zephyr
stackName: test
baseProject:
  location: ../../zephyr-infra
  name: zephyr-infra
  nickname: infra
platformProject:
  location: ../../zephyr-k8s
  name: zephyr-k8s
  nickname: k8s
appProject:
  location: ../../zephyr-app/infra
  name: zephyr-app
  nickname: app
dataProject:
  location: ../../zephyr-data
  name: zephyr-data
  nickname: data
```

## Accessing Kubernetes and the Online Store

Once the Automation API program has finished running (note: this can take upwards of 15 minutes), you can use the `pulumi` CLI to retrieve the Kubeconfig output from the `zephyr-k8s` stack (using `pulumi stack output`). Follow these steps:

1. Switch into the directory where the `zephyr-k8s` Pulumi project is found.
2. Run `pulumi -s <stack-name> stack output kubeconfig > eks-kubeconfig`, where `stack-name` corresponds to the stack name supplied in `config.yaml` using the `stackName` key.

    This will create a file named `eks-kubeconfig` that you can use with `kubectl` to access the Kubernetes cluster.
3. Use `kubectl` to retrieve the DNS name of the AWS Elastic Load Balancer (ELB) provisioned for the online store:

        KUBECONFIG=eks-kubeconfig kubectl -n ui get svc ui-lb

    Use the DNS name displayed in the `EXTERNAL-IP` column in your browser to access the application. Be sure to specify `http://` in order to connect; there is no SSL/TLS support currently.
