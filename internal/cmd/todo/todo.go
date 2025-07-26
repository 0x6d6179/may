package todo

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdTodo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "find todo, fixme, and hack comments",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := cmd.Flags().GetString("tag")
			if err != nil {
				return err
			}

			return runTodo(f, tags)
		},
	}

	cmd.Flags().StringP("tag", "t", "TODO,FIXME,HACK,XXX,BUG,NOTE", "filter by specific tags (comma-separated)")

	return cmd
}

func runTodo(f *factory.Factory, tagsFlag string) error {
	tagList := strings.Split(strings.ToUpper(tagsFlag), ",")
	tagMap := make(map[string]bool)
	for _, tag := range tagList {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tagMap[tag] = true
		}
	}

	pattern := regexp.MustCompile(`(?i)(TODO|FIXME|HACK|XXX|BUG|NOTE)[:( ]`)

	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		".next":        true,
		"dist":         true,
		"build":        true,
		"__pycache__":  true,
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	type match struct {
		file string
		line int
		tag  string
		msg  string
	}

	var matches []match

	err = filepath.WalkDir(cwd, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if skipDirs[name] || (len(name) > 0 && name[0] == '.') {
				return filepath.SkipDir
			}
			return nil
		}

		if isBinary(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			result := pattern.FindStringSubmatchIndex(line)
			if result == nil {
				continue
			}

			tagText := line[result[2]:result[3]]
			tagUpper := strings.ToUpper(tagText)

			if !tagMap[tagUpper] {
				continue
			}

			msgStart := result[1]
			if msgStart >= len(line) {
				msgStart = len(line)
			}
			msg := strings.TrimSpace(line[msgStart:])

			relPath, _ := filepath.Rel(cwd, path)

			matches = append(matches, match{
				file: relPath,
				line: lineNum,
				tag:  tagUpper,
				msg:  msg,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return nil
	}

	maxFileLen := 0
	maxLineLen := 0
	maxTagLen := 0

	for _, m := range matches {
		if len(m.file) > maxFileLen {
			maxFileLen = len(m.file)
		}
		lineStr := fmt.Sprintf("%d", m.line)
		if len(lineStr) > maxLineLen {
			maxLineLen = len(lineStr)
		}
		if len(m.tag) > maxTagLen {
			maxTagLen = len(m.tag)
		}
	}

	for _, m := range matches {
		lineStr := fmt.Sprintf("%d", m.line)
		fmt.Fprintf(f.IO.Out, "%s:%-*s  %-*s  %s\n",
			m.file,
			maxLineLen,
			lineStr,
			maxTagLen,
			m.tag,
			m.msg,
		)
	}

	return nil
}

func isBinary(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	return bytes.Contains(buf[:n], []byte{0})
}
