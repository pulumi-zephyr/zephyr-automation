package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremotedestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremoterefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremoteup"
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

	// Establish credentials
	creds := map[string]auto.EnvVarValue{
		"AWS_REGION":            {Value: env.Region},
		"AWS_ACCESS_KEY_ID":     {Value: os.Getenv("AWS_ACCESS_KEY_ID"), Secret: true},
		"AWS_SECRET_ACCESS_KEY": {Value: os.Getenv("AWS_SECRET_ACCESS_KEY"), Secret: true},
		"AWS_SESSION_TOKEN":     {Value: os.Getenv("AWS_SESSION_TOKEN"), Secret: true},
	}

	// Call the setupStack function to set up each of the four stacks
	// Set up base stack
	baseStack, err := setupStack(ctx, env.Organization, env.StackName, env.BaseProject, creds)
	if err != nil {
		fmt.Printf("Error encountered setting up %s stack: %v\n", env.BaseProject.Name, err)
		os.Exit(1)
	}

	// Set up platform stack
	platformStack, err := setupStack(ctx, env.Organization, env.StackName, env.PlatformProject, creds)
	if err != nil {
		fmt.Printf("Error encountered setting up %s stack: %v\n", env.PlatformProject.Name, err)
		os.Exit(1)
	}

	// Set up data stack
	dataStack, err := setupStack(ctx, env.Organization, env.StackName, env.DataProject, creds)
	if err != nil {
		fmt.Printf("Error encountered setting up %s stack: %v\n", env.DataProject.Name, err)
		os.Exit(1)
	}

	// Set up application stack
	appStack, err := setupStack(ctx, env.Organization, env.StackName, env.AppProject, creds)
	if err != nil {
		fmt.Printf("Error encountered setting up %s stack: %v\n", env.AppProject.Name, err)
		os.Exit(1)
	}

	// If destroy is true, call deleteStack function to destroy stacks
	// The order of operations is important; app, platform, data, base
	if destroy {
		_, err := deleteStack(ctx, appStack, env.AppProject.Name)
		if err != nil {
			fmt.Printf("Error deleting %s stack: %v\n", env.AppProject.Name, err)
			os.Exit(1)
		}
		_, err = deleteStack(ctx, platformStack, env.PlatformProject.Name)
		if err != nil {
			fmt.Printf("Error deleting %s stack: %v\n", env.PlatformProject.Name, err)
			os.Exit(1)
		}
		_, err = deleteStack(ctx, dataStack, env.DataProject.Name)
		if err != nil {
			fmt.Printf("Error deleting %s stack: %v\n", env.DataProject.Name, err)
			os.Exit(1)
		}
		_, err = deleteStack(ctx, baseStack, env.BaseProject.Name)
		if err != nil {
			fmt.Printf("Error deleting %s stack: %v\n", env.BaseProject.Name, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Destroy was not true, so set config, refresh, and then update
	// Call the refreshStack function to refresh the stacks
	// Call the updateStack function to update the stacks
	// Set config values for base stack
	baseStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})

	// Run a refresh on the base stack
	_, err = refreshStack(ctx, baseStack, env.BaseProject.Name)
	if err != nil {
		fmt.Printf("Error encountered refreshing %s stack: %v\n", env.BaseProject.Name, err)
		os.Exit(1)
	}

	// Run an update of the base stack
	_, err = updateStack(ctx, baseStack, env.BaseProject.Name)
	if err != nil {
		fmt.Printf("Error encountered updating %s stack: %v\n", env.BaseProject.Name, err)
		os.Exit(1)
	}

	// Set config values for platform stack
	platformStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})
	platformStack.SetConfig(ctx, "baseOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	platformStack.SetConfig(ctx, "baseProjName", auto.ConfigValue{Value: env.BaseProject.Name, Secret: false})
	platformStack.SetConfig(ctx, "baseStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the platform stack
	_, err = refreshStack(ctx, platformStack, env.PlatformProject.Name)
	if err != nil {
		fmt.Printf("Error encountered refreshing %s stack: %v\n", env.PlatformProject.Name, err)
		os.Exit(1)
	}

	// Run an update of the platform stack
	_, err = updateStack(ctx, platformStack, env.PlatformProject.Name)
	if err != nil {
		fmt.Printf("Error encountered updating %s stack: %v\n", env.PlatformProject.Name, err)
		os.Exit(1)
	}

	// Set config values for the data stack
	dataStack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: env.Region, Secret: false})
	dataStack.SetConfig(ctx, "baseOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	dataStack.SetConfig(ctx, "baseProjName", auto.ConfigValue{Value: env.BaseProject.Name, Secret: false})
	dataStack.SetConfig(ctx, "baseStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the data stack
	_, err = refreshStack(ctx, dataStack, env.DataProject.Name)
	if err != nil {
		fmt.Printf("Error encountered refreshing %s stack: %v\n", env.DataProject.Name, err)
		os.Exit(1)
	}

	// Run an update of the data stack
	_, err = updateStack(ctx, dataStack, env.DataProject.Name)
	if err != nil {
		fmt.Printf("Error encountered updating %s stack: %v\n", env.DataProject.Name, err)
		os.Exit(1)
	}

	// Set config values for app stack
	appStack.SetConfig(ctx, "platformOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	appStack.SetConfig(ctx, "dataOrgName", auto.ConfigValue{Value: env.Organization, Secret: false})
	appStack.SetConfig(ctx, "platformProjName", auto.ConfigValue{Value: env.PlatformProject.Name, Secret: false})
	appStack.SetConfig(ctx, "dataProjName", auto.ConfigValue{Value: env.DataProject.Name, Secret: false})
	appStack.SetConfig(ctx, "platformStackName", auto.ConfigValue{Value: env.StackName, Secret: false})
	appStack.SetConfig(ctx, "dataStackName", auto.ConfigValue{Value: env.StackName, Secret: false})

	// Run a refresh on the app stack
	_, err = refreshStack(ctx, appStack, env.AppProject.Name)
	if err != nil {
		fmt.Printf("Error encountered refreshing %s stack: %v\n", env.AppProject.Name, err)
		os.Exit(1)
	}

	// Run an update of the app stack
	_, err = updateStack(ctx, appStack, env.AppProject.Name)
	if err != nil {
		fmt.Printf("Error encountered updating %s stack: %v\n", env.AppProject.Name, err)
		os.Exit(1)
	}
}

func refreshStack(ctx context.Context, s auto.RemoteStack, n string) (auto.RefreshResult, error) {
	var res auto.RefreshResult
	fmt.Printf("Starting refresh of %s stack\n", n)
	// tmp, err := os.CreateTemp(os.TempDir(), "")
	// if err != nil {
	// 	fmt.Printf("Error creating temporary file: %v\n", err)
	// 	return res, err
	// }
	// progressStreams := []io.Writer{os.Stdout, tmp}
	stdoutStreamer := optremoterefresh.ProgressStreams(os.Stdout)
	res, err := s.Refresh(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to refresh %s stack: %v\n", n, err)
		return res, err
	}
	fmt.Printf("Successfully refreshed %s stack\n", n)
	return res, err

}

func updateStack(ctx context.Context, s auto.RemoteStack, n string) (auto.UpResult, error) {
	var res auto.UpResult
	fmt.Printf("Starting update of %s stack\n", n)
	// tmp, err := os.CreateTemp(os.TempDir(), "")
	// if err != nil {
	// 	fmt.Printf("Error creating temporary file: %v\n", err)
	// 	return res, err
	// }
	// progressStreams := []io.Writer{os.Stdout, tmp}
	stdoutStreamer := optremoteup.ProgressStreams(os.Stdout)
	res, err := s.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to update %s stack: %v\n", n, err)
		return res, err
	}
	fmt.Printf("Successfully updated %s stack\n", n)
	return res, err
}

func deleteStack(ctx context.Context, s auto.RemoteStack, n string) (auto.DestroyResult, error) {
	var res auto.DestroyResult
	fmt.Printf("Destroying %s stack\n", n)
	// tmp, err := os.CreateTemp(os.TempDir(), "")
	// if err != nil {
	// 	fmt.Printf("Error creating temporary file: %v\n", err)
	// 	return res, err
	// }
	// progressStreams := []io.Writer{os.Stdout, tmp}
	stdoutStreamer := optremotedestroy.ProgressStreams(os.Stdout)
	res, err := s.Destroy(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Error destroying %s stack: %v\n", n, err)
		return res, err
	}
	fmt.Printf("Successfully destroyed %s stack\n", n)
	return res, err
}

func setupStack(ctx context.Context, o string, n string, proj Project, c map[string]auto.EnvVarValue) (auto.RemoteStack, error) {
	var stack auto.RemoteStack
	stackName := auto.FullyQualifiedStackName(o, proj.Name, n)
	gitBranch := "main"
	if proj.Branch != "" {
		gitBranch = proj.Branch
	}
	repo := auto.GitRepo{
		URL:         proj.Location,
		ProjectPath: proj.Path,
		Branch:      gitBranch,
		// Setup: func(ctx context.Context, w auto.Workspace) error {
		// 	cmd := exec.Command("npm", "install")
		// 	cmd.Dir = w.WorkDir()
		// 	return cmd.Run()
		// },
	}
	stack, err := auto.UpsertRemoteStackGitSource(ctx, stackName, repo, auto.RemoteEnvVars(c))
	if err != nil {
		fmt.Printf("Failed to create or select %s stack: %v\n", proj.Name, err)
		return stack, err
	}
	fmt.Printf("Successfully created/selected %s stack\n", proj.Name)
	return stack, err
}
