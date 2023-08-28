package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"gopkg.in/yaml.v3"
)

type Project struct {
	Location string `yaml:"location"`
	Name     string `yaml:"name"`
	Path     string `yaml:"path,omitempty"`
	Branch   string `yaml:"branch,omitempty"`
}

type Environment struct {
	Region          string  `yaml:"region"`
	Organization    string  `yaml:"organization"`
	StackName       string  `yaml:"stackName"`
	BaseProject     Project `yaml:"baseProject"`
	PlatformProject Project `yaml:"platformProject"`
	AppProject      Project `yaml:"appProject"`
	DataProject     Project `yaml:"dataProject"`
}

func main() {
	// Determine mode of operation; default is to refresh/update
	destroy := false
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		destroy = argsWithoutProg[0] == "destroy"
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

	// Set up each of the four stacks
	// First, define a lambda expression that calls setupStack and does error checking
	setup := func(p Project) auto.Stack {
		stack, err := setupStack(ctx, env.Organization, env.StackName, p)
		if err != nil {
			fmt.Printf("Error encountered setting up %s stack: %v\n", p.Name, err)
			os.Exit(1)
		}
		return stack
	}

	// Set up base stack
	baseStack := setup(env.BaseProject)

	// Set up platform stack
	platformStack := setup(env.PlatformProject)

	// Set up data stack
	dataStack := setup(env.DataProject)

	// Set up application stack
	appStack := setup(env.AppProject)

	// If destroy is true, destroy each of the stacks
	// The order of operations is important; app, platform, data, base
	// First, define a lambda expression to call deleteStack and handle error checking
	delete := func(s auto.Stack, n string) auto.DestroyResult {
		res, err := deleteStack(ctx, s, n)
		if err != nil {
			fmt.Printf("Error encountered deleting %s stack: %v\n", n, err)
			os.Exit(1)
		}
		return res
	}
	// Delete the stacks if destroy is true
	if destroy {
		_ = delete(appStack, env.AppProject.Name)
		_ = delete(platformStack, env.PlatformProject.Name)
		_ = delete(dataStack, env.DataProject.Name)
		_ = delete(baseStack, env.BaseProject.Name)
		return
	}

	// Destroy was not true, so set config, refresh, and then update
	// Call the refreshStack function to refresh the stacks
	// Call the updateStack function to update the stacks
	// First, define lambda expressions to call refreshStack and updateStack and handle error checking
	refresh := func(s auto.Stack, n string) auto.RefreshResult {
		res, err := refreshStack(ctx, s, n)
		if err != nil {
			fmt.Printf("Error encountered refreshing %s stack: %v\n", n, err)
			os.Exit(1)
		}
		return res
	}
	update := func(s auto.Stack, n string) auto.UpResult {
		res, err := updateStack(ctx, s, n)
		if err != nil {
			fmt.Printf("Error encountered updating %s stack: %v\n", n, err)
			os.Exit(1)
		}
		return res
	}

	// Set config values for base stack
	baseStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})

	// Run a refresh on the base stack
	_ = refresh(baseStack, env.BaseProject.Name)

	// Run an update of the base stack
	_ = update(baseStack, env.BaseProject.Name)

	// Set config values for platform stack
	platformStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})
	platformStack.SetConfig(ctx, "baseOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	platformStack.SetConfig(ctx, "baseProjName", auto.ConfigValue{Value: env.BaseProject.Name, Secret: false})
	platformStack.SetConfig(ctx, "baseStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the platform stack
	_ = refresh(platformStack, env.PlatformProject.Name)

	// Run an update of the platform stack
	_ = update(platformStack, env.PlatformProject.Name)

	// Set config values for the data stack
	dataStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})
	dataStack.SetConfig(ctx, "baseOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	dataStack.SetConfig(ctx, "baseProjName", auto.ConfigValue{Value: env.BaseProject.Name, Secret: false})
	dataStack.SetConfig(ctx, "baseStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the data stack
	_ = refresh(dataStack, env.DataProject.Name)

	// Run an update of the data stack
	_ = update(dataStack, env.DataProject.Name)

	// Set config values for app stack
	appStack.SetConfig(ctx, "platformOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	appStack.SetConfig(ctx, "dataOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	appStack.SetConfig(ctx, "platformProjName", auto.ConfigValue{Value: env.PlatformProject.Name, Secret: false})
	appStack.SetConfig(ctx, "dataProjName", auto.ConfigValue{Value: env.DataProject.Name, Secret: false})
	appStack.SetConfig(ctx, "platformStackName", auto.ConfigValue{Value: env.StackName, Secret: false})
	appStack.SetConfig(ctx, "dataStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the app stack
	_ = refresh(appStack, env.AppProject.Name)

	// Run an update of the app stack
	_ = update(appStack, env.AppProject.Name)
}

func refreshStack(ctx context.Context, s auto.Stack, n string) (auto.RefreshResult, error) {
	var res auto.RefreshResult
	fmt.Printf("Starting refresh of %s stack\n", n)
	tmp, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return res, err
	}
	progressStreams := optrefresh.ProgressStreams(os.Stdout, tmp)
	res, err = s.Refresh(ctx, progressStreams)
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
	progressStreams := optup.ProgressStreams(os.Stdout, tmp)
	res, err = s.Up(ctx, progressStreams)
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
	progressStreams := optdestroy.ProgressStreams(os.Stdout, tmp)
	res, err = s.Destroy(ctx, progressStreams)
	if err != nil {
		fmt.Printf("Error destroying %s stack: %v\n", n, err)
		return res, err
	}
	fmt.Printf("Successfully destroyed %s stack\n", n)
	return res, err
}

func setupStack(ctx context.Context, o string, n string, proj Project) (auto.Stack, error) {
	var stack auto.Stack
	stackName := auto.FullyQualifiedStackName(o, proj.Name, n)
	gitBranch := "main"
	if proj.Branch != "" {
		gitBranch = proj.Branch
	}
	repo := auto.GitRepo{
		URL:         proj.Location,
		ProjectPath: proj.Path,
		Branch:      gitBranch,
		Setup: func(ctx context.Context, w auto.Workspace) error {
			cmd := exec.Command("npm", "install")
			cmd.Dir = w.WorkDir()
			return cmd.Run()
		},
	}
	stack, err := auto.UpsertStackRemoteSource(ctx, stackName, repo)
	if err != nil {
		fmt.Printf("Failed to create or select %s stack: %v\n", proj.Name, err)
		return stack, err
	}
	fmt.Printf("Successfully created/selected %s stack\n", proj.Name)
	return stack, err
}
