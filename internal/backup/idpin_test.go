package backup_test

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

// TestIDPinBackupInstance pins the instance shape BackupCopy names its
// staged copy with: a malformed instance is refused with its sentence
// before anything is staged, and the all-zeros instance is accepted.
func TestIDPinBackupInstance(t *testing.T) {
	h := openWrite(t, newStateDir(t))
	logger := slog.New(slog.DiscardHandler)

	for _, instance := range []string{
		"",
		"0123456789abcdef0123456789abcde",
		"0123456789ABCDEF0123456789ABCDEF",
		"0123456789abcdef0123456789abcdeg",
		"../../0123456789abcdef0123456789",
	} {
		t.Run(instance, func(t *testing.T) {
			staged, err := h.BackupCopy(deadline(t), logger, instance)
			want := `backup copy: instance "` + instance + `" is not 32 lowercase hex characters`
			if err == nil || err.Error() != want {
				t.Fatalf("BackupCopy(%q) = %q, %v; want the error %q", instance, staged, err, want)
			}
			if staged != "" {
				t.Errorf("BackupCopy(%q) staged %q, want nothing", instance, staged)
			}
		})
	}

	const zeros = "00000000000000000000000000000000"
	staged, err := h.BackupCopy(deadline(t), logger, zeros)
	if err != nil {
		t.Fatalf("BackupCopy(zeros): %v", err)
	}
	if !strings.HasSuffix(staged, zeros+".db") {
		t.Errorf("BackupCopy(zeros) staged %q, want a path ending %q", staged, zeros+".db")
	}
	if _, err := os.Stat(staged); err != nil {
		t.Errorf("the zeros copy is not on disk: %v", err)
	}
}
