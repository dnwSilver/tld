using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Containers;

namespace TUI.Engine.Rendering;

public interface IRenderingEngine
{
    void Render(IContainer container, Size? defaultSize);
}