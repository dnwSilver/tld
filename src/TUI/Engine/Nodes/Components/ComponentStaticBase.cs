using System.Text;

namespace TUI.Engine.Nodes.Components;

public abstract class ComponentStaticBase : ComponentBase
{
    private Content? _cache;

    protected abstract void RenderWithCache(StringBuilder builder);

    public override Content Render()
    {
        if (_cache is not null)
        {
            return _cache;
        }

        var builder = new StringBuilder();
        RenderWithCache(builder);
        _cache = new Content(builder.ToString());

        return _cache;
    }
}