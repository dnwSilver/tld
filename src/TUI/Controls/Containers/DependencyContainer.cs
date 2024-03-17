using TUI.Controls.Common;
using TUI.Controls.Components;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

namespace TUI.Controls.Containers;

public class DependencyContainer : ContainerBase
{
    private const int VersionColumnWidth = 10;

    public override Nodes GetNodes()
    {
        var stub = new StubComponent(new Size(20, 1));
        stub.SetPadding(Level.Normal);

        var version = new VersionComponent(VersionType.Convention, "10.20.30", Symbols.Site);
        version.SetPadding(Level.Normal);
        version.SetAlignment(Horizontal.Right);
        version.SetFixed(Orientation.Horizontal, VersionColumnWidth);

        return new Nodes { stub, version, version, };
    }
}