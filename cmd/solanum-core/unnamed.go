package main

import (
	"cmp"
	"net/url"
	"os"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

const defaultMirror = "deb.debian.org"

var genericNames = []string{"debian", "solanum", "localhost", "user", "admin", "root", "live", "test", "guest", "demo", "default", "linux", "computer", "laptop", "desktop", "pc", "home", "dev"}

type masked struct {
	word, with string
	path       bool
}

func telling(name string) bool {
	return utf8.RuneCountInString(name) > 2 && !slices.Contains(genericNames, strings.ToLower(name))
}

func wordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func goesOn(rest string, path bool) bool {
	next, size := utf8.DecodeRuneInString(rest)
	if path && next == '.' {
		after, _ := utf8.DecodeRuneInString(rest[size:])
		return wordRune(after)
	}
	return rest != "" && (wordRune(next) || path && next == '-')
}

func cameAfter(text string, path bool) bool {
	last, _ := utf8.DecodeLastRuneInString(text)
	return text != "" && (wordRune(last) || path && strings.ContainsRune("-.", last))
}

func replaceWhole(text string, mask masked) string {
	var out strings.Builder
	for {
		at := strings.Index(text, mask.word)
		if at < 0 {
			out.WriteString(text)
			return out.String()
		}
		end := at + len(mask.word)
		out.WriteString(text[:at])
		if cameAfter(text[:at], mask.path) || goesOn(text[end:], mask.path) {
			out.WriteString(mask.word)
		} else {
			out.WriteString(mask.with)
		}
		text = text[end:]
	}
}

func mirrorHost() string {
	address, err := url.Parse(config.Read(config.Path()).Get("DEBIAN_MIRROR"))
	if err != nil || address.Hostname() == defaultMirror {
		return ""
	}
	return address.Hostname()
}

func masks(person account.Account) []masked {
	host, _ := os.Hostname()
	written := strings.TrimSpace(fsutil.Optional(env.Or("SOLANUM_HOSTNAME_FILE", "/etc/hostname")))
	var found []masked
	if len(person.Home) > 1 {
		found = append(found, masked{person.Home, "~", true})
	}
	if mirror := mirrorHost(); mirror != "" {
		found = append(found, masked{mirror, "(your mirror)", true})
	}
	var names []masked
	for _, name := range []masked{{word: written, with: "(this computer)"}, {word: host, with: "(this computer)"}, {word: person.Name, with: "(you)"}} {
		if telling(name.word) && !slices.ContainsFunc(names, func(each masked) bool { return each.word == name.word }) {
			names = append(names, name)
		}
	}
	slices.SortStableFunc(names, func(one, other masked) int { return cmp.Compare(len(other.word), len(one.word)) })
	return append(found, names...)
}

func unnamed(person account.Account) func(string) string {
	found := masks(person)
	return func(text string) string {
		for _, mask := range found {
			text = replaceWhole(text, mask)
		}
		return text
	}
}
