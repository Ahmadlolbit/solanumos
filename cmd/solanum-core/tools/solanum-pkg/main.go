package pkg

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const usage = `Usage: solanum-pkg install|remove
       solanum-pkg search [WORDS] [--json]
       solanum-pkg installed [--json]

install    search every Debian package and install the ones you pick
remove     search the installed packages and remove the ones you pick
search     list the packages whose name or description has every word, best first,
           at most 200; with no words, every package
installed  list the installed packages

Tab picks several, Enter goes ahead, Esc leaves without changing anything.
`

const (
	shown    = 200
	misses   = -1
	jsonFlag = "--json"
)

type listing struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
}

func pick(prompt string, names []string) []string {
	output, _ := run.Filter(strings.Join(run.Sorted(names), "\n"), "fzf", "--multi", "--layout", "reverse", "--prompt", prompt+" > ",
		"--header", "Tab picks several, Enter goes ahead, Esc leaves",
		"--preview", `apt-cache show {1} 2>/dev/null | awk "NF == 0 { exit } { print }"`,
		"--preview-window", "right,60%,wrap")
	return run.Lines(output)
}

func available() []string {
	output, _ := run.Capture(run.Slow, "apt-cache", "pkgnames")
	return run.Lines(output)
}

func installedPackages() []listing {
	output, _ := run.Capture(run.Slow, "dpkg-query", "-W", "-f=${db:Status-Status}\t${Package}\t${binary:Summary}\n")
	var found []listing
	seen := map[string]bool{}
	for _, line := range run.Lines(output) {
		fields := append(strings.SplitN(line, "\t", 3), "", "")
		if fields[0] == "installed" && fields[1] != "" && !seen[fields[1]] {
			seen[fields[1]] = true
			found = append(found, listing{fields[1], fields[2], true})
		}
	}
	slices.SortFunc(found, func(a, b listing) int { return cmp.Compare(a.Name, b.Name) })
	return found
}

func installed() []string {
	var names []string
	for _, each := range installedPackages() {
		names = append(names, each.Name)
	}
	return names
}

func parsed(output string, present map[string]bool) []listing {
	var found []listing
	seen := map[string]bool{}
	for _, line := range run.Lines(output) {
		name, description, ok := strings.Cut(line, " - ")
		if ok && name != "" && !seen[name] {
			seen[name] = true
			found = append(found, listing{name, description, present[name]})
		}
	}
	slices.SortFunc(found, func(a, b listing) int { return cmp.Compare(a.Name, b.Name) })
	return found
}

func every() []listing {
	output, _ := run.Capture(run.Slow, "apt-cache", "search", ".")
	present := map[string]bool{}
	for _, each := range installedPackages() {
		present[each.Name] = true
	}
	return parsed(output, present)
}

func rank(each listing, words []string) int {
	description := strings.ToLower(each.Description)
	score := 0
	for _, word := range words {
		switch {
		case each.Name == word:
		case strings.HasPrefix(each.Name, word):
			score++
		case strings.Contains(each.Name, word):
			score += 2
		case strings.Contains(description, word):
			score += 3
		default:
			return misses
		}
	}
	return score
}

func matching(all []listing, words []string) []listing {
	if len(words) == 0 {
		return all
	}
	scores := map[string]int{}
	var found []listing
	for _, each := range all {
		if score := rank(each, words); score != misses {
			scores[each.Name] = score
			found = append(found, each)
		}
	}
	slices.SortStableFunc(found, func(a, b listing) int { return cmp.Compare(scores[a.Name], scores[b.Name]) })
	return found[:min(len(found), shown)]
}

func wordsOf(args []string) ([]string, bool) {
	var words []string
	asJSON := false
	for _, arg := range args {
		if arg == jsonFlag {
			asJSON = true
			continue
		}
		words = append(words, strings.Fields(strings.ToLower(arg))...)
	}
	return words, asJSON
}

func show(found []listing, asJSON bool) int {
	if asJSON {
		data, err := json.Marshal(append([]listing{}, found...))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(string(data))
		return 0
	}
	for _, each := range found {
		mark := " "
		if each.Installed {
			mark = "*"
		}
		fmt.Printf("%s %s - %s\n", mark, each.Name, each.Description)
	}
	return 0
}

func change(action, prompt string, names []string) int {
	packages := pick(prompt, names)
	if len(packages) == 0 {
		return 0
	}
	return run.Status("sudo", append([]string{"apt-get", action}, packages...)...)
}

func Main() {
	switch env.Arg(1) {
	case "install":
		os.Exit(change("install", "Install", available()))
	case "remove":
		os.Exit(change("remove", "Remove", installed()))
	case "search":
		words, asJSON := wordsOf(os.Args[2:])
		os.Exit(show(matching(every(), words), asJSON))
	case "installed":
		_, asJSON := wordsOf(os.Args[2:])
		os.Exit(show(installedPackages(), asJSON))
	case "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
