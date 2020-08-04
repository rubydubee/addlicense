// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This program ensures source code files have copyright license headers.
// See usage with "addlicense -h".
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

const helpText = `Usage: addlicense [flags] pattern [pattern ...]

The program ensures source code files have copyright license headers
by scanning directory patterns recursively.

It modifies all source files in place and avoids adding a license header
to any file that already has one.

The pattern argument can be provided multiple times, and may also refer
to single files.

Flags:
`

var (
	holder    = flag.String("c", "Google LLC", "copyright holder")
	license   = flag.String("l", "apache", "license type: apache, bsd, mit, mpl")
	licensef  = flag.String("f", "", "license file")
	year      = flag.String("y", fmt.Sprint(time.Now().Year()), "copyright year(s)")
	verbose   = flag.Bool("v", false, "verbose mode: print the name of the files that are modified")
	checkonly = flag.Bool("check", false, "check only mode: verify presence of license headers and exit with non-zero code if missing")
	config    = flag.String("config", "", "yaml config file, see examples/config.yml")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, helpText)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	data := &copyrightData{
		Year:   *year,
		Holder: *holder,
	}

	var t *template.Template
	if *licensef != "" {
		d, err := ioutil.ReadFile(*licensef)
		if err != nil {
			log.Printf("license file: %v", err)
			os.Exit(1)
		}
		t, err = template.New("").Parse(string(d))
		if err != nil {
			log.Printf("license file: %v", err)
			os.Exit(1)
		}
	} else {
		t = licenseTemplate[*license]
		if t == nil {
			log.Printf("unknown license: %s", *license)
			os.Exit(1)
		}
	}

	var c Config
	if *config != "" {
		yamlFile, err := ioutil.ReadFile(*config)
		if err != nil {
			log.Printf("extensions file: %v", err)
			os.Exit(1)
		}
		err = yaml.Unmarshal(yamlFile, &c)
		if err != nil {
			log.Printf("unmarshal: %v", err)
		}
	}

	ignoredPaths := c.IgnorePaths
	fileExtensions := c.FileExtensions
	hasLicensePatterns := c.HasLicensePatterns

	// process at most 1000 files in parallel
	ch := make(chan *file, 1000)
	done := make(chan struct{})
	go func() {
		var wg errgroup.Group
		for f := range ch {
			f := f // https://golang.org/doc/faq#closures_and_goroutines
			wg.Go(func() error {
				if *checkonly {
					// Check if file extension is known
					lic, err := licenseHeader(f.path, t, data, fileExtensions)
					if err != nil {
						log.Printf("%s: %v", f.path, err)
						return err
					}
					// Unknown fileExtension or ignored
					if lic == nil || pathInIgnoredPaths(f.path, ignoredPaths) {
						return nil
					}
					// Check if file has a license
					isMissingLicenseHeader, err := fileHasLicense(f.path, hasLicensePatterns)
					if err != nil {
						log.Printf("%s: %v", f.path, err)
						return err
					}
					if isMissingLicenseHeader {
						fmt.Printf("%s\n", f.path)
						return errors.New("missing license header")
					}
				} else {
					modified, err := addLicense(f.path, f.mode, t, data, ignoredPaths, fileExtensions, hasLicensePatterns)
					if err != nil {
						log.Printf("%s: %v", f.path, err)
						return err
					}
					if *verbose && modified {
						log.Printf("%s modified", f.path)
					}
				}
				return nil
			})
		}
		err := wg.Wait()
		close(done)
		if err != nil {
			os.Exit(1)
		}
	}()

	for _, d := range flag.Args() {
		walk(ch, d)
	}
	close(ch)
	<-done
}

// FileExtensions is a struct used for parsing file-extension commenting patterns
// from a config file.
type FileExtensions []struct {
	Extensions []string `yaml:"extensions"`
	Top        string   `yaml:"top"`
	Mid        string   `yaml:"mid"`
	Bot        string   `yaml:"bot"`
}

// Config is a struct used for parsing a yaml config file.
type Config struct {
	IgnorePaths        []string       `yaml:"ignorePaths"`
	FileExtensions     FileExtensions `yaml:"fileExtensions"`
	HasLicensePatterns []string       `yaml:"hasLicensePatterns"`
}

type file struct {
	path string
	mode os.FileMode
}

var ignoredPathsDefault = []string{"**.git/**", "**node_modules/**", "**.gradle/**"}

func walk(ch chan<- *file, start string) {
	filepath.Walk(start, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("%s error: %v", path, err)
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		if pathInIgnoredPaths(path, ignoredPathsDefault) {
			return nil
		}
		ch <- &file{path, fi.Mode()}
		return nil
	})
}

