using System.Text;
using TUI.Engine;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Theme;

namespace TUI.Components.Controls;

public class Dashboard : ComponentBase, IComponent
{
    private readonly string _title;

    public Dashboard(string title)
    {
        _title = title;
    }

    public void Render(AlignmentHorizontal alignmentHorizontal, Size size)
    {
        var dashboardBuilder = new StringBuilder();

        RenderTopLine(dashboardBuilder, size, _title);
        RenderMiddleLine(dashboardBuilder, size);
        RenderBottomLine(dashboardBuilder, size);

        // base.Render(dashboardBuilder, position, size);
    }

    private static void RenderTopLine(StringBuilder dashboardBuilder, Size size, string title)
    {
        var halfWidth = (size.Width - title.GetWidth() - (int)Indentation.BorderWidth * 2 -
                         (int)Indentation.Default * 2) / 2;
        dashboardBuilder.Append(Symbols.Angles.LeftTop);
        dashboardBuilder.Append(Symbols.Lines.Horizontal.Repeat(halfWidth));
        dashboardBuilder.AppendFormat("{0}{1}{0}", Symbols.Space.Repeat(Convert.ToInt32(Indentation.Default)), title);
        dashboardBuilder.Append(Symbols.Lines.Horizontal.Repeat(halfWidth));
        dashboardBuilder.Append(Symbols.Angles.RightTop);
    }

    private static void RenderMiddleLine(StringBuilder dashboardBuilder, Size size)
    {
        var dashboardHeight = size.Height - (int)Indentation.BorderWidth * 2;

        while (dashboardHeight > 0)
        {
            var bodyWidth = size.Width - (int)Indentation.BorderWidth * 2;
            dashboardBuilder.Append(Symbols.Lines.Vertical);
            dashboardBuilder.Append(Symbols.Space.Repeat(bodyWidth));
            dashboardBuilder.Append(Symbols.Lines.Vertical);

            dashboardHeight--;
        }
    }

    private static void RenderBottomLine(StringBuilder dashboardBuilder, Size size)
    {
        var width = size.Width - (int)Indentation.BorderWidth * 2;
        dashboardBuilder.Append(Symbols.Angles.LeftBottom);
        dashboardBuilder.Append(Symbols.Lines.Horizontal.Repeat(width));
        dashboardBuilder.Append(Symbols.Angles.RightBottom);
    }

    public override Sketch Draw()
    {
        throw new NotImplementedException();
    }
}