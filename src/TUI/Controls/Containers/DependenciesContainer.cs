using TUI.Controls.Common;
using TUI.Controls.Components;
using TUI.Domain;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Components;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

namespace TUI.Controls.Containers;

public class DependenciesContainer : ContainerBase
{
    public readonly Project? Project;
    
    private const int VersionColumnWidth = 11;
    
    private const int TitleColumnWidth = 25;
    
    private readonly Nodes _dependencies = new();
    
    public DependenciesContainer()
    {
    }
    
    public DependenciesContainer(Project project)
    {
        Project = project;
    }
    
    public void AddTitleStub()
    {
        var size = new Size(TitleColumnWidth, 1);
        var title = new StubComponent(size);
        title.SetPadding(Level.Normal);
        
        _dependencies.Add(title);
    }
    
    public void AddTitle(IComponent title)
    {
        title.SetPadding(Level.Normal);
        title.SetFixed(Orientation.Horizontal, TitleColumnWidth);
        title.SetAlignment(Horizontal.Left);
        
        if (Project is not null && Project.Legacy)
        {
            title.StyleContext = new StyleContext(Palette.DisableColor);
        }
        
        _dependencies.Add(title);
    }
    
    public void AddDependencyStub()
    {
        var size = new Size(VersionColumnWidth, 1);
        var stub = new StubComponent(size, Symbols.NotFound.Hint());
        stub.SetPadding(Level.Normal);
        stub.SetAlignment(Horizontal.Right);
        stub.SetFixed(Orientation.Horizontal, VersionColumnWidth);
        
        if (Project is not null && Project.Legacy)
        {
            stub.StyleContext = new StyleContext(Palette.DisableColor);
        }
        
        _dependencies.Add(stub);
    }
    
    public void AddError()
    {
        var size = new Size(25, 1);
        var stub = new StubComponent(size, (Symbols.Error + Symbols.Space + " Something went wrong").Error());
        stub.SetPadding(Level.Normal);
        stub.SetAlignment(Horizontal.Right);
        stub.SetFixed(Orientation.Horizontal, 25);
        
        if (Project is not null && Project.Legacy)
        {
            stub.StyleContext = new StyleContext(Palette.DisableColor);
        }
        
        _dependencies.Add(stub);
    }
    
    public void AddDependency(Dependency dependency, VersionStatus status = VersionStatus.BeNice)
    {
        var version = new VersionComponent(dependency.Version, dependency.Brand, status, dependency.Type);
        version.SetPadding(Level.Normal);
        version.SetAlignment(Horizontal.Right);
        version.SetFixed(Orientation.Horizontal, VersionColumnWidth);
        
        if (Project is not null && Project.Legacy)
        {
            version.StyleContext = new StyleContext(Palette.DisableColor);
        }
        
        _dependencies.Add(version);
    }
    
    public override Nodes GetNodes() => _dependencies;
}