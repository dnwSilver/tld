using System.Text;
using TUI.Domain;
using TUI.Engine.Attributes;
using TUI.Engine.Components;

namespace TUI.Controls.Components;

public class VersionComponent : ComponentBase
{
    private readonly VersionType _type;
    private readonly string _version;
    private readonly Brand? _brand;
    private readonly VersionStatus _status;
    
    public VersionComponent(string version, Brand brand, VersionStatus status = VersionStatus.SoGood,
        VersionType type = VersionType.Release)
    {
        _version = version;
        _brand = brand;
        _status = status;
        _type = type;
    }
    
    protected override Sketch DrawComponent(Size minSize)
    {
        var builder = new StringBuilder();
        
        // builder.Append(_type.ToString());
        
        if (_brand is not null)
        {
            builder.Append(_brand.ColorLogo());
        }
        
        builder.Append(_version);
        var sketch = builder.ToString();
        
        return new Sketch(_status.Colorize(sketch));
    }
}