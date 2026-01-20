package database

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type Rule struct {
	ID          int
	Description string
	Regex       string
	Action      string
	Re          *regexp.Regexp
}

func GetDefaultRules() []Rule {
	rules := []Rule{
		{ID: 1, Description: "Block Logic", Regex: `--<(!?)([\w]+)(?:(:|!:)([^>\s]+))?`, Action: "block_start"},
		{ID: 3, Description: "Block End", Regex: `^\s*--(?:[\w!:]+>|>(?:[\w!:]*))\s*$`, Action: "block_end"},
		{ID: 8, Description: "Line Filter", Regex: `#(!?)([\w]+)(?:(:|!:)([^>\s]+))?`, Action: "line_filter"},
		{ID: 11, Description: "Replace", Regex: `\B[$:]([a-zA-Z_]\w*)`, Action: "replace"},
	}

	for i := range rules {
		rules[i].Re = regexp.MustCompile(rules[i].Regex)
	}
	return rules
}

func ProcessSQL(sqlText string, inputMap map[string]interface{}) string {
	rules := GetDefaultRules()
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(sqlText))

	inBlock := false
	blockKeep := true

	for scanner.Scan() {
		line := scanner.Text()
		lineDeleted := false

		// Block Start Check
		for _, rule := range rules {
			if rule.Action == "block_start" {
				matches := rule.Re.FindStringSubmatch(line)
				if len(matches) > 0 {
					inBlock = true
					key := matches[2]
					_, exists := inputMap[key]
					blockKeep = exists
					if matches[1] == "!" {
						blockKeep = !blockKeep
					}
					lineDeleted = true
					break
				}
			}
			if rule.Action == "block_end" && rule.Re.MatchString(line) {
				inBlock = false
				blockKeep = true
				lineDeleted = true
				break
			}
		}

		if (inBlock && !blockKeep) || lineDeleted {
			continue
		}

		// Line Filter & Replace
		for _, rule := range rules {
			if rule.Action == "line_filter" {
				matches := rule.Re.FindAllStringSubmatch(line, -1)
				for _, m := range matches {
					key := m[2]
					_, exists := inputMap[key]
					keep := exists
					if m[1] == "!" {
						keep = !keep
					}
					if !keep {
						lineDeleted = true
						break
					}
				}
			}
			if rule.Action == "replace" {
				line = rule.Re.ReplaceAllStringFunc(line, func(match string) string {
					key := match[1:]
					if val, ok := inputMap[key]; ok {
						return fmt.Sprintf("%v", val)
					}
					return match
				})
			}
		}

		if !lineDeleted {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}
