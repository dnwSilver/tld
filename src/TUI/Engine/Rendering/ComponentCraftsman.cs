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

    public Size Draw(IComponent component, Position pencil, Size maxSize)
    {
        var sketch = component.MakeSketch();
        var sketchSize = sketch.GetSize();

        var correctedPencil = component.CorrectPosition(pencil, maxSize, sketchSize);

        Debug(correctedPencil, pencil, maxSize);

        foreach (var line in sketch.Crop(maxSize))
        {
            _canvas.SetPencil(correctedPencil.Left, correctedPencil.Top);
            _canvas.Paint(line);

            correctedPencil = correctedPencil with { Top = correctedPencil.Top + 1 };
        }

        return sketchSize;
    }
}