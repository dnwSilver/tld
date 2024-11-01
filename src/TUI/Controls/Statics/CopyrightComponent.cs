using System.Text;
using TUI.Engine;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Controls.Statics;

public class CopyrightComponent : StaticComponentBase
{
    protected override void RenderWithCache(StringBuilder builder)
    {
        builder.Append(Symbols.Copyright.Info());
        builder.Append(Symbols.Space);
        builder.Append("Kolosov A. aka \"dnwSilver\"".Hint());
        builder.Append(Symbols.Space);
        builder.Append(DateTime.Now.Year.Info());
    }
}
