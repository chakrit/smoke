package specs

import "time"

func resolveStrings(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}

func resolveDurations(durations ...*time.Duration) *time.Duration {
	for _, dur := range durations {
		if dur != nil {
			return dur
		}
	}
	return nil
}
