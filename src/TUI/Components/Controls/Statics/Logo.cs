using System.Text;
using TUI.Engine;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Theme;

namespace TUI.Components.Controls.Statics;

public class Logo : ComponentStaticBase
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