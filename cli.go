package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type cliOptions struct {
	outputFile string
	format     string
	keywords   []string
	exts       []string
	outputDir  string
	language   string
}

func runCLI(args []string) error {
	var opts cliOptions
	cmd := &cobra.Command{
		Use:   "goxgettext [path ...]",
		Short: "Extract gettext messages from Go and GoHTML files",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := args
			if len(paths) == 0 {
				paths = []string{"."}
			}
			cfg := newExtractorConfig()
			cfg.keywords = append(cfg.keywords, opts.keywords...)
			if len(opts.exts) > 0 {
				cfg.exts = opts.exts
			}
			var all []message
			for _, path := range paths {
				entries, err := collectMessagesWithConfig(path, cfg)
				if err != nil {
					return err
				}
				all = append(all, entries...)
			}

			var content string
			switch strings.ToLower(opts.format) {
			case "po":
				content = buildPO(all)
			default:
				content = buildPOT(all)
			}

			if opts.outputFile != "" {
				if err := os.WriteFile(opts.outputFile, []byte(content), 0o644); err != nil {
					return err
				}
			}

			if opts.outputDir != "" {
				if err := os.MkdirAll(opts.outputDir, 0o755); err != nil {
					return err
				}
				languages, err := resolveLanguages(paths, opts.language)
				if err != nil {
					return err
				}
				if err := writeLanguageFiles(opts.outputDir, all, languages); err != nil {
					return err
				}
				if err := writePOTFILES(opts.outputDir, paths); err != nil {
					return err
				}
			}

			if opts.outputFile == "" {
				fmt.Print(content)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.outputFile, "output", "o", "", "write the generated catalog to a file")
	cmd.Flags().StringVar(&opts.format, "format", "pot", "output format: pot or po")
	cmd.Flags().StringSliceVar(&opts.keywords, "keyword", []string{}, "additional translation function names")
	cmd.Flags().StringSliceVar(&opts.exts, "extension", []string{}, "additional file extensions to scan (for example .html)")
	cmd.Flags().StringVar(&opts.outputDir, "output-dir", "", "directory for generated language files")
	cmd.Flags().StringVar(&opts.language, "language", "", "language code to generate/update (for example cs)")
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func main() {
	if err := runCLI(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
