package build

import (
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
)

type Compiler struct {
	options BuildOptions
}

func NewCompiler(options BuildOptions) *Compiler {
	return &Compiler{options: options}
}

func (c *Compiler) Compile(entryPoint string) (string, error) {
	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{entryPoint},
		Bundle:            true,
		Write:             false,
		Format:            api.FormatIIFE,
		Platform:          api.PlatformBrowser,
		Target:            api.ES2020,
		MinifyWhitespace:  c.options.Minify,
		MinifyIdentifiers: c.options.Minify,
		MinifySyntax:      c.options.Minify,
		TreeShaking:       api.TreeShakingTrue,
		External:          []string{"react", "react-dom", "dashspace-lib"},
		Define: map[string]string{
			"process.env.NODE_ENV": func() string {
				if c.options.Dev {
					return `"development"`
				}
				return `"production"`
			}(),
		},
		Loader: map[string]api.Loader{
			".ts":   api.LoaderTS,
			".tsx":  api.LoaderTSX,
			".js":   api.LoaderJS,
			".jsx":  api.LoaderJSX,
			".json": api.LoaderJSON,
			".css":  api.LoaderCSS,
		},
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Printf("Build error: %s\n", err.Text)
			if err.Location != nil {
				fmt.Printf("   at %s:%d:%d\n", err.Location.File, err.Location.Line, err.Location.Column)
			}
		}
		return "", fmt.Errorf("build failed with %d errors", len(result.Errors))
	}

	if len(result.OutputFiles) == 0 {
		return "", fmt.Errorf("no output generated")
	}

	return string(result.OutputFiles[0].Contents), nil
}
