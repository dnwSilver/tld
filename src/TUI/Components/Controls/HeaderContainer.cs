using TUI.Components.Controls.Statics;
using TUI.Components.Controls.Statics.Hints;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Containers;
using TUI.Engine.Theme;

namespace TUI.Components.Controls;

public class HeaderContainer : IContainer
{
    public Orientation Orientation => Orientation.Horizontal;

    public Nodes Nodes
    {
        get
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
                .Set(Horizontal.Right)
                .Set(left: Indentation.Default, bottom: Indentation.Default, right: Indentation.Default);

            return new Nodes { versionHints, tagsHints, appTypeHints, hotkeysHints, logo };
        }
    }
}
