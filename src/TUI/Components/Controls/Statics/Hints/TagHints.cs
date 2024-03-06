using System.Text;
using TUI.Engine;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Theme;
using TUI.UserInterface;

namespace TUI.Components.Controls.Statics.Hints;

public class TagHints : ComponentStaticBase
{
    private readonly Dictionary<string, string> _hints = new()
    {
        { Icons.Auth, "Auth" },
        { Icons.NetworkPublic, "WWW" },
        { Icons.SEO, "SEO" },
        { Icons.GitLab, "VCS" }
    };

    protected override void RenderWithCache(StringBuilder builder)
    {
        foreach (var hint in _hints)
        {
            builder.Append(hint.Key);
            builder.Append(Symbols.Space);
            builder.AppendLine(hint.Value.Hint());
        }
    }
}