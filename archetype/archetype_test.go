package archetype_test

import (
  "fmt"
  "testing"

  . "github.com/codeJack/helm-archetype/archetype"
  iowrap "github.com/spf13/afero"
  "github.com/stretchr/testify/assert"

  "sigs.k8s.io/yaml"
)

func init() {
  // In-memory FS to ease unit testing
  FS = iowrap.NewMemMapFs()
  FSUtil = &iowrap.Afero{Fs: FS}
}

func TestNew(t *testing.T) {

  vals := make(map[string]interface{})
  archetype := New("/chartname", &vals)

  assert.Equal(t, "chartname", archetype.Chartname)
  assert.Equal(t, "/chartname", archetype.Chartpath)
  assert.Equal(t, vals, *archetype.Values)
}

func TestChartMetadata_defaults(t *testing.T) {
  
  vals := make(map[string]interface{})
  archetype := New("chartname", &vals)
  cmetadata:= archetype.ChartMetadata()

  assert.Equal(t, "A Helm chart for Kubernetes", cmetadata.Description, "The chart metadata description does not match.")
  assert.Equal(t, "0.1.0", cmetadata.Version, "The chart metadata version does not match.")
  assert.Equal(t, "0.1.0", cmetadata.AppVersion, "The chart metadata appVersion does not match.")
}

func TestChartMetadata_override(t *testing.T) {

  description := "This is a Chart description"
  version := "1.2.3"
  appVersion := "3.2.1"

  data := fmt.Sprintf(`
Chart: 
  description: "%s"
  version: "%s"
  appVersion: "%s"`, description, version, appVersion)
  
  vals := fromYAML(data)

  archetype := New("chartname", &vals)
  cmetadata:= archetype.ChartMetadata()

  assert.Equal(t, description, cmetadata.Description, "The chart metadata description does not match.")
  assert.Equal(t, version, cmetadata.Version, "The chart metadata version does not match.")
  assert.Equal(t, appVersion, cmetadata.AppVersion, "The chart metadata appVersion does not match.")
}

func TestRun(t *testing.T) {
  values := `
Values: 
  template:
    type: scaffold
    renders: now
  enable: true`

  cvalues := `
somevar: true
some:
  templated:
    var: (( .Values.enable ))`

  renderedValues := `
somevar: true
some:
  templated:
    var: true`

  template := `
This is a ((  .Values.template.type )) template, that renders (( .Values.template.renders )).
This is a {{  .Values.template.type }} template, that renders {{ .Values.template.renders }}.`
    
  renderedTemplate := `
This is a scaffold template, that renders now.
This is a {{  .Values.template.type }} template, that renders {{ .Values.template.renders }}.`

  vfilename:= "/test-chart/values.yaml"
  tfilename:= "/test-chart/templates/template.yaml"

  vals := fromYAML(values)

  FS.MkdirAll("/test-chart/templates", 0644)
  iowrap.WriteFile(FS, vfilename, []byte(cvalues), 0644)
  iowrap.WriteFile(FS, tfilename, []byte(template), 0644)

  archetype := New("/test-chart", &vals)
  assert.Nil(t, archetype.Run())

  contents, _ := FSUtil.ReadFile(vfilename)
  assert.Equal(t, renderedValues, string(contents[:]), "The rendered values do not match.")  
  contents, _ = FSUtil.ReadFile(tfilename)
  assert.Equal(t, renderedTemplate, string(contents[:]), "The rendered template does not match.")  
}

func TestRun_removeEmptyTemplate(t *testing.T) {
  values := `
Values: 
  template:
    enabled: false`  
    
  template := `
   

    ((- if  .Values.template.enabled )) 
This is a conditional scaffold template
((- end ))
   

`

  filename := "/test-chart/templates/template.yaml"

  vals := fromYAML(values)

  FS.MkdirAll("/test-chart/templates", 0644)
  iowrap.WriteFile(FS, filename, []byte(template), 0644)

  archetype := New("/test-chart", &vals)
  assert.Nil(t, archetype.Run())

  assert.NoFileExists(t, filename)
}

// From helm sources https://github.com/helm/helm/blob/v3.5.1/pkg/engine/funcs.go#L98
func fromYAML(str string) map[string]interface{} {
  m := map[string]interface{}{}
  if err := yaml.Unmarshal([]byte(str), &m); err != nil {
    m["Error"] = err.Error()
  }
  return m
}