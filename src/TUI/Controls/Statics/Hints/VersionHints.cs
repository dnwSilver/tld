using System.Text;
using TUI.Controls.Components;
using TUI.Engine;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Controls.Statics.Hints;

public class VersionHints : StaticComponentBase
{
    private readonly Dictionary<string, string> _hints = new()
    {
        { "󰎔", VersionType.ToNew.Colorize("too new") },
        { "", VersionType.SoGood.Colorize("so good") },
        { "", VersionType.BeNice.Colorize("be nice") },
        { "󰬟", VersionType.TooOld.Colorize("too old") }
    };

    protected override void RenderWithCache(StringBuilder builder)
    {
        foreach (var hint in _hints)
        {
            builder.Append(hint.Key.Hint());
            builder.Append(Symbols.Space);
            builder.AppendLine(hint.Value);
        }
    }
}