using Moq;
using TUI.Engine.Attributes;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering;
using TUI.Engine.Rendering.Canvas;
using TUI.Engine.Tests.Stubs;

namespace TUI.Engine.Tests.Draw;

public class DrawStaticTests : ComponentBaseTests
{
    [Fact]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawStaticComponentVerticalOrientation()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(6, 4));

        var firstComponent = Prepare.Component();
        firstComponent.SetContent("First");
        firstComponent.SetRelative();
        var secondComponent = Prepare.Component();
        secondComponent.SetContent("Second");
        secondComponent.SetAbsolute();
        var thirdComponent = Prepare.Component();
        thirdComponent.SetContent("Third");

        var root = Prepare.Container(firstComponent, secondComponent, thirdComponent);
        root.SetOrientationVertical();

        Craftsman(canvas).Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionTimes(0, 2, 2);
        Mock.Get(canvas).Verify(w => w.Paint("First"), Times.Once());
        Mock.Get(canvas).Verify(w => w.Paint("Second"), Times.Once());
        Mock.Get(canvas).Verify(w => w.Paint("Third"), Times.Once());
    }
}