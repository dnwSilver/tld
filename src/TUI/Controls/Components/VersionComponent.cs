using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Components;

namespace TUI.Controls.Components;

public class VersionComponent : ComponentBase
{
    private readonly VersionType _type;
    private readonly string _version;
    private readonly string? _icon;

    public VersionComponent(VersionType type, string version, string? icon = null)
    {
        _type = type;
        _version = version;
        _icon = icon;
    }

    protected override Sketch DrawComponent(Size minSize)
    {
        var builder = new StringBuilder();

        if (_icon is not null)
        {
            builder.Append(_icon.Colorized());
            builder.Append(Symbols.Space);
        }

        builder.Append(_version);
        var sketch = builder.ToString();

        return new Sketch(_type.Colorize(sketch));
    }
}