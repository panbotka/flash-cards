package importer

import (
	"testing"
)

func TestDetectDelimiter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "tab-separated",
			content: "hello\tworld\nfoo\tbar\n",
			want:    "\t",
		},
		{
			name:    "semicolon-separated",
			content: "hello;world\nfoo;bar\n",
			want:    ";",
		},
		{
			name:    "dash-separated",
			content: "hello - world\nfoo - bar\n",
			want:    " - ",
		},
		{
			name:    "equals-separated",
			content: "hello = world\nfoo = bar\n",
			want:    " = ",
		},
		{
			name:    "comma-separated",
			content: "hello,world\nfoo,bar\n",
			want:    ",",
		},
		{
			name:    "defaults to tab when no delimiter matches",
			content: "single word\nanother word\n",
			want:    "\t",
		},
		{
			name:    "empty content defaults to tab",
			content: "",
			want:    "\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectDelimiter(tt.content)
			if got != tt.want {
				t.Errorf("DetectDelimiter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantFirst [2]string // [czech, english]
	}{
		{
			name:      "tab-separated cards",
			content:   "ahoj\thello\nsvět\tworld\n",
			wantCount: 2,
			wantFirst: [2]string{"ahoj", "hello"},
		},
		{
			name:      "semicolon-separated cards",
			content:   "ahoj;hello\nsvět;world\n",
			wantCount: 2,
			wantFirst: [2]string{"ahoj", "hello"},
		},
		{
			name:      "dash-separated cards",
			content:   "ahoj - hello\nsvět - world\n",
			wantCount: 2,
			wantFirst: [2]string{"ahoj", "hello"},
		},
		{
			name:      "equals-separated cards",
			content:   "ahoj = hello\nsvět = world\n",
			wantCount: 2,
			wantFirst: [2]string{"ahoj", "hello"},
		},
		{
			name:      "comma-separated cards",
			content:   "ahoj,hello\nsvět,world\n",
			wantCount: 2,
			wantFirst: [2]string{"ahoj", "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards, err := Parse(tt.content)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if len(cards) != tt.wantCount {
				t.Fatalf("Parse() returned %d cards, want %d", len(cards), tt.wantCount)
			}
			if cards[0].Czech != tt.wantFirst[0] || cards[0].English != tt.wantFirst[1] {
				t.Errorf("first card = {%q, %q}, want {%q, %q}",
					cards[0].Czech, cards[0].English, tt.wantFirst[0], tt.wantFirst[1])
			}
		})
	}
}

func TestParse_SkipsEmptyLines(t *testing.T) {
	content := "ahoj\thello\n\n\nsvět\tworld\n\n"
	cards, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("Parse() returned %d cards, want 2", len(cards))
	}
}

func TestParse_SkipsInvalidLines(t *testing.T) {
	content := "ahoj\thello\nno delimiter here\nthree\tparts\textra\nsvět\tworld\n"
	cards, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("Parse() returned %d cards, want 2", len(cards))
	}
}

func TestParse_TrimsWhitespace(t *testing.T) {
	content := "  ahoj  \t  hello  \n"
	cards, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("Parse() returned %d cards, want 1", len(cards))
	}
	if cards[0].Czech != "ahoj" {
		t.Errorf("Czech = %q, want %q", cards[0].Czech, "ahoj")
	}
	if cards[0].English != "hello" {
		t.Errorf("English = %q, want %q", cards[0].English, "hello")
	}
}

func TestParse_SkipsEmptyFields(t *testing.T) {
	content := "\t hello\nahoj\t\n\t\nahoj\thello\n"
	cards, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("Parse() returned %d cards, want 1 (only the valid one)", len(cards))
	}
}
