using Moq;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering;
using TUI.Engine.Rendering.Canvas;
using TUI.Engine.Tests.Stubs;

namespace TUI.Engine.Tests.DrawTests;

public class DrawResizingTests
{
    private readonly ICanvas _canvas;
    private readonly TestContainer _container;
    private readonly ContainerCraftsman _craftsman;
    private readonly TestContainer _root;

    public DrawResizingTests()
    {
        var component = Prepare.Component();
        _canvas = Mock.Of<ICanvas>(w => w.Size == new Size(20, 2));
        _container = Prepare.Container(component);

        _root = Prepare.Container(_container, component);
        _root.SetOrientationHorizontal();

        var componentCraftsman = new ComponentCraftsman(_canvas);
        _craftsman = new ContainerCraftsman(componentCraftsman);
    }

    [Fact]
    public void DrawResizingFixedContainer()
    {
        _container.SetFixed(Orientation.Horizontal, 6);
        _container.SetFixed(Orientation.Vertical, 2);

        _craftsman.Draw(_root, Position.Default, _canvas.Size);

        Mock.Get(_canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(_canvas).VerifyPositionOnce(6, 0);
        Mock.Get(_canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Fact]
    public void DrawResizingAdaptiveContainer()
    {
        _container.SetAdaptive(Orientation.Horizontal);

        _craftsman.Draw(_root, Position.Default, _canvas.Size);

        Mock.Get(_canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(_canvas).VerifyPositionOnce(10, 0);
        Mock.Get(_canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }
}