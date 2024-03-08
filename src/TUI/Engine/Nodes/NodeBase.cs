using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizing;

namespace TUI.Engine.Nodes;

public abstract class NodeBase : INode
{
    private int _fixedWidth;
    private int _fixedHeight;

    #region Resizing

    public Resizing ResizingHorizontal { get; private set; } = Resizing.Adaptive;
    public Resizing ResizingVertical { get; private set; } = Resizing.Hug;

    public void SetAdaptive(Orientation orientation)
    {
        switch (orientation)
        {
            case Orientation.Horizontal:
                ResizingHorizontal = Resizing.Adaptive;
                break;
            case Orientation.Vertical:
                ResizingVertical = Resizing.Adaptive;
                break;
            default:
                throw new ArgumentOutOfRangeException(nameof(orientation), orientation, null);
        }
    }

    public void SetHug(Orientation orientation)
    {
        switch (orientation)
        {
            case Orientation.Horizontal:
                ResizingHorizontal = Resizing.Hug;
                break;
            case Orientation.Vertical:
                ResizingVertical = Resizing.Hug;
                break;
            default:
                throw new ArgumentOutOfRangeException(nameof(orientation), orientation, null);
        }
    }

    public void SetFixed(Orientation orientation, int value)
    {
        switch (orientation)
        {
            case Orientation.Horizontal:
                ResizingHorizontal = Resizing.Fixed;
                _fixedWidth = value;
                break;
            case Orientation.Vertical:
                ResizingVertical = Resizing.Fixed;
                _fixedHeight = value;
                break;
            default:
                throw new ArgumentOutOfRangeException(nameof(orientation), orientation, null);
        }
    }

    #endregion Resizing
}