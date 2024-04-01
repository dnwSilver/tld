using TUI.Controls.Common;
using TUI.Controls.Components;
using TUI.Domain;
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
    private const int VersionColumnWidth = 10;
    private const int TitleColumnWidth = 25;

    private readonly Nodes _dependencies = new();

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

        _dependencies.Add(title);
    }

    public void AddDependency(Dependency dependency)
    {
        var version = new VersionComponent(VersionType.Convention, dependency.Version, dependency.Brand);
        version.SetPadding(Level.Normal);
        version.SetAlignment(Horizontal.Right);
        version.SetFixed(Orientation.Horizontal, VersionColumnWidth);

        _dependencies.Add(version);
    }

    public override Nodes GetNodes() => _dependencies;
}