using System.Text;
using TUI.Engine;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Controls.Statics.Hints;

public class TagHints : StaticComponentBase
{
    private readonly Dictionary<string, string> _hints = new()
    {
        { Symbols.Auth, "Auth" },
        { Symbols.NetworkPublic, "WWW" },
        { Symbols.SEO, "SEO" },
        { Symbols.Git, "VCS" }
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