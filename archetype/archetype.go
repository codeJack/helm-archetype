package archetype

import (
  "fmt"
  "os"
  "path/filepath"
  "regexp"
  "text/template"

  "github.com/Masterminds/sprig"
  iowrap "github.com/spf13/afero"

  "helm.sh/helm/v3/pkg/chart"
)

var (
  FS     iowrap.Fs
  FSUtil *iowrap.Afero
)

func init() {
  FS = iowrap.NewOsFs()
  FSUtil = &iowrap.Afero{Fs: FS}
}

type Archetype struct {
  Chartname string
  Chartpath string
  Values    *map[string]interface{}
}

// New 
func New(chartpath string, vals *map[string]interface{}) *Archetype {
  Archetype := &Archetype{
    Chartname: filepath.Base(chartpath),
    Chartpath: chartpath,
    Values:    vals,
  }

  return Archetype
}

// ChartMetadata creates a new chart metadata with default values
func (e *Archetype) ChartMetadata() *chart.Metadata {

  description := "A Helm chart for Kubernetes"
  version := "0.1.0"
  appVersion := "0.1.0"
  
  if chartMetadata, ok := (*e.Values)["Chart"]; ok {
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
    Name:        e.Chartname,
    Description: description,
    Type:        "application",
    Version:     version,
    AppVersion:  appVersion,
    APIVersion:  chart.APIVersionV2,
  }
}

// Parses and renders Chart`s values and templates in-place
func (e *Archetype) Run() error {

  // Render values file
  vfile := filepath.Join(e.Chartpath, "values.yaml")
  if _, err := FS.Stat(vfile); err == nil {
    err = render(vfile, e.Values)  
    if err != nil {
      return err
    }
  }

  // Render templates
  tdir := filepath.Join(e.Chartpath, "templates")
  if _, err := FS.Stat(tdir); err == nil {

    files, err := FSUtil.ReadDir(tdir)
    if err != nil {
      return err
    }  
    for _, file := range files {
      tfile := filepath.Join(e.Chartpath, "templates", file.Name())

      err = render(tfile, e.Values)
      if err != nil {
        return err
      }

      err = removeIfBlank(tfile)
      if err != nil {
        return err
      }
    }
  }

  return nil
}

func render(file string, vals *map[string]interface{}) error {
  fmt.Printf("Rendering file %s\n", file)

  tpl := template.New("gotpl").Funcs(sprig.TxtFuncMap())
  tpl.Delims("((", "))")
  
  contents, err := FSUtil.ReadFile(file)
  if err != nil {
    return err
  }
  
  f, err := FS.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
  if err != nil {
    return err
  }
  
  _, err = tpl.Parse(string(contents[:]))
  if err != nil {
    return err
  }
  err = tpl.Execute(f, vals)  
  if err != nil {
    return err
  }
  
  if err := f.Close(); err != nil {
    return err
  }

  return nil
}

func removeIfBlank(file string) error {
  // Check the rendered file and eventually delete it if blank
  contents, err := FSUtil.ReadFile(file)
  if err != nil {
    return err
  }

  // Get rid of blank lines
  s := string(contents)
  s = regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`).ReplaceAllString(s, "")

  // No need to keep the template file if it is only made of blanks
  if len(s) == 0 {
    e := FS.Remove(file) 
    if e != nil { 
        return err; 
    }
  }

  return nil
}

