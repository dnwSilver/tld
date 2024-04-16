using System.Text;
using TUI.Engine.Attributes;
using TUI.Engine.Components;

namespace TUI.Controls.Components;

public class VersionComponent : ComponentBase
{
    private readonly VersionType _type;
    private readonly string _version;
    private readonly Brand? _brand;

    public VersionComponent(VersionType type, string version, Brand? brand)
    {
        _type = type;
        _version = version;
        _brand = brand;
    }

    protected override Sketch DrawComponent(Size minSize)
    {
        var builder = new StringBuilder();

        if (_brand is not null)
        {
            builder.Append(_brand.ColorLogo());
        }

        builder.Append(_version);
        var sketch = builder.ToString();

        return new Sketch(_type.Colorize(sketch));
    }
}