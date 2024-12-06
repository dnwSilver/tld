using System.Diagnostics;
using TUI.Controls.Components;
using TUI.Domain;
using TUI.Engine;
using TUI.Engine.Theme;
using TUI.Logs;
using TUI.Providers.Dependencies;

namespace TUI.Store;

public class DependenciesStore
{
    public IEnumerable<Dependency> ConventionDependencies;

    public IEnumerable<Project> Projects;

    private DependencyRepository Repository = new();

    public IEnumerable<Dependency> ActualDependencies(Project project)
    {
        SpeakerComponent.Instance.Shout(Symbols.Download.Info(),
            $"Fetch actual dependencies for project {project.Name.Main()}");

        try
        {
            return Repository.ReadActual(project);
        }
        catch(Exception ex)
        {
            Log.Error("Fail load actual deps for project " +project.Name + ". " + ex.Message);
            Debugger.Log(0, "error", ex.Message);
            SpeakerComponent.Instance.Shout(Symbols.Error.Error(), $"Fetch failed for project{project.Name}");
            return new List<Dependency>();
        }
    }

    public void Bind()
    {
        try
        {
            SpeakerComponent.Instance.Shout("🤔", "Prepare javascript conventions");
            ConventionDependencies = Repository.ReadConventions("javascript");

            SpeakerComponent.Instance.Shout("🤩", "Prepare javascript projects");
            Projects = Repository.ReadProjects("javascript");
            Log.Info("Loading settings.");
        }
        catch (Exception ex)
        {
            Log.Error("Fail Loading settings. " + ex.Message);
        }
    }
}