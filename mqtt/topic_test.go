package mqtt

import (
	"testing"
)

func TestCheckTopic(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		wildcard bool
		want     bool
	}{
		{name: "1", topic: "topic", wildcard: false, want: true},
		{name: "2", topic: "topic/a", wildcard: false, want: true},
		{name: "3", topic: "topic/a/b", wildcard: false, want: true},
		{name: "4", topic: "$baidu/services/a", wildcard: false, want: true},
		{name: "5", topic: "$link/services/a", wildcard: false, want: true},
		{name: "6", topic: "a/b/c/d/e/f/g/h/i", wildcard: true, want: true},
		{name: "7", topic: "$baidu/a/b/c/d/e/f/g/h/i", wildcard: true, want: true},

		{name: "8", topic: "+", wildcard: false, want: false},
		{name: "9", topic: "#", wildcard: false, want: false},
		{name: "10", topic: "topic/+", wildcard: false, want: false},
		{name: "11", topic: "topic/#", wildcard: false, want: false},
		{name: "12", topic: "$SYS", wildcard: false, want: false},
		{name: "13", topic: "$SYS/services/a", wildcard: false, want: false},
		{name: "14", topic: "$SYS+/a", wildcard: false, want: false},
		{name: "15", topic: "$SYS/+", wildcard: false, want: false},
		{name: "16", topic: "$SYS#/a", wildcard: false, want: false},
		{name: "17", topic: "$SYS/#", wildcard: false, want: false},
		{name: "18", topic: "$baidu", wildcard: false, want: false},
		{name: "19", topic: "$baidu+/a", wildcard: false, want: false},
		{name: "20", topic: "$baidu/+", wildcard: false, want: false},
		{name: "21", topic: "$baidu/#", wildcard: false, want: false},
		{name: "22", topic: "$link", wildcard: false, want: false},
		{name: "23", topic: "$link+/a", wildcard: false, want: false},
		{name: "24", topic: "$link/+", wildcard: false, want: false},
		{name: "25", topic: "$link/#", wildcard: false, want: false},
		{name: "26", topic: "a/b/c/d/e/f/g/h/i/", wildcard: false, want: false},
		{name: "27", topic: "a/b/c/d/e/f/g/h/i/j", wildcard: false, want: false},
		{name: "28", topic: "$baidu/a/b/c/d/e/f/g/h/i/j", wildcard: false, want: false},

		{name: "29", topic: "topic", wildcard: true, want: true},
		{name: "30", topic: "topic/a", wildcard: true, want: true},
		{name: "31", topic: "topic/a/b", wildcard: true, want: true},
		{name: "32", topic: "+", wildcard: true, want: true},
		{name: "33", topic: "#", wildcard: true, want: true},
		{name: "34", topic: "topic/+", wildcard: true, want: true},
		{name: "35", topic: "topic/#", wildcard: true, want: true},
		{name: "36", topic: "topic/+/b", wildcard: true, want: true},
		{name: "37", topic: "topic/a/+", wildcard: true, want: true},
		{name: "38", topic: "topic/a/#", wildcard: true, want: true},
		{name: "39", topic: "+/a/#", wildcard: true, want: true},
		{name: "40", topic: "+/+/#", wildcard: true, want: true},
		{name: "41", topic: "$baidu/+/a", wildcard: true, want: true},
		{name: "42", topic: "$baidu/+/#", wildcard: true, want: true},
		{name: "43", topic: "$baidu/services/a", wildcard: true, want: true},
		{name: "44", topic: "$link/+/a", wildcard: true, want: true},
		{name: "45", topic: "$link/+/#", wildcard: true, want: true},
		{name: "46", topic: "$link/services/a", wildcard: true, want: true},
		{name: "47", topic: "a/b/c/d/e/f/g/h/i", wildcard: true, want: true},
		{name: "48", topic: "$baidu/a/b/c/d/e/f/g/h/i", wildcard: true, want: true},

		{name: "49", topic: "", wildcard: true, want: false},
		{name: "50", topic: "++", wildcard: true, want: false},
		{name: "51", topic: "##", wildcard: true, want: false},
		{name: "52", topic: "#/+", wildcard: true, want: false},
		{name: "53", topic: "#/#", wildcard: true, want: false},
		{name: "54", topic: "$baidu", wildcard: false, want: false},
		{name: "55", topic: "$link", wildcard: false, want: false},
		{name: "56", topic: "$SYS", wildcard: false, want: false},
		{name: "57", topic: "$SYS/+", wildcard: false, want: false},
		{name: "58", topic: "$SYS/#", wildcard: false, want: false},
		{name: "59", topic: "$SYS/services/a", wildcard: false, want: false},
		{name: "60", topic: "$+", wildcard: true, want: false},
		{name: "61", topic: "$#", wildcard: true, want: false},
		{name: "62", topic: "$+/a", wildcard: true, want: false},
		{name: "63", topic: "$#/a", wildcard: true, want: false},
		{name: "64", topic: "a/b/c/d/e/f/g/h/i/", wildcard: false, want: false},
		{name: "65", topic: "a/b/c/d/e/f/g/h/i/j", wildcard: false, want: false},
		{name: "66", topic: "$baidu/a/b/c/d/e/f/g/h/i/j", wildcard: false, want: false},
	}

	tc := NewTopicChecker([]string{"$baidu", "$link"})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tc.CheckTopic(tt.topic, tt.wildcard); got != tt.want {
				t.Errorf("topic = %s CheckTopic() = %v, want %v", tt.topic, got, tt.want)
			}
		})
	}
}
