using System.Diagnostics;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Engine.Rendering;

public sealed class ComponentCraftsman : IDrawable<IComponent>
{
    private readonly ICanvas _canvas;

    public ComponentCraftsman(ICanvas canvas)
    {
        _canvas = canvas;
    }

    public Size Draw(IComponent component, Position sketchPosition, Size allowableSize)
    {
        var sketch = component.Draw();
        var actualSize = sketch.GetSize();

        var maxWidth = _canvas.Width - sketchPosition.Left;
        var maxHeight = _canvas.Height - sketchPosition.Top;

        var pencilPosition = component.GetPosition(sketchPosition, allowableSize, actualSize);

        Debugger.Log(0, "Render", $"{pencilPosition}{component.GetType().Name}.\n");
        Helper.ShowBackground(sketchPosition, allowableSize);

        foreach (var row in sketch.Rows(maxWidth, maxHeight))
        {
            _canvas.SetPencil(pencilPosition.Left, pencilPosition.Top);
            _canvas.Paint(row);

            pencilPosition = pencilPosition with { Top = pencilPosition.Top + 1 };
        }

        return actualSize;
    }
}