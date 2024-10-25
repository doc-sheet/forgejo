//nolint:forbidigo
package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/ini.v1" //nolint:depguard
)

var (
	policy     *bluemonday.Policy
	tagRemover *strings.Replacer
	safeURL    = "https://TO-BE-REPLACED.COM"

	// Matches href="", href="#", href="%s", href="#%s", href="%[1]s" and href="#%[1]s".
	placeHolderRegex = regexp.MustCompile(`href="#?(%s|%\[\d\]s)?"`)
)

func initBlueMondayPolicy() {
	policy = bluemonday.NewPolicy()

	policy.RequireParseableURLs(true)
	policy.AllowURLSchemes("https")

	// Only allow safe URL on href.
	// Only allow target="_blank".
	// Only allow rel="nopener noreferrer".
	// Only Allow placeholder on id.
	// Only allow "ui sha" as class.
	policy.AllowAttrs("href").Matching(regexp.MustCompile(regexp.QuoteMeta(safeURL))).OnElements("a")
	policy.AllowAttrs("target").Matching(regexp.MustCompile(regexp.QuoteMeta("_blank"))).OnElements("a")
	policy.AllowAttrs("rel").Matching(regexp.MustCompile("noopener|noreferrer|nopener noreferrer")).OnElements("a")
	policy.AllowAttrs("id").Matching(regexp.MustCompile(`%s|%\[\d\]s`)).OnElements("a")
	policy.AllowAttrs("class").Matching(regexp.MustCompile("ui sha")).OnElements("a")

	// Only allow "branch name" as class.
	policy.AllowAttrs("class").Matching(regexp.MustCompile("branch-name")).OnElements("strong")

	// Only allow "branch_target" as id.
	policy.AllowAttrs("id").Matching(regexp.MustCompile("branch_target")).OnElements("code")

	policy.AllowElements("strong", "br", "b", "strike", "code", "i")

	// TODO: Remove <c> in `actions.workflow.dispatch.trigger_found`.
	policy.AllowNoAttrs().OnElements("c")
}

func initRemoveTags() {
	oldnew := []string{}
	for _, el := range []string{
		"email@example.com", "correu@example.com", "epasts@domens.lv", "email@exemplo.com", "eposta@ornek.com", "email@példa.hu", "email@esempio.it",
		"user", "utente", "lietotājs", "gebruiker", "usuário", "Benutzer", "Bruker",
		"server", "servidor", "kiszolgáló", "serveris",
		"label", "etichetta", "etiķete", "rótulo", "Label", "utilizador",
		"filename", "bestandsnaam", "dosyaadi", "fails", "nome do arquivo",
		"c",
	} {
		oldnew = append(oldnew, "<"+el+">", "REPLACED-TAG")
	}

	tagRemover = strings.NewReplacer(oldnew...)
}

func preprocessTranslationValue(value string) string {
	// href should be a parsable URL, replace placeholder strings with a safe url.
	value = placeHolderRegex.ReplaceAllString(value, `href="`+safeURL+`"`)

	// Remove tags that aren't tags but will be parsed as tags. We already know they are safe and sound.
	value = tagRemover.Replace(value)

	// Some translation strings contain escaped HTML characters and some don't. Canonicalize it to unescaped strings.
	value = html.UnescapeString(value)

	return value
}

func checkLocaleFile(localeFile string) []string {
	// Same configuration as Forgejo uses.
	cfg := ini.Empty(ini.LoadOptions{
		IgnoreContinuation: true,
	})
	cfg.NameMapper = ini.SnackCase

	localeContent, err := os.ReadFile(filepath.Join("options", "locale", localeFile))
	if err != nil {
		panic(err)
	}
	if err := cfg.Append(localeContent); err != nil {
		panic(err)
	}

	dmp := diffmatchpatch.New()
	errors := []string{}

	for _, section := range cfg.Sections() {
		for _, key := range section.Keys() {
			var trKey string
			if section.Name() == "" || section.Name() == "DEFAULT" || section.Name() == "common" {
				trKey = key.Name()
			} else {
				trKey = section.Name() + "." + key.Name()
			}

			keyValue := preprocessTranslationValue(key.Value())

			if html.UnescapeString(policy.Sanitize(keyValue)) != keyValue {
				// Create a nice diff of the difference.
				diffs := dmp.DiffMain(keyValue, html.UnescapeString(policy.Sanitize(keyValue)), false)
				diffs = dmp.DiffCleanupSemantic(diffs)
				diffs = dmp.DiffCleanupEfficiency(diffs)

				errors = append(errors, trKey+": "+dmp.DiffPrettyText(diffs))
			}
		}
	}
	return errors
}

func main() {
	initBlueMondayPolicy()
	initRemoveTags()

	localeFiles, err := os.ReadDir(filepath.Join("options", "locale"))
	if err != nil {
		panic(err)
	}

	exitCode := 0

	for _, localeFile := range localeFiles {
		if !strings.HasSuffix(localeFile.Name(), ".ini") {
			continue
		}

		if err := checkLocaleFile(localeFile.Name()); len(err) > 0 {
			fmt.Println(localeFile.Name())
			fmt.Println(strings.Join(err, "\n"))
			fmt.Println()
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
