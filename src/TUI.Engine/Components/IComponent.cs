using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Attributes.Paddings;
using TUI.Engine.Nodes;

namespace TUI.Engine.Components;

public interface IComponent : INode, IWithAlignment, IWithPadding
{
    internal bool IsRelative { get; }

    public void SetRelative();

    public void SetAbsolute();

    internal Sketch MakeSketch(Size minSize);
}