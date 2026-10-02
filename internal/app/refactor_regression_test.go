package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/ui"
)

func TestProjectSyncFinishesAfterSelectionChanges(t *testing.T) {
	m := newModel(nil)
	m.selectedProjectID = 2
	m.projectSyncStatus = ui.ProjectSyncStatus{ProjectID: 1, Running: true}
	ch := make(chan projectSyncMsg, 1)
	m.projectSyncCh = ch
	next, _ := m.Update(projectSyncMsg{projectID: 1, done: true, message: "done"})
	updated := next.(model)
	if updated.projectSyncStatus.Running || updated.projectSyncCh != nil {
		t.Fatalf("sync remained active after selection changed: %#v", updated.projectSyncStatus)
	}
}

func TestStaleProjectDependencyResponseIsIgnored(t *testing.T) {
	m := newModel(nil)
	m.selectedProjectID = 1
	old := m.loadProjectDependencies()
	m.selectedProjectID = 2
	newer := m.loadProjectDependencies()
	newMsg := newer().(projectDependenciesLoadedMsg)
	oldMsg := old().(projectDependenciesLoadedMsg)
	next, _ := m.Update(newMsg)
	next, _ = next.(model).Update(oldMsg)
	if got := next.(model).selectedProjectID; got != 2 {
		t.Fatalf("selection changed to %d", got)
	}
	if next.(model).err != nil {
		t.Fatalf("stale response changed error: %v", next.(model).err)
	}
}

func TestStaleProjectsListDoesNotReplaceNewerSnapshot(t *testing.T) {
	m := newModel(nil)
	old := m.loadProjects()
	newer := m.loadProjects()
	newMsg := newer().(projectsLoadedMsg)
	newMsg.projects = []ui.Project{{ID: 2, Name: "new"}}
	oldMsg := old().(projectsLoadedMsg)
	oldMsg.projects = []ui.Project{{ID: 1, Name: "old"}}
	next, _ := m.Update(newMsg)
	next, _ = next.(model).Update(oldMsg)
	updated := next.(model)
	if len(updated.projects) != 1 || updated.projects[0].ID != 2 {
		t.Fatalf("stale projects replaced newer snapshot: %#v", updated.projects)
	}
}

func TestRetryProjectsReloadsLocalList(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenProjects
	m.lastError = "temporary read failure"
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil || next.(model).lastError != "" {
		t.Fatal("retry did not start a fresh local load")
	}
	msg := cmd()
	if _, ok := msg.(projectsLoadedMsg); !ok {
		t.Fatalf("retry returned %T, want projectsLoadedMsg", msg)
	}
}

func TestProjectNameAcceptsSpace(t *testing.T) {
	form := ui.ProjectForm{Open: true, Focus: ui.ProjectFormFieldName, Name: "My"}
	form, _ = updateProjectFormState(key(" "), form, nil, nil, nil)
	if form.Name != "My " {
		t.Fatalf("project name = %q", form.Name)
	}
}

func TestDoubleEnterStartsOneSave(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenProjects
	m.projectForm = ui.ProjectForm{Open: true, CanSave: true, Name: "Project", ProjectID: "42"}
	next, first := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if first == nil || !next.(model).savePending {
		t.Fatal("first Enter did not start save")
	}
	next, second := next.(model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if second != nil {
		t.Fatal("second Enter started another save")
	}
}

func TestLateRegistryVersionDoesNotOverwriteNewForm(t *testing.T) {
	m := newModel(nil)
	m.policyValueForm = ui.PolicyValueForm{Open: true, PolicyID: 1, DependencyID: 10, RegistryID: 5, Version: "old"}
	old := m.loadPolicyValueLatest()
	m.policyValueForm = ui.PolicyValueForm{Open: true, PolicyID: 2, DependencyID: 20, RegistryID: 6, Version: "current"}
	_ = m.loadPolicyValueLatest()
	next, _ := m.Update(old())
	if next.(model).policyValueForm.Version != "current" || next.(model).policyValueForm.Error != "" {
		t.Fatalf("stale latest response changed form: %#v", next.(model).policyValueForm)
	}
}

func TestCanceledJobSendDoesNotBlockWithoutReceiver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		sendJobMsg(ctx, make(chan projectSyncMsg), projectSyncMsg{message: "progress"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("canceled job send blocked without receiver")
	}
}

func TestCancelProjectSyncIgnoresLateMessages(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenProjects
	ctx, cancel := context.WithCancel(context.Background())
	m.projectSyncCancel = cancel
	m.projectSyncRunID = 7
	m.projectSyncStatus = ui.ProjectSyncStatus{Running: true, Message: "working"}
	m.projectSyncCh = make(chan projectSyncMsg, 1)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := next.(model)
	if ctx.Err() != context.Canceled || updated.projectSyncStatus.Running || updated.projectSyncCh != nil || updated.projectSyncRunID != 8 {
		t.Fatalf("project sync not canceled: %#v", updated.projectSyncStatus)
	}
	next, _ = updated.Update(projectSyncMsg{runID: 7, done: true, message: "old result"})
	if next.(model).projectSyncStatus.Message != "Sync canceled" {
		t.Fatalf("late result replaced cancel status: %#v", next.(model).projectSyncStatus)
	}
}

func TestCancelOtherRefreshesIgnoresLateMessages(t *testing.T) {
	for _, screen := range []ui.Screen{ui.ScreenSettings, ui.ScreenReleases, ui.ScreenVulnerabilities} {
		t.Run(fmt.Sprint(screen), func(t *testing.T) {
			m := newModel(nil)
			m.screen = screen
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch screen {
			case ui.ScreenSettings:
				m.checksCancel, m.checksRunID = cancel, 4
				m.checksStatus = ui.SettingsStatus{Running: true}
				m.checksSyncCh = make(chan checkSyncMsg, 1)
			case ui.ScreenReleases:
				m.releasesCancel, m.releasesRunID = cancel, 4
				m.releasesStatus = ui.SettingsStatus{Running: true}
				m.releasesSyncCh = make(chan releaseSyncMsg, 1)
			case ui.ScreenVulnerabilities:
				m.vulnsCancel, m.vulnsRunID = cancel, 4
				m.vulnsStatus = ui.SettingsStatus{Running: true}
				m.vulnsSyncCh = make(chan vulnSyncMsg, 1)
			}
			next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			updated := next.(model)
			if ctx.Err() != context.Canceled || updated.screenJobRunning() {
				t.Fatal("refresh was not canceled")
			}
			switch screen {
			case ui.ScreenSettings:
				next, _ = updated.Update(checkSyncMsg{runID: 4, done: true, message: "old"})
				if next.(model).checksStatus.Message != "Checks canceled" {
					t.Fatal("stale checks result was applied")
				}
			case ui.ScreenReleases:
				next, _ = updated.Update(releaseSyncMsg{runID: 4, done: true, message: "old"})
				if next.(model).releasesStatus.Message != "Releases canceled" {
					t.Fatal("stale releases result was applied")
				}
			case ui.ScreenVulnerabilities:
				next, _ = updated.Update(vulnSyncMsg{runID: 4, done: true, message: "old"})
				if next.(model).vulnsStatus.Message != "Scan canceled" {
					t.Fatal("stale scan result was applied")
				}
			}
		})
	}
}
