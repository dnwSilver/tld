using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;

namespace TUI.Engine.Rendering;

public sealed class ComponentCraftsman : CraftsmanBase, IDrawable<IComponent>
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
        var maxSize = _canvas.GetSize() - sketchPosition;
        var pencilPosition = component.GetPosition(sketchPosition, allowableSize, actualSize);

        foreach (var row in sketch.Rows(maxSize))
        {
            _canvas.SetPencil(pencilPosition.Left, pencilPosition.Top);
            _canvas.Paint(row);

            pencilPosition = pencilPosition with { Top = pencilPosition.Top + 1 };
        }

        Debug(pencilPosition, sketchPosition, allowableSize);

        return actualSize;
    }
}