using System.Text;
using TUI.Engine;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Components.Controls.Statics.Hints;

public class VersionHints : StaticComponentBase
{
    private readonly Dictionary<string, string> _hints = new()
    {
        { "󰎔", "too new".Info() },
        { "", "so good".Hint() },
        { "", "be nice".Main() },
        { "󰬟", "too old".Warning() }
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