func addLicense(path string, fmode os.FileMode, tmpl *template.Template,
	data *copyrightData, ignoredPaths []string, extensions FileExtensions,
	hasLicensePatterns []string) (bool, error) {

	if pathInIgnoredPaths(path, ignoredPaths) {
		if *verbose {
			log.Printf("%s: ignored", path)
		}
		return false, nil
	}

	var lic []byte
	var err error

	lic, err = licenseHeader(path, tmpl, data, extensions)
	if lic == nil && err == nil && *verbose {
		log.Printf("%s: file extension not supported", path)
		return false, nil
	}

	if err != nil || lic == nil {
		return false, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil || hasLicense(b, hasLicensePatterns) {
		return false, err
	}

	line := hashBang(b)
	if len(line) > 0 {
		b = b[len(line):]
		if line[len(line)-1] != '\n' {
			line = append(line, '\n')
		}
		lic = append(line, lic...)
	}
	b = append(lic, b...)
	return true, ioutil.WriteFile(path, b, fmode)
}

// fileHasLicense reports whether the file at path contains a license header.
func fileHasLicense(path string, hasLicensePatterns []string) (bool, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil || hasLicense(b, hasLicensePatterns) {
		return false, err
	}
	return true, nil
}

func pathInIgnoredPaths(path string, ignoredPaths []string) bool {
	var g glob.Glob
	for _, p := range ignoredPaths {
		g = glob.MustCompile(p)
		if g.Match(path) {
			return true
		}
	}
	return false
}

func licenseHeader(path string, tmpl *template.Template, data *copyrightData,
	extensions FileExtensions) ([]byte, error) {
	var lic []byte
	var err error
	fileExtension := fileExtension(path)

	// If a custom file-extension configuration was provided, use that.
	if len(extensions) != 0 {
		for _, e := range extensions {
			if stringInSlice(fileExtension, e.Extensions) {
				lic, err = prefix(tmpl, data, e.Top, e.Mid, e.Bot)
				return lic, err
			}
		}
		return nil, nil
	}

	switch fileExtension {
	default:
		return nil, nil
	case ".c", ".h":
		lic, err = prefix(tmpl, data, "/*", " * ", " */")
	case ".js", ".mjs", ".cjs", ".jsx", ".tsx", ".css", ".tf", ".ts":
		lic, err = prefix(tmpl, data, "/**", " * ", " */")
	case ".cc", ".cpp", ".cs", ".go", ".hh", ".hpp", ".java", ".m", ".mm", ".proto", ".rs", ".scala", ".swift", ".dart", ".groovy", ".kt", ".kts":
		lic, err = prefix(tmpl, data, "", "// ", "")
	case ".py", ".sh", ".yaml", ".yml", ".dockerfile", "dockerfile", ".rb", "gemfile":
		lic, err = prefix(tmpl, data, "", "# ", "")
	case ".el", ".lisp":
		lic, err = prefix(tmpl, data, "", ";; ", "")
	case ".erl":
		lic, err = prefix(tmpl, data, "", "% ", "")
	case ".hs", ".sql":
		lic, err = prefix(tmpl, data, "", "-- ", "")
	case ".html", ".xml", ".vue":
		lic, err = prefix(tmpl, data, "<!--", " ", "-->")
	case ".php":
		lic, err = prefix(tmpl, data, "", "// ", "")
	case ".ml", ".mli", ".mll", ".mly":
		lic, err = prefix(tmpl, data, "(**", "   ", "*)")
	}
	return lic, err
}

func fileExtension(name string) string {
	if v := filepath.Ext(name); v != "" {
		return strings.ToLower(v)
	}
	return strings.ToLower(filepath.Base(name))
}

var head = []string{
	"#!",                       // shell script
	"<?xml",                    // XML declaratioon
	"<!doctype",                // HTML doctype
	"# encoding:",              // Ruby encoding
	"# frozen_string_literal:", // Ruby interpreter instruction
	"<?php",                    // PHP opening tag
}

func hashBang(b []byte) []byte {
	var line []byte
	for _, c := range b {
		line = append(line, c)
		if c == '\n' {
			break
		}
	}
	first := strings.ToLower(string(line))
	for _, h := range head {
		if strings.HasPrefix(first, h) {
			return line
		}
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if a == b {
			return true
		}
	}
	return false
}

func hasLicense(b []byte, hasLicensePatterns []string) bool {
	n := 1000
	if len(b) < 1000 {
		n = len(b)
	}

	if len(hasLicensePatterns) != 0 {
		for _, p := range hasLicensePatterns {
			if bytes.Contains(bytes.ToLower(b[:n]), []byte(p)) {
				return true
			}
		}
		return false
	}

	return bytes.Contains(bytes.ToLower(b[:n]), []byte("copyright")) ||
		bytes.Contains(bytes.ToLower(b[:n]), []byte("mozilla public"))
}
