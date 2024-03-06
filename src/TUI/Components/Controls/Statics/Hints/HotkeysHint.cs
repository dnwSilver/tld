using System.Text;
using TUI.Engine;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Theme;

namespace TUI.Components.Controls.Statics.Hints;

public class HotkeysHint : ComponentStaticBase
{
    private readonly Dictionary<string, string> _hints = new()
    {
        { "", "select prev" },
        { "", "select next" },
        { "󰬌", "toggle head" },
        { "󰬘", "quit" }
    };

    protected override void RenderWithCache(StringBuilder builder)
    {
        foreach (var hint in _hints)
        {
            builder.Append(hint.Key.Hint());
            builder.Append(Symbols.Space);
            builder.AppendLine(hint.Value.Hint());
        }
    }
}