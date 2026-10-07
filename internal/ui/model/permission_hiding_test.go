package model

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/ui/attachments"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type permissionHidingDialog struct {
	messages int
}

func (*permissionHidingDialog) ID() string { return dialog.PermissionsID }

func (d *permissionHidingDialog) HandleMsg(tea.Msg) dialog.Action {
	d.messages++
	return nil
}

func (*permissionHidingDialog) Draw(uv.Screen, uv.Rectangle) *tea.Cursor { return nil }

func TestPermissionHidingTogglePreservesPendingDialog(t *testing.T) {
	t.Parallel()

	u, _ := newPlanUI(t, "sess-1")
	u.keyMap = DefaultKeyMap()
	u.state = uiChat
	u.focus = uiFocusEditor
	u.textarea.Focus()
	u.textarea.Placeholder = "Original prompt"
	pending := &permissionHidingDialog{}
	u.dialog.OpenDialog(pending)

	ctrlH := tea.KeyPressMsg{Code: 'h', Mod: tea.ModCtrl}
	u.handleKeyPressMsg(ctrlH)
	require.True(t, u.isPermissionDialogFrontHidden())
	require.Equal(t, uiFocusMain, u.focus)
	require.False(t, u.textarea.Focused())
	require.True(t, u.dialog.ContainsDialog(dialog.PermissionsID))
	require.False(t, u.dialog.HasDialogsExcept(dialog.PermissionsID))

	for _, msg := range []tea.KeyPressMsg{{Code: tea.KeyTab}, {Code: tea.KeyEscape}} {
		require.Nil(t, u.handleKeyPressMsg(msg))
		require.Equal(t, uiFocusMain, u.focus)
	}

	u.textarea.Placeholder = "Prompt input disabled while permissions pending"
	u.handleKeyPressMsg(ctrlH)
	require.False(t, u.permissionDialogHidden)
	require.Equal(t, uiFocusEditor, u.focus)
	require.True(t, u.textarea.Focused())
	require.Equal(t, "Original prompt", u.textarea.Placeholder)
	require.True(t, u.dialog.ContainsDialog(dialog.PermissionsID))
	require.Zero(t, pending.messages, "hiding must not respond to the permission request")
}

func TestPermissionHidingHelpPreservesModeBinding(t *testing.T) {
	t.Parallel()

	for _, hidden := range []bool{false, true} {
		u, _ := newPlanUI(t, "sess-1")
		u.keyMap = DefaultKeyMap()
		u.state = uiChat
		u.focus = uiFocusMain
		u.attachments = attachments.New(nil, attachments.Keymap{})
		u.dialog.OpenDialog(&permissionHidingDialog{})
		u.permissionDialogHidden = hidden
		u.agentBusyCache.val = true

		groups := [][]key.Binding{u.ShortHelp()}
		groups = append(groups, u.FullHelp()...)
		for _, bindings := range [][]key.Binding{groups[0], flattenPermissionHelp(groups[1:])} {
			counts := map[string]int{}
			cancelHelp := false
			for _, binding := range bindings {
				cancelHelp = cancelHelp || binding.Help().Desc == "cancel"
				for _, name := range binding.Keys() {
					counts[name]++
				}
			}
			require.Equal(t, 1, counts["shift+tab"])
			require.Equal(t, !hidden, cancelHelp)
			if hidden {
				require.Zero(t, counts["tab"])
			} else {
				require.Equal(t, 1, counts["tab"])
			}
		}
	}
}

func flattenPermissionHelp(groups [][]key.Binding) []key.Binding {
	var bindings []key.Binding
	for _, group := range groups {
		bindings = append(bindings, group...)
	}
	return bindings
}
