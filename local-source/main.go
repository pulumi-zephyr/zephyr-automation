package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"gopkg.in/yaml.v3"
)

type Project struct {
	Location string `yaml:"location"`
	Name     string `yaml:"name"`
	Nickname string `yaml:"nickname"`
}

type Environment struct {
	Region          string  `yaml:"region"`
	Organization    string  `yaml:"organization"`
	StackName       string  `yaml:"stackName"`
	BaseProject     Project `yaml:"baseProject"`
	PlatformProject Project `yaml:"platformProject"`
	AppProject      Project `yaml:"appProject"`
}

func main() {
	// Determine mode of operation; default is to refresh/update
	destroy := false
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		if argsWithoutProg[0] == "destroy" {
			destroy = true
		}
	}

	// Get configuration data
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
		os.Exit(1)
	}
	var env Environment
	err = yaml.Unmarshal(data, &env)
	if err != nil {
		fmt.Printf("Error parsing configuration information: %v\n", err)
		os.Exit(1)
	}

	// Set up context
	ctx := context.Background()

	// Set up base stack
	baseStackName := auto.FullyQualifiedStackName(env.Organization, env.BaseProject.Name, env.StackName)
	baseStack, err := auto.UpsertStackLocalSource(ctx, baseStackName, env.BaseProject.Location)
	if err != nil {
		fmt.Printf("Failed to create or select %s stack: %v\n", env.BaseProject.Nickname, err)
		os.Exit(1)
	}
	fmt.Printf("Successfully created/selected %s stack\n", env.BaseProject.Nickname)

	// Set up platform stack
	platformStackName := auto.FullyQualifiedStackName(env.Organization, env.PlatformProject.Name, env.StackName)
	platformStack, err := auto.UpsertStackLocalSource(ctx, platformStackName, env.PlatformProject.Location)
	if err != nil {
		fmt.Printf("Failed to create or select %s stack: %v\n", env.PlatformProject.Nickname, err)
		os.Exit(1)
	}
	fmt.Printf("Successfully created/selected %s stack\n", env.PlatformProject.Nickname)

	// Set up application stack
	appStackName := auto.FullyQualifiedStackName(env.Organization, env.AppProject.Name, env.StackName)
	appStack, _ := auto.UpsertStackLocalSource(ctx, appStackName, env.AppProject.Location)
	if err != nil {
		fmt.Printf("Failed to create or select %s stack: %v\n", env.AppProject.Nickname, err)
		os.Exit(1)
	}
	fmt.Printf("Successfully created/selected %s stack\n", env.AppProject.Nickname)

	// Tear down the stacks if destroy was specified
	if destroy {
		_, err := deleteStack(ctx, appStack, env.AppProject.Nickname)
		if err != nil {
			fmt.Printf("Error deleting %s stack: %v\n", env.AppProject.Nickname, err)
			os.Exit(1)
		}
		_, err = deleteStack(ctx, platformStack, env.PlatformProject.Nickname)
		if err != nil {
			fmt.Printf("Error deleting %s stack: %v\n", env.PlatformProject.Nickname, err)
			os.Exit(1)
		}
		_, err = deleteStack(ctx, baseStack, env.BaseProject.Nickname)
		if err != nil {
			fmt.Printf("Error deleting %s stack: %v\n", env.BaseProject.Nickname, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Set config values for base stack
	baseStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})

	// Run a refresh on the base stack
	_, err = refreshStack(ctx, baseStack, env.BaseProject.Nickname)
	if err != nil {
		fmt.Printf("Error encountered refreshing %s stack: %v\n", env.BaseProject.Nickname, err)
		os.Exit(1)
	}

	// Run an update of the base stack
	_, err = updateStack(ctx, baseStack, env.BaseProject.Nickname)
	if err != nil {
		fmt.Printf("Error encountered updating %s stack: %v\n", env.BaseProject.Nickname, err)
		os.Exit(1)
	}

	// Set config values for platform stack
	platformStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})
	platformStack.SetConfig(ctx, "infraOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	platformStack.SetConfig(ctx, "infraProjName", auto.ConfigValue{Value: env.BaseProject.Name, Secret: false})
	platformStack.SetConfig(ctx, "infraStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the platform stack
	_, err = refreshStack(ctx, platformStack, env.PlatformProject.Nickname)
	if err != nil {
		fmt.Printf("Error encountered refreshing %s stack: %v\n", env.PlatformProject.Nickname, err)
		os.Exit(1)
	}

	// Run an update of the platform stack
	_, err = updateStack(ctx, platformStack, env.PlatformProject.Nickname)
	if err != nil {
		fmt.Printf("Error encountered updating %s stack: %v\n", env.PlatformProject.Nickname, err)
		os.Exit(1)
	}

	// Set config values for app stack
	appStack.SetConfig(ctx, "k8sOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	appStack.SetConfig(ctx, "k8sProjName", auto.ConfigValue{Value: env.PlatformProject.Name, Secret: false})
	appStack.SetConfig(ctx, "k8sStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the app stack
	_, err = refreshStack(ctx, appStack, env.AppProject.Nickname)
	if err != nil {
		fmt.Printf("Error encountered refreshing %s stack: %v\n", env.AppProject.Nickname, err)
		os.Exit(1)
	}

	// Run an update of the app stack
	_, err = updateStack(ctx, appStack, env.AppProject.Nickname)
	if err != nil {
		fmt.Printf("Error encountered updating %s stack: %v\n", env.AppProject.Nickname, err)
		os.Exit(1)
	}
}

func refreshStack(ctx context.Context, s auto.Stack, n string) (auto.RefreshResult, error) {
	var res auto.RefreshResult
	fmt.Printf("Starting refresh of %s stack\n", n)
	tmp, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return res, err
	}
	progressStreams := []io.Writer{os.Stdout, tmp}
	res, err = s.Refresh(ctx, optrefresh.ProgressStreams(progressStreams...))
	if err != nil {
		fmt.Printf("Failed to refresh %s stack: %v\n", n, err)
		return res, err
	}
	fmt.Printf("Successfully refreshed %s stack\n", n)
	return res, err

}

func updateStack(ctx context.Context, s auto.Stack, n string) (auto.UpResult, error) {
	var res auto.UpResult
	fmt.Printf("Starting update of %s stack\n", n)
	tmp, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return res, err
	}
	progressStreams := []io.Writer{os.Stdout, tmp}
	res, err = s.Up(ctx, optup.ProgressStreams(progressStreams...))
	if err != nil {
		fmt.Printf("Failed to update %s stack: %v\n", n, err)
		return res, err
	}
	fmt.Printf("Successfully updated %s stack\n", n)
	return res, err
}

func deleteStack(ctx context.Context, s auto.Stack, n string) (auto.DestroyResult, error) {
	var res auto.DestroyResult
	fmt.Printf("Destroying %s stack\n", n)
	tmp, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return res, err
	}
	progressStreams := []io.Writer{os.Stdout, tmp}
	res, err = s.Destroy(ctx, optdestroy.ProgressStreams(progressStreams...))
	if err != nil {
		fmt.Printf("Error destroying %s stack: %v\n", n, err)
		return res, err
	}
	fmt.Printf("Successfully destroyed %s stack\n", n)
	return res, err
}
