package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

// migrateGlyphsCmd represents the glyphs command
var migrateGlyphsCmd = &cobra.Command{
	Use:   "glyphs",
	Short: "Migrate the nerd font glyphs in your config",
	Long: `Migrate the nerd font glyphs in your config.

You can choose to print the output to stdout, or migrate your config in the format of your choice.

Example usage

> oh-my-posh config migrate glyphs --config ~/myconfig.omp.json

Migrates the ~/myconfig.omp.json config file's glyphs and prints the result to stdout.

> oh-my-posh config migrate glyphs --config ~/myconfig.omp.json --format toml

Migrates the ~/myconfig.omp.json config file's glyphs and prints the result to stdout in a TOML format.

> oh-my-posh config migrate glyphs --config ~/myconfig.omp.json --format toml --write

Migrates the ~/myconfig.omp.json config file's glyphs and writes the result to your config file in a TOML format.

A backup of the current config can be found at ~/myconfig.omp.json.bak.`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		configFile := config.Path(configFlag)
		cfg, _ := config.Load(configFile, shell.GENERIC, false)

		flags := &runtime.Flags{
			Config: configFile,
		}

		env := &runtime.Terminal{}
		env.Init(flags)
		defer env.Close()

		cfg.MigrateGlyphs = true
		if len(format) == 0 {
			format = cfg.Format
		}

		if write {
			cfg.Backup()
			cfg.Write(format)
			return
		}

		fmt.Print(cfg.Export(format))
	},
}

func init() {
	migrateGlyphsCmd.Flags().BoolVarP(&write, "write", "w", false, "write the migrated config back to the config file")
	migrateGlyphsCmd.Flags().StringVarP(&format, "format", "f", "", "the config format to migrate to")
	migrateCmd.AddCommand(migrateGlyphsCmd)
}
