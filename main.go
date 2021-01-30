package main

import (
  "os"
  "path/filepath"
  
  "github.com/codeJack/helm-archetype/archetype"
  "github.com/spf13/cobra"

  "helm.sh/helm/v3/cmd/helm/require"
  "helm.sh/helm/v3/pkg/chartutil"
  "helm.sh/helm/v3/pkg/cli"
  "helm.sh/helm/v3/pkg/cli/values"
  "helm.sh/helm/v3/pkg/getter"
  "helm.sh/helm/v3/pkg/helmpath"
)

type archetypeOptions struct {
  name       string
  starter    string
  starterDir string
}

func main() {
  o := &archetypeOptions{}
  valuesOpts := &values.Options{}

  cmd := &cobra.Command{
    Use:   "archetype [NAME] [STARTER] [flags]",
    Short: "create a new Helm chart from a templated starter scaffold",
    Args:  require.ExactArgs(2),
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
      if len(args) == 0 {
        // Allow file completion when completing the argument for the name
        // which could be a path
        return nil, cobra.ShellCompDirectiveDefault
      }
      // No more completions, so disable file completion
      return nil, cobra.ShellCompDirectiveNoFileComp
    },
    RunE: func(cmd *cobra.Command, args []string) error {
      o.name = args[0]
      o.starter = args[1]
      o.starterDir = helmpath.DataPath("starters")
      return o.run(valuesOpts)
    },
  }

  cmd.Flags().StringSliceVarP(&valuesOpts.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
  cmd.Flags().StringArrayVar(&valuesOpts.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
  cmd.Flags().StringArrayVar(&valuesOpts.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
  cmd.Flags().StringArrayVar(&valuesOpts.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
  if err := cmd.Execute(); err != nil {
    os.Exit(1)
  }
}

func (o *archetypeOptions) run(valuesOpts *values.Options) error {

  // Create from the starter
  lstarter := filepath.Join(o.starterDir, o.starter)
  // If path is absolute, we don't want to prefix it with helm starters folder
  if filepath.IsAbs(o.starter) {
    lstarter = o.starter
  }

  p := getter.All(cli.New())
  vals, err := valuesOpts.MergeValues(p)
  if err != nil {
    return err
  }

  archetype := archetype.New(o.name, &vals)
  cfile := archetype.ChartMetadata()
  err = chartutil.CreateFrom(cfile, filepath.Dir(o.name), lstarter)
  if err != nil {
    return err
  }

  return archetype.Run()
}


