using TUI.Engine.Attributes;
using TUI.Engine.Components;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering.Canvas;

namespace TUI.Engine.Rendering;

internal sealed class ComponentCraftsman : CraftsmanBase, IDrawable<IComponent>
{
    private readonly ICanvas _canvas;

    public ComponentCraftsman(ICanvas canvas)
    {
        _canvas = canvas;
    }

    public Size Draw(IComponent component, Position pencil, Size maxSize)
    {
        var sketch = component.MakeSketch(maxSize);
        var sketchSize = sketch.GetSize();

        var correctedPencil = component.CorrectContentPosition(pencil, maxSize, sketchSize);

        Debug(pencil, maxSize);

        foreach (var line in sketch.Crop(maxSize))
        {
            _canvas.SetPencil(correctedPencil);
            _canvas.Paint(line);

            correctedPencil = correctedPencil with { Top = correctedPencil.Top + 1 };
        }

        return sketchSize;
    }
}