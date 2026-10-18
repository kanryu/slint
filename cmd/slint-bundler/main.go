// slint-bundler bundles every Slint source below a directory into one deterministic
// Slint source file. It is the Hike equivalent of Knuckle's bundleSlint.ts.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type directoryEntry struct {
	name      string
	directory bool
}

func readDirectory(directory string) []directoryEntry {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return make([]directoryEntry, 0)
	}
	result := make([]directoryEntry, 0)
	for _, entry := range entries {
		result = append(result, directoryEntry{name: entry.Name(), directory: entry.IsDir()})
	}
	return result
}

type slintFile struct {
	path         string
	relative     string
	source       string
	std_widgets  []string
	definitions  []string
	dependencies []int
	visited      bool
	visiting     bool
}

func usage() {
	fmt.Printf("Usage: slint-bundler <slint-directory> [-o output.slint]\n")
	fmt.Printf("       slint-bundler --help\n")
}

func indexOf(value string, needle string) int {
	if len(needle) == 0 {
		return 0
	}
	if len(needle) > len(value) {
		return -1
	}
	for i := 0; i <= len(value)-len(needle); i = i + 1 {
		if value[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

func hasPrefix(value string, prefix string) bool {
	return len(prefix) <= len(value) && value[0:len(prefix)] == prefix
}

func trimSpace(value string) string {
	start := 0
	end := len(value)
	for start < end && (value[start] == 32 || value[start] == 9 || value[start] == 13 || value[start] == 10) {
		start = start + 1
	}
	for end > start && (value[end-1] == 32 || value[end-1] == 9 || value[end-1] == 13 || value[end-1] == 10) {
		end = end - 1
	}
	return value[start:end]
}

func splitLines(value string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(value); i = i + 1 {
		if value[i] == 10 {
			result = append(result, value[start:i])
			start = i + 1
		}
	}
	result = append(result, value[start:len(value)])
	return result
}

func fields(value string) []string {
	result := make([]string, 0)
	start := -1
	for i := 0; i < len(value); i = i + 1 {
		space := value[i] == 32 || value[i] == 9 || value[i] == 13 || value[i] == 10
		if !space && start < 0 {
			start = i
		}
		if space && start >= 0 {
			result = append(result, value[start:i])
			start = -1
		}
	}
	if start >= 0 {
		result = append(result, value[start:len(value)])
	}
	return result
}

func joinLines(lines []string) string {
	result := ""
	for i := 0; i < len(lines); i = i + 1 {
		if i > 0 {
			result = result + "\n"
		}
		result = result + lines[i]
	}
	return result
}

func lowerASCII(value string) string {
	result := ""
	for i := 0; i < len(value); i = i + 1 {
		c := value[i]
		if c >= 65 && c <= 90 {
			c = c + 32
		}
		result = result + string([]byte{c})
	}
	return result
}

func sortStrings(values []string) {
	sort.Strings(values)
}

func joinPath(parent string, child string) string {
	if parent == "" || parent == "." {
		return child
	}
	return filepath.Join(parent, child)
}

func hasSuffix(value string, suffix string) bool {
	if len(suffix) > len(value) {
		return false
	}
	return value[len(value)-len(suffix):len(value)] == suffix
}

func isWordByte(value byte) bool {
	return (value >= 65 && value <= 90) ||
		(value >= 97 && value <= 122) ||
		(value >= 48 && value <= 57) || value == 95
}

func containsWord(source string, word string) bool {
	if word == "" {
		return false
	}
	start := 0
	for {
		found := indexOf(source[start:len(source)], word)
		if found < 0 {
			return false
		}
		at := start + found
		beforeOK := at == 0 || !isWordByte(source[at-1])
		after := at + len(word)
		afterOK := after >= len(source) || !isWordByte(source[after])
		if beforeOK && afterOK {
			return true
		}
		start = after
		if start >= len(source) {
			return false
		}
	}
}

func firstDefinitionName(line string) string {
	marker := "component "
	at := indexOf(line, marker)
	if at < 0 {
		marker = "global "
		at = indexOf(line, marker)
	}
	if at < 0 {
		return ""
	}
	start := at + len(marker)
	end := start
	for end < len(line) {
		c := line[end]
		if c == 32 || c == 9 || c == 123 {
			break
		}
		end = end + 1
	}
	return line[start:end]
}

func stripImports(source string) string {
	lines := splitLines(source)
	kept := make([]string, 0)
	for i := 0; i < len(lines); i = i + 1 {
		trimmed := trimSpace(lines[i])
		if hasPrefix(trimmed, "import ") {
			// Imports are reconstructed once for the complete bundle. Keeping
			// one here would leave a path relative to the original source
			// file and could also duplicate the standard-widget import.
			continue
		}
		kept = append(kept, lines[i])
	}
	return trimSpace(joinLines(kept))
}

func splitComma(value string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i <= len(value); i = i + 1 {
		if i == len(value) || value[i] == 44 {
			item := widgetName(stripUtf8Bom(trimSpace(value[start:i])))
			if item != "" {
				result = append(result, item)
			}
			start = i + 1
		}
	}
	return result
}

func stripUtf8Bom(value string) string {
	if len(value) >= 3 && value[0] == 239 && value[1] == 187 && value[2] == 191 {
		return value[3:len(value)]
	}
	return value
}

func widgetName(value string) string {
	result := ""
	for i := 0; i < len(value); i = i + 1 {
		c := value[i]
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) {
			result = result + string([]byte{c})
		}
	}
	return result
}

func collectStdWidgets(source string) []string {
	result := make([]string, 0)
	lines := splitLines(source)
	for i := 0; i < len(lines); i = i + 1 {
		line := trimSpace(lines[i])
		if indexOf(line, "std-widgets.slint") < 0 {
			continue
		}
		fmt.Printf("slint-bundler: std-widgets import: %s\n", line)
		left := indexOf(line, "{")
		right := indexOf(line, "}")
		if left < 0 || right <= left {
			continue
		}
		names := splitComma(line[left+1 : right])
		for n := 0; n < len(names); n = n + 1 {
			duplicate := false
			for e := 0; e < len(result); e = e + 1 {
				if result[e] == names[n] {
					duplicate = true
					break
				}
			}
			if !duplicate {
				result = append(result, names[n])
			}
		}
	}
	return result
}

func readSlintFile(path string, relative string) *slintFile {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	sourceText := stripUtf8Bom(string(data))
	if sourceText == "" {
		return nil
	}
	stdWidgets := collectStdWidgets(sourceText)
	source := stripImports(sourceText)
	definitions := make([]string, 0)
	lines := splitLines(source)
	for i := 0; i < len(lines); i = i + 1 {
		name := firstDefinitionName(trimSpace(lines[i]))
		if name != "" {
			definitions = append(definitions, name)
		}
	}
	return &slintFile{path: path, relative: relative, source: source, std_widgets: stdWidgets, definitions: definitions, dependencies: make([]int, 0)}
}

func collectSlintFiles(directory string, relativeDirectory string, excluded string, files *[]slintFile) bool {
	entries := readDirectory(directory)
	names := make([]string, 0)
	for i := 0; i < len(entries); i = i + 1 {
		names = append(names, entries[i].name)
	}
	sortStrings(names)

	for i := 0; i < len(names); i = i + 1 {
		name := names[i]
		if name == ".git" || name == "node_modules" {
			continue
		}
		path := joinPath(directory, name)
		relative := name
		if relativeDirectory != "" {
			relative = relativeDirectory + "/" + name
		}
		isDirectory := false
		for e := 0; e < len(entries); e = e + 1 {
			if entries[e].name == name {
				isDirectory = entries[e].directory
				break
			}
		}
		if isDirectory {
			collectSlintFiles(path, relative, excluded, files)
			continue
		}
		if !hasSuffix(lowerASCII(name), ".slint") {
			continue
		}
		if path == excluded {
			continue
		}
		file := readSlintFile(path, relative)
		if file != nil {
			*files = append(*files, *file)
		}
	}
	return true
}

func buildDependencies(files []slintFile) {
	for i := 0; i < len(files); i = i + 1 {
		for j := 0; j < len(files); j = j + 1 {
			if i == j {
				continue
			}
			found := false
			for d := 0; d < len(files[j].definitions); d = d + 1 {
				if indexOf(files[i].source, files[j].definitions[d]) >= 0 {
					found = true
					break
				}
			}
			if found {
				fmt.Printf("slint-bundler: dependency %s -> %s\n", files[i].relative, files[j].relative)
				files[i].dependencies = append(files[i].dependencies, j)
			}
		}
	}
}

func dependencyIndices(files []slintFile, index int) []int {
	result := make([]int, 0)
	for j := 0; j < len(files); j = j + 1 {
		if index == j {
			continue
		}
		found := false
		for d := 0; d < len(files[j].definitions); d = d + 1 {
			if indexOf(files[index].source, files[j].definitions[d]) >= 0 {
				found = true
				break
			}
		}
		if found {
			result = append(result, j)
		}
	}
	return result
}

func visitFile(files []slintFile, index int, order *[]int, visited []bool, visiting []bool) {
	if visited[index] || visiting[index] {
		return
	}
	visiting[index] = true
	dependencies := dependencyIndices(files, index)
	for i := 0; i < len(dependencies); i = i + 1 {
		visitFile(files, dependencies[i], order, visited, visiting)
	}
	visiting[index] = false
	visited[index] = true
	*order = append(*order, index)
}

func hasImportedWidget(files []slintFile, wanted string) bool {
	for i := 0; i < len(files); i = i + 1 {
		for j := 0; j < len(files[i].std_widgets); j = j + 1 {
			if widgetName(files[i].std_widgets[j]) == wanted {
				return true
			}
		}
	}
	return false
}

func anyFileContains(files []slintFile, name string) bool {
	for i := 0; i < len(files); i = i + 1 {
		if containsWord(files[i].source, name) {
			return true
		}
	}
	return false
}

func standardWidgetImport(files []slintFile) string {
	imported := make(map[string]bool)
	for i := 0; i < len(files); i = i + 1 {
		for n := 0; n < len(files[i].std_widgets); n = n + 1 {
			name := widgetName(files[i].std_widgets[n])
			imported[name] = true
		}
	}
	names := make([]string, 0)
	for name, _ := range imported {
		names = append(names, name)
	}
	sortStrings(names)
	if len(names) == 0 {
		return ""
	}
	used := names[0]
	for i := 1; i < len(names); i = i + 1 {
		used = used + ", " + names[i]
	}
	return "import { " + used + " } from \"std-widgets.slint\";"
}

func bundle(files []slintFile) string {
	order := make([]int, 0)
	visited := make([]bool, len(files))
	visiting := make([]bool, len(files))
	for i := 0; i < len(files); i = i + 1 {
		visitFile(files, i, &order, visited, visiting)
	}
	result := standardWidgetImport(files)
	for i := 0; i < len(order); i = i + 1 {
		file := files[order[i]]
		if file.source == "" {
			continue
		}
		result = result + "\n\n// --- " + file.relative + " ---\n" + file.source
	}
	return result + "\n"
}

func run() int {
	input := ""
	output := "appbundle.slint"
	args := os.Args
	for i := 1; i < len(args); i = i + 1 {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			usage()
			return 0
		}
		if arg == "-o" || arg == "--output" {
			if i+1 >= len(args) {
				usage()
				return 2
			}
			output = args[i+1]
			i = i + 1
			continue
		}
		if hasPrefix(arg, "-") {
			fmt.Printf("slint-bundler: unknown option %s\n", arg)
			usage()
			return 2
		}
		if input != "" {
			fmt.Printf("slint-bundler: only one input directory is supported\n")
			return 2
		}
		input = arg
	}
	if input == "" {
		usage()
		return 2
	}

	files := make([]slintFile, 0)
	collectSlintFiles(input, "", output, &files)
	if len(files) == 0 {
		fmt.Printf("slint-bundler: no .slint files found below %s\n", input)
		return 1
	}
	result := bundle(files)
	if err := os.WriteFile(output, []byte(result), 0644); err != nil {
		fmt.Printf("slint-bundler: cannot write %s\n", output)
		return 1
	}
	fmt.Printf("slint-bundler: bundled %d Slint files into %s\n", len(files), output)
	return 0
}

func main() { os.Exit(run()) }
