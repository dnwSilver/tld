using System.Diagnostics;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Engine.Rendering;

public class ConsoleRenderingEngine : IRenderingEngine
{
    private readonly IWindow _window;

    public ConsoleRenderingEngine(IWindow window)
    {
        _window = window;
    }

    public void Render(IContainer container, Size? defaultSize = null)
    {
        defaultSize ??= new Size(_window.Width, _window.Height);

        var defaultChildrenSize = new Size(
            Width: container.Orientation == Orientation.Horizontal
                ? defaultSize.Width / container.Nodes.Count
                : defaultSize.Width,
            Height: container.Orientation == Orientation.Vertical
                ? defaultSize.Height / container.Nodes.Count
                : defaultSize.Height
        );

        var controlNumber = 0;
        var nodePosition = new NodePosition(Left: 0, Top: 0);

        Debugger.Log(0, "Render", $"{nodePosition} {defaultSize} {container.GetType().Name}\n");
        Helper.ShowBackground(nodePosition, defaultSize);

        while (controlNumber < container.Nodes.Count)
        {
            var node = container.Nodes[controlNumber];

            nodePosition = RenderNode(node, container.Orientation, defaultChildrenSize, nodePosition);
            controlNumber++;
        }
    }


    private NodePosition RenderNode(INode node, Orientation orientation, Size defaultSize, NodePosition position)
    {
        switch (node)
        {
            case IContainer container when orientation == Orientation.Horizontal:
                Render(container, defaultSize);
                return position with
                {
                    Left = position.Left + defaultSize.Width
                };
            case IContainer container when orientation == Orientation.Vertical:
                Render(container, defaultSize);
                return position with
                {
                    Top = position.Top + defaultSize.Height
                };
            case IComponent component when orientation == Orientation.Horizontal:
                var componentWidth = RenderComponent(component, position, defaultSize).Width;
                return position with
                {
                    Left = position.Left + (defaultSize.Width <= componentWidth ? componentWidth : defaultSize.Width)
                };
            case IComponent component when orientation == Orientation.Vertical:
                var componentHeight = RenderComponent(component, position, defaultSize).Height;
                return position with
                {
                    Top = position.Top + (defaultSize.Height <= componentHeight ? componentHeight : defaultSize.Height)
                };
            default:
                throw new InvalidCastException();
        }
    }

    private Size RenderComponent(IComponent component, NodePosition defaultPosition, Size defaultSize)
    {
        var content = component.Render();

        var maxWidth = _window.Width - defaultPosition.Left;
        var maxHeight = _window.Height - defaultPosition.Top;

        var leftPosition = defaultPosition.Left + (int)(component.Padding?.Left ?? 0) +
                           CompensationLeft(component.Alignment.Horizontal, defaultSize, content.GetSize());
        var topPosition = defaultPosition.Top + (int)(component.Padding?.Top ?? 0) +
                          CompensationTop(component.Alignment.Vertical, defaultSize, content.GetSize());


        Debugger.Log(0, "Render", $"{component.GetType().Name} with position [{leftPosition},{topPosition}]\n");

        var rows = content.Rows(maxWidth, maxHeight);

        Helper.ShowBackground(defaultPosition, defaultSize);

        foreach (var row in rows)
        {
            _window.SetCursorPosition(leftPosition, topPosition);
            _window.Write(row);

            topPosition++;
        }

        return content.GetSize();
    }

    private static int CompensationLeft(Horizontal componentHorizontal, Size defaultSize, Size realSize) =>
        componentHorizontal switch
        {
            Horizontal.Left => 0,
            Horizontal.Center => (defaultSize.Width - realSize.Width) / 2,
            Horizontal.Right => defaultSize.Width - realSize.Width,
            _ => 0
        };

    private static int CompensationTop(Vertical componentVertical, Size defaultSize, Size realSize)
        =>
            componentVertical switch
            {
                Vertical.Top => 0,
                Vertical.Center => (defaultSize.Height - realSize.Height) / 2,
                Vertical.Bottom => defaultSize.Height - realSize.Height,
                _ => 0
            };
}