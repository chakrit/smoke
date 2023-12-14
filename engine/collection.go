package engine

import "strings"

type Collection []*Test

func (c Collection) Whitelist(whitelist []string) Collection {
	output := make(Collection, 0, len(c))
NextTest:
	for _, test := range c {
		for _, item := range whitelist {
			if strings.Contains(test.Name, item) {
				output = append(output, test)
				continue NextTest
			}
		}
	}
	return output
}

func (c Collection) Blacklist(blacklist []string) Collection {
	output := make(Collection, 0, len(c))
NextTest:
	for _, test := range c {
		for _, item := range blacklist {
			if strings.Contains(test.Name, item) {
				continue NextTest
			}
		}
		output = append(output, test)
	}
	return output
}
