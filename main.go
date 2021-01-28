package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/chart"
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

  p := getter.All(cli.New())
  vals, err := valuesOpts.MergeValues(p)
  if err != nil {
    return fmt.Errorf("Could not merge values")
  }

  chartname := filepath.Base(o.name)
  cfile := chartMetadata(chartname, vals)

  // Create from the starter
  lstarter := filepath.Join(o.starterDir, o.starter)
  // If path is absolute, we don't want to prefix it with helm starters folder
  if filepath.IsAbs(o.starter) {
    lstarter = o.starter
  }
  
  err = chartutil.CreateFrom(cfile, filepath.Dir(o.name), lstarter)
  if err != nil {
    return fmt.Errorf("Could not create chart %s from starter %s\n", o.name, lstarter)
  }

  // Render values file
  vfile := filepath.Join(o.name, "values.yaml")
  if _, err := os.Stat(vfile); err == nil {
    err = render(vfile, vals)  
    if err != nil {
      return err;
    }
  }

  // Render templates
  tdir := filepath.Join(o.name, "templates")
  if _, err := os.Stat(tdir); err == nil {
    files, err := ioutil.ReadDir(tdir)
    if err != nil {
      return fmt.Errorf("Could not read directory %s\n", tdir)
    }  
    for _, file := range files {
      err = render(filepath.Join(o.name, "templates", file.Name()), vals)
      if err != nil {
        return err;
      }
    }
  }
  return nil
}


// chartMetadata creates a new chart metadata with default values
func chartMetadata(chartname string, vals map[string]interface{}) *chart.Metadata {

  description := "A Helm chart for Kubernetes"
  version := "0.1.0"
  appVersion := "0.1.0"

  if chartMetadata, ok := vals["Chart"]; ok {
    chartMetadata := chartMetadata.(map[string]interface{})
    if field, ok := chartMetadata["description"]; ok {
      description = field.(string)
    }
    if field, ok := chartMetadata["version"]; ok {
      version = field.(string)
    }
    if field, ok := chartMetadata["appVersion"]; ok {
      appVersion = field.(string)
    }
  }

  return &chart.Metadata{
    Name:        chartname,
    Description: description,
    Type:        "application",
    Version:     version,
    AppVersion:  appVersion,
    APIVersion:  chart.APIVersionV2,
  }
}

func render(file string, vals map[string]interface{}) error {
  fmt.Printf("Rendering file %s\n", file)
  
  tpl := template.New("gotpl").Funcs(sprig.TxtFuncMap())
  tpl.Delims("((", "))")

  contents, err := ioutil.ReadFile(file)
  if err != nil {
    return fmt.Errorf("Could not read contents of file %s\n", file)
  }

  f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
  if err != nil {
    return fmt.Errorf("Could not open file %s for writing\n", file)
  }

  _, err = tpl.Parse(string(contents))
  if err != nil {
    return fmt.Errorf("Could not parse file %s contents\n", file)
  }
  err = tpl.Execute(f, vals)  
  if err != nil {
    return fmt.Errorf("Could not render file %s\n", file)
  }

  if err := f.Close(); err != nil {
    return fmt.Errorf("Could close file %s after rendering\n", file)
  }

  return nil
}
