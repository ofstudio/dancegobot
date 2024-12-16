package services

import "testing"

func Test_getUsername(t *testing.T) {
	tests := []struct {
		input        string
		wantUsername string
		wantOK       bool
		gotUsername  string
		gotOK        bool
	}{
		{input: "@username", wantUsername: "username", wantOK: true},
		{input: "Hello @user_name123!", wantUsername: "user_name123", wantOK: true},
		{input: "No username here", wantUsername: "", wantOK: false},
		{input: "@tiny", wantUsername: "", wantOK: false},
		{input: "@five5", wantUsername: "five5", wantOK: true}, // Minimum length
		{input: "Hey @this_is_32_chars_long___username!", wantUsername: "this_is_32_chars_long___username", wantOK: true},
		{input: "@this_is_33_chars_long____username!", wantUsername: "", wantOK: false},
		{input: "@ew-342", wantUsername: "", wantOK: false},
		{input: "@_ew342", wantUsername: "", wantOK: false},
		{input: "@3ew342", wantUsername: "", wantOK: false},
		{input: "@", wantUsername: "", wantOK: false},
		{input: "@this_is_a_very_long_username_that_should_not_be_valid", wantUsername: "", wantOK: false},
		{input: "@validUsername123", wantUsername: "validUsername123", wantOK: true},
	}

	for _, tt := range tests {
		gotUsername, gotOK := getUsername(tt.input)
		if gotUsername != tt.wantUsername || gotOK != tt.wantOK {
			t.Errorf("username(%q) = %v, %v; want %v, %v", tt.input, gotUsername, gotOK, tt.wantUsername, tt.wantOK)
		}
	}

}
