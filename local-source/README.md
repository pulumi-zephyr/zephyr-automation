# Standing up the Zephyr Online Store with Pulumi Automation API

## Prerequisites

* The Pulumi CLI installed
* A valid set of AWS credentials configured on your system
* Go 1.19 or later installed
* NodeJS installed
* The `kubectl` CLI tool installed
* The existing Zephyr repositories (`zephyr-infra`, `zephyr-k8s`, and `zephyr-app`) cloned onto your system
* The contents of this directory copied or cloned to your system

## Providing the Correct Configuration

Before running the Pulumi Automation API program, you must first supply the necessary configuration. The configuration is stored in a file named `config.yaml` in the same directory as the Pulumi Automation API program itself.

You need to make the following changes to this file:

1. Set your AWS region using the `region` key.
2. Set the organization in which all your stacks should be created using the `organization` key.
3. Set the desired name of the stacks you want to create in each project using the `stackName` key.
4. Provide the correct relative paths to each of the three repositories that are used to stand up the Zephyr online store. For example:

    ```yaml
    baseProject:
      location: ../../Zephyr/zephyr-infra
      name: zephyr-infra
      nickname: base
    ```

    You'll need to do this for all three projects (using the `baseProject`, `platformProject`, and `appProject` sections).
5. Leave the `name` field unchanged for each project. (This name should correspond to the project name provided in `Pulumi.yaml` in the specified location.)
6. Feel free to modify the `nickname` field for each project; it is only used in messages being sent to the screen while the program is running.

## Running the Program

Once you've appropriately modified `config.yaml`, you can run the Pulumi Automation API program with `go run main.go`. This will refresh/update the stacks.

If you wish to destroy the stacks, run `go run main.go destroy`.

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
```
