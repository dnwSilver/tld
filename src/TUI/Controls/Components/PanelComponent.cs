using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Components;
using TUI.Engine.Theme;
using static TUI.Engine.Symbols;

namespace TUI.Controls.Components;

public class PanelComponent : ComponentBase, IComponent
{
    private readonly string _title;

    public PanelComponent(string title)
    {
        _title = title;
        SetAbsolute();
    }

    private static void RenderTopLine(StringBuilder builder, Size size, string title)
    {
        var halfWidth = (size.Width - title.GetWidth() - (int)Indentation.BorderWidth * 2 -
                         (int)Indentation.Default * 2) / 2;
        builder.Append(Angles.LeftTop);
        builder.Append(Lines.Horizontal.Repeat(halfWidth));
        builder.AppendFormat("{0}{1}{0}", Space.Repeat(Convert.ToInt32(Indentation.Default)), title);
        builder.Append(Lines.Horizontal.Repeat(halfWidth));
        builder.Append(Angles.RightTop);
        builder.Append(LineBreak);
    }

    private static void RenderMiddleLine(StringBuilder builder, Size size)
    {
        var dashboardHeight = size.Height - (int)Indentation.BorderWidth * 2;

        while (dashboardHeight > 0)
        {
            var bodyWidth = size.Width - (int)Indentation.BorderWidth * 2;
            builder.Append(Lines.Vertical);
            builder.Append(Space.Repeat(bodyWidth));
            builder.Append(Lines.Vertical);
            builder.Append(LineBreak);

            dashboardHeight--;
        }
    }

    private static void RenderBottomLine(StringBuilder builder, Size size)
    {
        var width = size.Width - (int)Indentation.BorderWidth * 2;
        builder.Append(Angles.LeftBottom);
        builder.Append(Lines.Horizontal.Repeat(width));
        builder.Append(Angles.RightBottom);
        builder.Append(LineBreak);
    }

    protected override Sketch DrawComponent(Size minSize)
    {
        var builder = new StringBuilder();

        RenderTopLine(builder, minSize, _title);
        RenderMiddleLine(builder, minSize);
        RenderBottomLine(builder, minSize);

        return new Sketch(builder.ToString().Main());
    }
}