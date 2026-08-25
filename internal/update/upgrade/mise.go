package upgrade

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v2/internal/pathidentity"
)

// currentExecutableFn and userHomeDirFn are swappable for testability,
// following this package's existing var-swap idiom (see homebrewOwnershipDetector
// in homebrew.go).
var currentExecutableFn = os.Executable
var userHomeDirFn = os.UserHomeDir

// miseInstallsRoot resolves mise's install root using mise's own precedence:
// $MISE_INSTALLS_DIR, then $MISE_DATA_DIR/installs, then
// $XDG_DATA_HOME/mise/installs, then ~/.local/share/mise/installs. The first
// SET, non-blank environment variable wins — an empty or whitespace-only
// value counts as unset, otherwise an exported-but-empty override would
// resolve the root to a cwd-relative "installs" and the containment test
// would answer a different question than the one asked.
//
// "" means unresolvable — then no binary can be mise-managed.
func miseInstallsRoot() string {
	if root := strings.TrimSpace(os.Getenv("MISE_INSTALLS_DIR")); root != "" {
		return root
	}
	if dataDir := strings.TrimSpace(os.Getenv("MISE_DATA_DIR")); dataDir != "" {
		return filepath.Join(dataDir, "installs")
	}
	if xdgDataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "mise", "installs")
	}
	home, err := userHomeDirFn()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".local", "share", "mise", "installs")
}

// runningBinaryIsMiseManaged reports whether the executable backing this
// process lives under the resolved mise installs root. It uses
// pathidentity.Contains, not a hand-rolled EvalSymlinks + prefix check:
// Contains already returns false when the root does not exist, which is the
// correct "mise not installed" answer with no extra guard.
func runningBinaryIsMiseManaged() bool {
	root := miseInstallsRoot()
	if root == "" {
		return false
	}
	exe, err := currentExecutableFn()
	if err != nil {
		return false
	}
	return pathidentity.Contains(root, exe)
}
