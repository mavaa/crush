package dialog

import (
	"testing"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/csync"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/crush/internal/workspace"
	"github.com/stretchr/testify/require"
)

type testWorkspace struct {
	workspace.Workspace
	cfg *config.Config
}

func (w *testWorkspace) Config() *config.Config {
	return w.cfg
}

func TestModelsSetProviderItems(t *testing.T) {
	t.Parallel()

	providers := []catwalk.Provider{
		{
			ID:   "copilot",
			Name: "GitHub Copilot",
			Models: []catwalk.Model{
				{ID: "gpt-4.1", Name: "GPT-4.1"},
			},
		},
		{
			ID:   "openai",
			Name: "OpenAI",
			Models: []catwalk.Model{
				{ID: "gpt-5", Name: "GPT-5"},
			},
		},
	}

	newModels := func(t *testing.T, showOnlyConfiguredProviders bool, isOnboarding bool) *Models {
		t.Helper()

		cfg := &config.Config{
			Options: &config.Options{
				ShowOnlyConfiguredProviders: showOnlyConfiguredProviders,
			},
			Models:       make(map[config.SelectedModelType]config.SelectedModel),
			RecentModels: make(map[config.SelectedModelType][]config.SelectedModel),
			Providers: csync.NewMapFrom(map[string]config.ProviderConfig{
				"copilot": {
					ID:   "copilot",
					Name: "GitHub Copilot",
				},
			}),
		}

		sty := styles.CharmtonePantera()
		models := &Models{
			com: &common.Common{
				Workspace: &testWorkspace{cfg: cfg},
				Styles:    &sty,
			},
			isOnboarding: isOnboarding,
			modelType:    ModelTypeLarge,
			providers:    providers,
			list:         NewModelsList(&sty),
		}
		require.NoError(t, models.setProviderItems())
		return models
	}

	t.Run("shows all providers by default", func(t *testing.T) {
		t.Parallel()

		models := newModels(t, false, false)
		require.Len(t, models.list.groups, 2)
		require.Equal(t, []string{"GitHub Copilot", "OpenAI"}, []string{models.list.groups[0].Title, models.list.groups[1].Title})
	})

	t.Run("shows only configured providers when enabled", func(t *testing.T) {
		t.Parallel()

		models := newModels(t, true, false)
		require.Len(t, models.list.groups, 1)
		require.Equal(t, "GitHub Copilot", models.list.groups[0].Title)
		require.Len(t, models.list.groups[0].Items, 1)
		require.Equal(t, "copilot:gpt-4.1", models.list.groups[0].Items[0].ID())
	})

	t.Run("keeps all providers visible during onboarding", func(t *testing.T) {
		t.Parallel()

		models := newModels(t, true, true)
		require.Len(t, models.list.groups, 2)
		require.Equal(t, []string{"GitHub Copilot", "OpenAI"}, []string{models.list.groups[0].Title, models.list.groups[1].Title})
	})
}
