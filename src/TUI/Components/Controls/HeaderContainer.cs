using TUI.Components.Controls.Statics;
using TUI.Components.Controls.Statics.Hints;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Containers;
using TUI.Engine.Theme;

namespace TUI.Components.Controls;

public class HeaderContainer : ContainerBase, IContainer
{
    public override Nodes GetNodes()
    {
        var versionHints = new VersionHints();
        versionHints.SetPadding(Indentation.Default);

        var tagsHints = new TagHints();
        tagsHints.SetPadding(Indentation.Default);

        var appTypeHints = new AppTypeHints();
        appTypeHints.SetPadding(Indentation.Default);

        var hotkeysHints = new HotkeysHint();
        hotkeysHints.SetPadding(Indentation.Default);

        var logo = new Logo();
        logo.SetAlignment(Horizontal.Right);
        logo.SetPaddingLeft(Indentation.Default);
        logo.SetPaddingBottom(Indentation.Default);
        logo.SetPaddingRight(Indentation.Default);

        return new Nodes { versionHints, tagsHints, appTypeHints, hotkeysHints, logo };
    }
}