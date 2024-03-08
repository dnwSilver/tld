using System.Text;

namespace TUI.Engine.Nodes.Components;

public abstract class ComponentStaticBase : ComponentBase
{
    private Sketch? _cache;

    protected abstract void RenderWithCache(StringBuilder builder);

    public override Sketch Draw()
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