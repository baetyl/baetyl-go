package mqtt

import (
	"strings"
)

const (
	maxTopicLevels = 9
	maxTopicLength = 255
)

// CheckTopic check topic
func CheckTopic(topic string, wildcard bool) bool {
	return checkTopic(topic, wildcard, nil)
}

func checkTopic(topic string, wildcard bool, sysTopics map[string]struct{}) bool {
	if topic == "" {
		return false
	}
	if len(topic) > maxTopicLength || strings.Contains(topic, "\u0000") {
		return false
	}
	segments := strings.Split(topic, "/")
	if strings.HasPrefix(segments[0], "$") {
		if len(segments) < 2 {
			return false
		}
		if len(sysTopics) == 0 {
			if strings.Contains(segments[0], "+") || strings.Contains(segments[0], "#") {
				return false
			}
		} else {
			if _, ok := sysTopics[segments[0]]; !ok {
				return false
			}
		}
		segments = segments[1:]
	}
	levels := len(segments)
	if levels > maxTopicLevels {
		return false
	}
	for index := 0; index < levels; index++ {
		segment := segments[index]
		// check use of wildcards
		if len(segment) > 1 && (strings.Contains(segment, "+") || strings.Contains(segment, "#")) {
			return false
		}
		// check if wildcards are allowed
		if !wildcard && (segment == "#" || segment == "+") {
			return false
		}
		// check if # is the last level
		if segment == "#" && index != levels-1 {
			return false
		}
	}
	return true
}

// TopicChecker checks topic
type TopicChecker struct {
	sysTopics map[string]struct{}
}

// NewTopicChecker create topicChecker
func NewTopicChecker(sysTopics []string) *TopicChecker {
	tc := &TopicChecker{
		sysTopics: make(map[string]struct{}),
	}
	for _, t := range sysTopics {
		tc.sysTopics[t] = struct{}{}
	}
	return tc
}

// CheckTopic checks the topic
func (tc *TopicChecker) CheckTopic(topic string, wildcard bool) bool {
	return checkTopic(topic, wildcard, tc.sysTopics)
}
