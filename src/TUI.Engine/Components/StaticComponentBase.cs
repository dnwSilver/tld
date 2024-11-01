using System.Text;
using TUI.Engine.Attributes;

namespace TUI.Engine.Components;

public abstract class StaticComponentBase : ComponentBase
{
    private Sketch? _cache;

    protected abstract void RenderWithCache(StringBuilder builder);

    protected override Sketch DrawComponent(Size sketchMinSize)
    {
        if (_cache is not null)
        {
            return _cache;
        }

        var builder = new StringBuilder();

        RenderWithCache(builder);

        _cache = new Sketch(builder.ToString());

        return _cache;
    }
}