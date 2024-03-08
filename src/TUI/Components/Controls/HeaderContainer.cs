using TUI.Components.Controls.Statics;
using TUI.Components.Controls.Statics.Hints;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizings;
using TUI.Engine.Nodes.Containers;
using TUI.Engine.Theme;

namespace TUI.Components.Controls;

public abstract class ContainerBase : IContainer
{
    public Orientation Orientation => Orientation.Horizontal;

    public Resizing ResizingHorizontal => Resizing.Adaptive;

    public Resizing ResizingVertical => Resizing.Hug;

    public Size Fixed { get; }

    public abstract Nodes GetNodes();
}

public class HeaderContainer : ContainerBase, IContainer
{
    public override Nodes GetNodes()
    {
        var versionHints = new VersionHints()
            .Set(Indentation.Default);

        var tagsHints = new TagHints()
            .Set(Indentation.Default);

        var appTypeHints = new AppTypeHints()
            .Set(Indentation.Default);

        var hotkeysHints = new HotkeysHint()
            .Set(Indentation.Default);

        var logo = new Logo()
            .Set(AlignmentHorizontal.Right)
            .Set(left: Indentation.Default, bottom: Indentation.Default, right: Indentation.Default);

        return new Nodes { versionHints, tagsHints, appTypeHints, hotkeysHints, logo };
    }
}