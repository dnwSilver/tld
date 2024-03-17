using System.Text;
using TUI.Engine;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Controls.Statics.Hints;

public class AppTypeHints : StaticComponentBase
{
    private readonly Dictionary<string, string> _hints = new()
    {
        { Symbols.NpmPackage, "package" },
        { Symbols.DockerImage, "image" },
        { Symbols.Site, "site" },
        { Symbols.Api, "api" }
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