package main

import "fmt"
import "os"
import "io/ioutil"
import "strings"
import "flag"

var expandTabsFlag = flag.Bool("e", false, "Expand tabs into spaces")
var contractTabsFlag = flag.Bool("c", false, "Contract spaces into tabs")
var tabSize = flag.Int("ts", 4, "Size of tabs")
var trailingNewlineFlag = flag.Bool("t", false, "Remove trailing newlines")
var addTrailingNewlineFlag = flag.Bool("at", false, "Add trailing newline")
var helpFlag = flag.Bool("h", false, "Display usage")

func isText(filename string) bool {
	f, _ := os.Open(filename, 0, 0)
	contents := make([]uint8, 1024)
	n, _ := f.Read(contents)

	for i := 0; i < 1024 && i < n; i++ {
		if contents[i] > 0x7e || contents[i] < 0x09 {
			return false
		}
	}
	return true
}

func expandTabs(str string) (result string) {
	start := true
	for i := range str {
		if start && str[i] == '\t' {
			result += strings.Repeat(" ", *tabSize)
		} else {
			start = false
			result += str[i:i+1]
		}
	}
	return
}

func chompTab(str string) string {
	if len(str) < *tabSize {
		return str
	}
	tab := strings.Repeat(" ", *tabSize)

	if strings.HasPrefix(str, tab) {
		return str[*tabSize:]
	}
	return str
}

func contractTabs(str string) string {
	tabsChomped := 0
	for {
		old_len := len(str)
		str = chompTab(str)
		if len(str) != old_len {
			tabsChomped++
		} else {
			break
		}
	}
	return strings.Repeat("\t", tabsChomped) + str
}

func removeTrailingNewlines(lines []string) []string {
	empties := 0
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] == "" {
			empties++
		} else {
			break
		}
	}

	return lines[:len(lines) - empties]
}

func cleanFile(filename string) (trimmed, tabs int) {
	contents, _ := ioutil.ReadFile(filename)

	if len(contents) == 0 {
		return
	}

	// Chomp the last newline, because split creates an extra blank after it.
	if contents[len(contents)-1] == '\n' {
		contents = contents[:len(contents)-1]
	}


	lines := strings.Split(string(contents), "\n", -1)

	if *addTrailingNewlineFlag && lines[len(lines)-1] != "" {
		lines = append(lines, "")
		fmt.Printf("Added trailing newline.\n")
	}

	if *trailingNewlineFlag {
		l := len(lines)
		lines = removeTrailingNewlines(lines)
		if len(lines) != l {
			fmt.Printf("Removed trailing newlines.\n")
		}
	}

	output := ""
	for _, str := range lines {
		ts := strings.TrimRight(str, " \t")
		if len(ts) < len(str) {
			trimmed++
		}
		lts := len(ts)
		switch {
			case *contractTabsFlag:
				ts = contractTabs(ts)
				if len(ts) != lts {
					tabs++
				}
				break
			case *expandTabsFlag:
				ts = expandTabs(ts)
				if len(ts) != lts {
					tabs++
				}
				break
		}
		output += ts + "\n"
	}
	ioutil.WriteFile(filename, []uint8(output), 0666)
	return
}

func pluralize(s string, n int) string {
	if n == 1 {
		return s
	}
	return s + "s"
}

func isRegular(fn string) bool {
	s, err := os.Stat(fn)
	if err != nil {
		return false
	}
	return s.IsRegular()
}

func logModifications(fn string, trims, tabs int) {
	fmt.Printf("Fixed %d %s", trims, pluralize("line", trims))
	if tabs > 0 {
		fmt.Printf(" and %d %s", tabs, pluralize("tab", tabs))
	}
	if fn != "" {
		fmt.Printf(" in %s", fn)
	}
	fmt.Println()
}

func processFile(fn string) (trims, tabs int) {
	if !isRegular(fn) {
		fmt.Printf("Couldn't clean %s\n", fn)
		return
	}
	if isText(fn) {
		trims, tabs = cleanFile(fn)
	} else {
		fmt.Printf("Didn't clean binary file %s\n", fn)
	}
	return
}

func main() {
	flag.Parse()
	if *helpFlag {
		flag.Usage()
		return
	}
	if *contractTabsFlag && *expandTabsFlag {
		fmt.Println("Can't contract and expand tabs.")
		os.Exit(1)
	}

	trims, tabs := 0, 0
	if len(flag.Args()) == 0 {
		fmt.Println("No files to work on.")
		os.Exit(0)
	}
	for _, fn := range flag.Args() {
		trs, tas := processFile(fn)
		logModifications(fn, trs, tas)
		trims += trs
		tabs += tas
	}

	logModifications("", trims, tabs)
}
