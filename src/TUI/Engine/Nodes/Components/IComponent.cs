using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Paddings;

namespace TUI.Engine.Nodes.Components;

public interface IComponent : INode, 
    IWithAlignment<IComponent>, 
    IWithPadding<IComponent>
{
    Content Render();
}