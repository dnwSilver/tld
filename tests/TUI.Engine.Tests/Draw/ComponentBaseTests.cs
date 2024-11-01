using TUI.Engine.Nodes;
using TUI.Engine.Rendering;
using TUI.Engine.Rendering.Canvas;
using TUI.Engine.Tests.Stubs;

namespace TUI.Engine.Tests.Draw;

public class ComponentBaseTests
{
    protected readonly TestComponent Component = Prepare.Component();

    protected IDrawable<INode> Craftsman(ICanvas canvas)
    {
        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        return new DrawCraftsman(componentCraftsman, containerCraftsman);
    }
}