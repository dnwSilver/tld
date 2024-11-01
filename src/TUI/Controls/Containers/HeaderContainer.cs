using TUI.Controls.Statics;
using TUI.Controls.Statics.Hints;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

namespace TUI.Controls.Containers;

public class HeaderContainer : ContainerBase, IContainer
{
    public override Nodes GetNodes()
    {
        var versionHints = new VersionHints();
        versionHints.SetPadding(Indentation.Default);
        versionHints.SetAlignment(Horizontal.Left);

        var tagsHints = new TagHints();
        tagsHints.SetPadding(Indentation.Default);
        tagsHints.SetAlignment(Horizontal.Left);

        var appTypeHints = new AppTypeHints();
        appTypeHints.SetPadding(Indentation.Default);
        appTypeHints.SetAlignment(Horizontal.Left);

        var hotkeysHints = new HotkeysHint();
        hotkeysHints.SetPadding(Indentation.Default);
        hotkeysHints.SetAlignment(Horizontal.Left);

        var logo = new LogoComponent();
        logo.SetAlignment(Horizontal.Right);
        logo.SetPaddingLeft(Indentation.Default);
        logo.SetPaddingBottom(Indentation.Default);
        logo.SetPaddingRight(Indentation.Default);

        return new Nodes { versionHints, tagsHints, appTypeHints, hotkeysHints, logo };
    }
}