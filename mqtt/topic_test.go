package mqtt

import (
	"math/rand"
	"testing"
)

func TestCheckTopic(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		wildcard bool
		want1    bool
		want2    bool
	}{
		// disable wildcard
		{name: "11", topic: "topic", wildcard: false, want1: true, want2: true},
		{name: "12", topic: "topic/a", wildcard: false, want1: true, want2: true},
		{name: "13", topic: "topic/a/b", wildcard: false, want1: true, want2: true},
		{name: "14", topic: "$baidu/services/a", wildcard: false, want1: true, want2: true},
		{name: "15", topic: "$link/services/a", wildcard: false, want1: true, want2: true},
		{name: "16", topic: "a/b/c/d/e/f/g/h/i", wildcard: false, want1: true, want2: true},
		{name: "17", topic: "$baidu/a/b/c/d/e/f/g/h/i", wildcard: true, want1: true, want2: true},
		{name: "18", topic: genRandomString(255), wildcard: false, want1: true, want2: true},

		{name: "21", topic: "$SYS/services/a", wildcard: false, want1: true, want2: false},

		{name: "31", topic: "$baidu", wildcard: false, want1: false, want2: false},
		{name: "32", topic: "$link", wildcard: false, want1: false, want2: false},
		{name: "33", topic: "$SYS", wildcard: false, want1: false, want2: false},
		{name: "35", topic: "", wildcard: false, want1: false, want2: false},
		{name: "36", topic: "+", wildcard: false, want1: false, want2: false},
		{name: "37", topic: "#", wildcard: false, want1: false, want2: false},
		{name: "38", topic: "topic/+", wildcard: false, want1: false, want2: false},
		{name: "39", topic: "topic/#", wildcard: false, want1: false, want2: false},
		{name: "40", topic: "$SYS+/a", wildcard: false, want1: false, want2: false},
		{name: "41", topic: "$SYS/+", wildcard: false, want1: false, want2: false},
		{name: "42", topic: "$SYS#/a", wildcard: false, want1: false, want2: false},
		{name: "43", topic: "$SYS/#", wildcard: false, want1: false, want2: false},
		{name: "44", topic: "$baidu+/a", wildcard: false, want1: false, want2: false},
		{name: "45", topic: "$baidu/+", wildcard: false, want1: false, want2: false},
		{name: "46", topic: "$baidu/#", wildcard: false, want1: false, want2: false},
		{name: "47", topic: "$link+/a", wildcard: false, want1: false, want2: false},
		{name: "48", topic: "$link/+", wildcard: false, want1: false, want2: false},
		{name: "49", topic: "$link/#", wildcard: false, want1: false, want2: false},
		{name: "50", topic: "a/b/c/d/e/f/g/h/i/", wildcard: false, want1: false, want2: false},
		{name: "52", topic: "a/b/c/d/e/f/g/h/i/j", wildcard: false, want1: false, want2: false},
		{name: "53", topic: "$baidu/a/b/c/d/e/f/g/h/i/j", wildcard: false, want1: false, want2: false},
		{name: "54", topic: genRandomString(256), wildcard: false, want1: false, want2: false},

		// enable wildcard
		{name: "111", topic: "topic", wildcard: true, want1: true, want2: true},
		{name: "112", topic: "topic/a", wildcard: true, want1: true, want2: true},
		{name: "113", topic: "topic/a/b", wildcard: true, want1: true, want2: true},
		{name: "114", topic: "+", wildcard: true, want1: true, want2: true},
		{name: "115", topic: "#", wildcard: true, want1: true, want2: true},
		{name: "116", topic: "topic/+", wildcard: true, want1: true, want2: true},
		{name: "117", topic: "topic/#", wildcard: true, want1: true, want2: true},
		{name: "118", topic: "topic/+/b", wildcard: true, want1: true, want2: true},
		{name: "119", topic: "topic/a/+", wildcard: true, want1: true, want2: true},
		{name: "120", topic: "topic/a/#", wildcard: true, want1: true, want2: true},
		{name: "121", topic: "+/a/#", wildcard: true, want1: true, want2: true},
		{name: "122", topic: "+/+/#", wildcard: true, want1: true, want2: true},
		{name: "123", topic: "$baidu/+/a", wildcard: true, want1: true, want2: true},
		{name: "124", topic: "$baidu/+/#", wildcard: true, want1: true, want2: true},
		{name: "125", topic: "$baidu/services/a", wildcard: true, want1: true, want2: true},
		{name: "126", topic: "$link/+/a", wildcard: true, want1: true, want2: true},
		{name: "127", topic: "$link/+/#", wildcard: true, want1: true, want2: true},
		{name: "128", topic: "$link/services/a", wildcard: true, want1: true, want2: true},
		{name: "129", topic: "a/b/c/d/e/f/g/h/i", wildcard: true, want1: true, want2: true},
		{name: "130", topic: "$baidu/a/b/c/d/e/f/g/h/i", wildcard: true, want1: true, want2: true},

		{name: "141", topic: "$SYS/+", wildcard: true, want1: true, want2: false},
		{name: "142", topic: "$SYS/#", wildcard: true, want1: true, want2: false},
		{name: "143", topic: "$SYS/services/a", wildcard: false, want1: true, want2: false},

		{name: "151", topic: "$baidu", wildcard: true, want1: false, want2: false},
		{name: "152", topic: "$link", wildcard: true, want1: false, want2: false},
		{name: "153", topic: "$SYS", wildcard: true, want1: false, want2: false},
		{name: "154", topic: "", wildcard: true, want1: false, want2: false},
		{name: "155", topic: "++", wildcard: true, want1: false, want2: false},
		{name: "156", topic: "##", wildcard: true, want1: false, want2: false},
		{name: "157", topic: "#/+", wildcard: true, want1: false, want2: false},
		{name: "158", topic: "#/#", wildcard: true, want1: false, want2: false},
		{name: "160", topic: "$+", wildcard: true, want1: false, want2: false},
		{name: "161", topic: "$#", wildcard: true, want1: false, want2: false},
		{name: "162", topic: "$+/a", wildcard: true, want1: false, want2: false},
		{name: "163", topic: "$#/a", wildcard: true, want1: false, want2: false},
		{name: "164", topic: "a/b/c/d/e/f/g/h/i/", wildcard: true, want1: false, want2: false},
		{name: "165", topic: "a/b/c/d/e/f/g/h/i/j", wildcard: true, want1: false, want2: false},
		{name: "166", topic: "$baidu/a/b/c/d/e/f/g/h/i/j", wildcard: true, want1: false, want2: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTopicChecker(nil)
			if got := tc.CheckTopic(tt.topic, tt.wildcard); got != tt.want1 {
				t.Errorf("topic = %s CheckTopic() = %v, want1 %v", tt.topic, got, tt.want1)
			}
			tc = NewTopicChecker([]string{"$baidu", "$link"})
			if got := tc.CheckTopic(tt.topic, tt.wildcard); got != tt.want2 {
				t.Errorf("topic = %s CheckTopic() = %v, want2 %v", tt.topic, got, tt.want2)
			}
		})
	}
}

func genRandomString(n int) string {
	c := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_")
	b := make([]byte, n)
	for i := range b {
		b[i] = c[rand.Intn(len(c))]
	}
	return string(b)
}
