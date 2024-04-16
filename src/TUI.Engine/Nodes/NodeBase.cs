using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Attributes.Resizings;
using TUI.Engine.Theme;

namespace TUI.Engine.Nodes;

public abstract class NodeBase : INode
{
    private int _fixedWidth;
    private int _fixedHeight;

    Size IResizable.GetFixedSize() => new(_fixedWidth, _fixedHeight);

    #region Resizing

    private Resizing ResizingHorizontal { get; set; } = Defaults.HorizontalResizing;

    Resizing IResizable.ResizingHorizontal => ResizingHorizontal;

    private Resizing ResizingVertical { get; set; } = Defaults.VerticalResizing;

    Resizing IResizable.ResizingVertical => ResizingVertical;

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

    public DrawContext? DrawContext { get; set; }
}