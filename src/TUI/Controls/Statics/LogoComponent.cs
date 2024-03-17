using System.Text;
using TUI.Engine;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Controls.Statics;

public class LogoComponent : StaticComponentBase
{
    protected override void RenderWithCache(StringBuilder builder)
    {
        builder.Append($"  {"╭━━━━┳╮".Main()}{"╱╱".Hint()}{"╭━━━╮".Main()}").Append(Symbols.LineBreak);
        builder.Append($"  {"┃╭╮╭╮┃┃".Main()}{"╱╱".Hint()}{"╰╮╭╮┃".Main()}").Append(Symbols.LineBreak);
        builder.Append($"  {"╰╯┃┃╰┫┃".Main()}{"╱╱╱".Hint()}{"┃┃┃┃".Main()}").Append(Symbols.LineBreak);
        builder.Append($"  {"╱╱".Hint()}{"┃┃".Main()}{"╱".Hint()}{"┃┃".Main()}{"╱".Hint()}{"╭╮┃┃┃┃".Main()}")
            .Append(Symbols.LineBreak);
        builder.Append($" {"╱╱╱".Hint()}{"┃┃".Main()}{"╱".Hint()}{"┃╰━╯┣╯╰╯┃".Main()}").Append(Symbols.LineBreak);
        builder.Append($"{"╱╱╱╱".Hint()}{"╰╯".Main()}{"╱".Hint()}{"╰━━━┻━━━╯".Main()}").Append(Symbols.LineBreak);
    }
}