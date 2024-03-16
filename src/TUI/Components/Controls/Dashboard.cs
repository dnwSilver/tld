using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Components;
using TUI.Engine.Theme;

namespace TUI.Components.Controls;

public class Dashboard : ComponentBase, IComponent
{
    private readonly string _title;

    public Dashboard(string title)
    {
        _title = title;
    }

    private static void RenderTopLine(StringBuilder builder, Size size, string title)
    {
        var halfWidth = (size.Width - title.GetWidth() - (int)Indentation.BorderWidth * 2 -
                         (int)Indentation.Default * 2) / 2;
        builder.Append(Symbols.Angles.LeftTop);
        builder.Append(Symbols.Lines.Horizontal.Repeat(halfWidth));
        builder.AppendFormat("{0}{1}{0}", Symbols.Space.Repeat(Convert.ToInt32(Indentation.Default)), title);
        builder.Append(Symbols.Lines.Horizontal.Repeat(halfWidth));
        builder.Append(Symbols.Angles.RightTop);
        builder.Append(Symbols.LineBreak);
    }

    private static void RenderMiddleLine(StringBuilder builder, Size size)
    {
        var dashboardHeight = size.Height - (int)Indentation.BorderWidth * 2;

        while (dashboardHeight > 0)
        {
            var bodyWidth = size.Width - (int)Indentation.BorderWidth * 2;
            builder.Append(Symbols.Lines.Vertical);
            builder.Append(Symbols.Space.Repeat(bodyWidth));
            builder.Append(Symbols.Lines.Vertical);
            builder.Append(Symbols.LineBreak);

            dashboardHeight--;
        }
    }

    private static void RenderBottomLine(StringBuilder builder, Size size)
    {
        var width = size.Width - (int)Indentation.BorderWidth * 2;
        builder.Append(Symbols.Angles.LeftBottom);
        builder.Append(Symbols.Lines.Horizontal.Repeat(width));
        builder.Append(Symbols.Angles.RightBottom);
        builder.Append(Symbols.LineBreak);
    }

    protected override Sketch DrawComponent()
    {
        var builder = new StringBuilder();

        var size = new Size(40, 5);

        RenderTopLine(builder, size, _title);
        RenderMiddleLine(builder, size);
        RenderBottomLine(builder, size);

        return new Sketch(builder.ToString());
    }
}