using TUI.Controls.Components;
using TUI.Domain;
using TUI.Engine.Theme;
using TUI.Providers.Dependencies;

namespace TUI.Store;

public class DependenciesStore
{
    public IEnumerable<Dependency> ConventionDependencies;
    
    public IEnumerable<Project> Projects;
    
    private DependencyRepository Repository = new();
    
    public IEnumerable<Dependency> ActualDependencies(Project project)
    {
        SpeakerComponent.Instance.Shout("", $"Fetch actual dependencies for project {project.Name.Main()}");
        try
        {
            return Repository.ReadActual(project);
        }
        catch
        {
            SpeakerComponent.Instance.Shout("", $"Fetch failed for project{project.Name}");
            return new List<Dependency>();
        }
    }
    
    public void Bind()
    {
        SpeakerComponent.Instance.Shout("🤔", "Prepare javascript conventions");
        ConventionDependencies = Repository.ReadConventions("javascript");
        
        SpeakerComponent.Instance.Shout("🤩", "Prepare javascript projects");
        Projects = Repository.ReadProjects("javascript");
    }
}