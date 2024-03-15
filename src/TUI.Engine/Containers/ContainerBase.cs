using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Nodes;
using TUI.Engine.Theme;

namespace TUI.Engine.Containers;

public abstract class ContainerBase : NodeBase, IContainer
{
    private Orientation _orientation = Defaults.Orientation;

    Orientation IWithOrientation.Orientation => _orientation;

    public void SetOrientationHorizontal()
    {
        _orientation = Orientation.Horizontal;
    }

    public void SetOrientationVertical()
    {
        _orientation = Orientation.Vertical;
    }

    public abstract Nodes.Nodes GetNodes();
}