package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/panosg/gbd/pkg/gbd"
)

var stack *gbd.Stack

var version = "0.0.1"

func main() {

	var config string
	var contextDir string
	var dumpConfig bool

	var dryRun = &cobra.Command{
		Use:   "dry-run {context path} {config file (*.yaml)}",
		Short: "Test a configuration file for deployment",
		Run:   dryRun,
	}

	var watchConfig = &cobra.Command{
		Use:   "watch {context path} {config file (*.yaml)}",
		Short: "Deploy a predefined stack from a config file & watch for changes",
		Run:   watchConfig,
	}

	dryRun.Flags().StringVarP(&contextDir, "context", "c", "", "context path")
	dryRun.Flags().StringVarP(&config, "config", "f", "", "config file (*.yaml) from context path")

	watchConfig.Flags().StringVarP(&contextDir, "context", "c", "", "context path")
	watchConfig.Flags().StringVarP(&config, "config", "f", "", "config file (*.yaml) from context path")
	watchConfig.Flags().BoolVarP(&dumpConfig, "dump", "d", false, "dump config file to context path")

	var rootCmd = &cobra.Command{Use: "gbd", Version: version}
	rootCmd.AddCommand(dryRun)
	rootCmd.AddCommand(watchConfig)

	log.Printf("GBD - GoBrewDock %s\n", version)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func dryRun(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	contextDir, _ := cmd.Flags().GetString("context")
	config, _ := cmd.Flags().GetString("config")

	path := strings.Join([]string{contextDir, config}, "/")
	stack = buildStack(ctx, path, false)

	cancel()

	<-ctx.Done()
	log.Println("Shutting down...")
	if err := stack.Teardown(context.Background()); err != nil {
		log.Println(err)
		os.Exit(1)
	}

}

func watchConfig(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	contextDir, _ := cmd.Flags().GetString("context")
	config, _ := cmd.Flags().GetString("config")
	dump, _ := cmd.Flags().GetBool("dump")

	path := filepath.Join(contextDir, config)
	stack = buildStack(ctx, path, dump)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = watcher.Add(contextDir)
	log.Println("Changes Monitor: ", contextDir)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer watcher.Close()

	go waitForInput(ctx, cancel, path, dump)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write && event.Name == path {
					log.Println("File Modified: ", event.Name)
					err = handleReload(ctx, path, dump)
					if err != nil {
						cancel()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")
	if err := stack.Teardown(context.Background()); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func buildStack(ctx context.Context, path string, dump bool) *gbd.Stack {
	env, err := gbd.NewEnvFromConfig(path)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	stack, err := env.Build(ctx, dump)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return stack
}

func handleReload(ctx context.Context, path string, dump bool) error {
	log.Println("Reloading...")
	if err := stack.Teardown(ctx); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	stack = buildStack(ctx, path, dump)
	env, err := gbd.NewEnvFromConfig(path)
	if err != nil {
		return err
	}
	stack, err = env.Build(ctx, false)
	if err != nil {
		return err
	}
	log.Println("Reloaded")
	return nil
}

func waitForInput(ctx context.Context, cancel context.CancelFunc, path string, dump bool) {
	var keystroke string
	for {
		log.Printf("Press 'r' to reload, 'p' to print dev stack, 'q' to quit:\t")
		fmt.Scanln(&keystroke)
		switch keystroke {
		case "r":
			err := handleReload(ctx, path, dump)
			if err != nil {
				cancel()
				return
			}
		case "q":
			cancel()
			return
		case "p":
			log.Printf("Containers: \n")
			log.Print(string(stack.Print()))
		}
	}
}
