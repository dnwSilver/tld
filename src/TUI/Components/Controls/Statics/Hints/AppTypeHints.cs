using System.Text;
using TUI.Engine;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Theme;
using TUI.UserInterface;

namespace TUI.Components.Controls.Statics.Hints;

public class AppTypeHints : ComponentStaticBase
{
    private readonly Dictionary<string, string> _hints = new()
    {
        { Icons.NpmPackage, "package" },
        { Icons.DockerImage, "image" },
        { Icons.Site, "site" },
        { Icons.Api, "api" }
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