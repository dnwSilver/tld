using Moq;
using TUI.Engine.Attributes;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering;
using TUI.Engine.Rendering.Canvas;
using TUI.Engine.Tests.Stubs;

namespace TUI.Engine.Tests.Draw;

public class DrawOverloadTests : ComponentBaseTests
{
    [Fact]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawWithOverloadHorizontal()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(9, 1));
        var root = Prepare.Container(Component, Component);
        root.SetOrientationHorizontal();

        Craftsman(canvas).Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(4, 0);
        Mock.Get(canvas).Verify(w => w.Paint("Lore"), Times.Exactly(2));
    }

    [Theory]
    [InlineData(4, 4, new[] { 0, 1, 2, 3 })]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawWithOverloadVertical(int rootWidth, int rootHeight, int[] expectedTopPositions)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(rootWidth, rootHeight));
        Component.SetContent("Lorem\nLorem\nLorem");
        var root = Prepare.Container(Component, Component);
        root.SetOrientationVertical();

        Craftsman(canvas).Draw(root, Position.Default, canvas.Size);

        foreach (var expectedTopPosition in expectedTopPositions)
        {
            Mock.Get(canvas).VerifyPositionOnce(0, expectedTopPosition);
        }

        Mock.Get(canvas).Verify(w => w.Paint("Lore"), Times.Exactly(rootHeight));
    }
}