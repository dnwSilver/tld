using System.Text;
using TUI.Engine;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Theme;

namespace TUI.Components.Controls;

public class Copyright : ComponentStaticBase
{
    protected override void RenderWithCache(StringBuilder builder)
    {
        builder.Append(Symbols.Copyright);
        builder.Append(Symbols.Space);
        builder.Append("Kolosov A. aka \"dnwSilver\"".Hint());
        builder.Append(Symbols.Space);
        builder.Append(DateTime.Now.Year);
    }
}
