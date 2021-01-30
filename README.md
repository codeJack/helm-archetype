# Helm Archetype Plugin

This plugin extends Helm's built-in *create from* capabilities by allowing **templated** starters. Such templated starters get rendered at *creation time* as if their `helm template` output got dumped into a brand new chart.

Such an approach aims at reducing the overall complexity of helm charts, by shifting left part of the rendering logic and narrowing down their scope, yet preserving the flexibility to later re-assess such a scope through repeated creation.  

To distinguish between *creation time* templates and *install time* templates, *creation time* templates are to be delimited by double **round brackets** as opposed to double **curly brackets**

```yaml
# Somewhere within the starter templates
((- if .Values.to.be.rendered.at.creation.time ))
version: {{ .Values.to.be.rendered.at.install.time }}
((- end ))
```

Starter's `values.yaml` can also contain *creation time* templates.

The *values* to be provided at `helm archetype` time are structured with two root nodes : `Chart` and `Values`.
- `Chart` node will provide *Chart.yaml* metadata (apart from the chart name which is provided as a positional argument to the plug-in) 
- `Values` node will provide the necessary values for the rendering exercise

Additionally, as this plugin extends Helm's built-in *create from* capabilities, the `<CHARTNAME>` value can also be used within the starter, and, quoting [Helm documentation](https://helm.sh/docs/topics/charts/) *"All occurrences of `<CHARTNAME>` will be replaced with the specified chart name so that starter charts can be used as templates."*

```yaml
Chart:
  description: "Helm archetype chart"
  version: "0.2.1"
  appVersion: "0.3.0"

Values:
  version: "0.1.0"
```

Template files which would result as empty (or blank) past the rendering exercise will be deleted, in order to guarantee the thinner possible chart structure.

## Usage

Create a new chart from a templated starter scaffold.

```
$ helm archetype [NAME] [STARTER] [flags]
```

### Flags:

```
     --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
     --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
     --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
 -f, --values strings           specify values in a YAML file or a URL (can specify multiple)
```


## Install

```
$ helm plugin install https://github.com/codeJack/helm-archetype
```

The above will fetch the latest binary release of `helm archetype` and install it.

In case the remote `helm plugin install` fails you can clone this repository and rely on `make install`.