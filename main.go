/**
 * Copyright 2022 chyroc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "add-license",
		Description: "add license to source file",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "source", Usage: "which dir or file to run add license(default: .)", DefaultText: "."},
			&cli.StringFlag{Name: "ext", Usage: "which file ext to add license"},
			&cli.StringFlag{Name: "license", Usage: "which license content to add"},
			&cli.StringSliceFlag{Name: "exclude", Usage: "exclude file/dir to add license"},
		},
		Action: func(c *cli.Context) error {
			source := c.String("source")
			exclude := c.StringSlice("exclude")
			ext := c.String("ext")
			license := c.String("license")

			gs := []glob.Glob{}
			for _, v := range exclude {
				g, err := glob.Compile(v)
				if err != nil {
					return err
				}
				gs = append(gs, g)
			}

			bs, err := ioutil.ReadFile(license)
			if err != nil {
				return err
			}

			return walk(source, gs, ext, string(bs), process)
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func walk(source string, exclude Glob, ext string, license string, process func(string, string) error) error {
	f, _ := os.Stat(source)
	if f != nil && !f.IsDir() {
		if exclude.Match(source) {
			return nil
		}

		return process(source, license)
	}

	return filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(info.Name(), ext) {
			return nil
		}
		if exclude.Match(path) {
			return nil
		}
		return process(path, license)
	})
}

func process(file, license string) error {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, []byte(wrapContent(wrapLicense(license), string(bs))), 0666)
}

func wrapLicense(license string) string {
	res := []string{}
	res = append(res, "/**")
	for _, v := range strings.Split(license, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			res = append(res, " *")
		} else {
			res = append(res, " * "+v)
		}
	}
	res = append(res, " */")
	return strings.Join(res, "\n")
}

func wrapContent(license, file string) string {
	if strings.Contains(file, license) {
		return file
	}
	files := strings.Split(file, "\n")
	licenses := strings.Split(license, "\n")
	res := []string{}

	if strings.HasPrefix(file, "// Code generated") {
		res = append(res, files[0])
		files = files[1:]
	}

	res = append(res, licenses...)
	res = append(res, files...)

	return strings.Join(res, "\n")
}

type Glob []glob.Glob

func (r Glob) Match(s string) bool {
	for _, v := range r {
		if v.Match(s) {
			return true
		}
	}
	return false
}
