package cli

import "testing"

func TestStripAnsi_Colors(t *testing.T) {
	input := "\x1b[32mhello\x1b[0m world"
	got := stripAnsi(input)
	want := "hello world"
	if got != want {
		t.Errorf("stripAnsi colors: got %q, want %q", got, want)
	}
}

func TestStripAnsi_CarriageReturn(t *testing.T) {
	input := "line one\r\nline two\r\n"
	got := stripAnsi(input)
	want := "line one\nline two"
	if got != want {
		t.Errorf("stripAnsi CR: got %q, want %q", got, want)
	}
}

func TestStripAnsi_CursorMovement(t *testing.T) {
	input := "\x1b[2J\x1b[Hhello\x1b[1A"
	got := stripAnsi(input)
	want := "hello"
	if got != want {
		t.Errorf("stripAnsi cursor: got %q, want %q", got, want)
	}
}

func TestStripAnsi_CollapseBlankLines(t *testing.T) {
	input := "line one\n\n\n\n\nline two"
	got := stripAnsi(input)
	want := "line one\n\nline two"
	if got != want {
		t.Errorf("stripAnsi collapse: got %q, want %q", got, want)
	}
}

func TestStripAnsi_PlainText(t *testing.T) {
	input := "hello world"
	got := stripAnsi(input)
	if got != input {
		t.Errorf("stripAnsi plain: got %q, want %q", got, input)
	}
}

func TestStripAnsi_Empty(t *testing.T) {
	got := stripAnsi("")
	if got != "" {
		t.Errorf("stripAnsi empty: got %q, want %q", got, "")
	}
}

func TestStripAnsi_BoldAndUnderline(t *testing.T) {
	input := "\x1b[1mbold\x1b[0m \x1b[4munderline\x1b[0m"
	got := stripAnsi(input)
	want := "bold underline"
	if got != want {
		t.Errorf("stripAnsi bold/underline: got %q, want %q", got, want)
	}
}

func TestStripAnsi_256Color(t *testing.T) {
	input := "\x1b[38;5;196mred\x1b[0m"
	got := stripAnsi(input)
	want := "red"
	if got != want {
		t.Errorf("stripAnsi 256color: got %q, want %q", got, want)
	}
}

func TestWrapCmd_RequiresArgs(t *testing.T) {
	cmd := newWrapCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestWrapCmd_Help(t *testing.T) {
	cmd := newWrapCmd()
	cmd.SetArgs([]string{"--help"})
	// Should not error on --help.
	_ = cmd.Execute()
	if cmd.Use != "wrap <tool> [tool-args...]" {
		t.Errorf("unexpected Use: %q", cmd.Use)
	}
}
