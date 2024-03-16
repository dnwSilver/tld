using Moq;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering;
using TUI.Engine.Rendering.Canvas;

namespace TUI.Engine.Tests.Draw;

public class DrawAlignmentTests : ComponentBaseTests
{
    [Theory]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    [InlineData(Horizontal.Left, "Lorem", 10, 0)]
    [InlineData(Horizontal.Center, "Lorem", 10, 2)]
    [InlineData(Horizontal.Center, "Lo", 10, 4)]
    [InlineData(Horizontal.Center, "Lorem", 9, 2)]
    [InlineData(Horizontal.Center, "Lorem", 11, 3)]
    [InlineData(Horizontal.Right, "Lorem", 10, 5)]
    [InlineData(Horizontal.Right, "Lo", 10, 8)]
    public void DrawWithHorizontalAlignment(Horizontal alignment, string content, int canvasSize,
        int expectedPosition)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(canvasSize, canvasSize));
        Component.SetContent(content);
        Component.SetAlignment(Vertical.Top);
        Component.SetAlignment(alignment);

        Craftsman(canvas).Draw(Component, Position.Default, canvas.Size);

        Mock.Get(canvas).Verify(w => w.Paint(content), Times.Once());
        Mock.Get(canvas).Verify(w => w.SetPencil(new Position(expectedPosition, 0)), Times.Once());
    }

    [Theory]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    [InlineData(Vertical.Top, "v", 5, new[] { 0 })]
    [InlineData(Vertical.Top, "v\nv", 5, new[] { 0, 1 })]
    [InlineData(Vertical.Top, "v\nv\nv", 5, new[] { 0, 1, 2 })]
    [InlineData(Vertical.Center, "v", 1, new[] { 0 })]
    [InlineData(Vertical.Center, "v", 4, new[] { 1 })]
    [InlineData(Vertical.Center, "v", 5, new[] { 2 })]
    [InlineData(Vertical.Center, "v", 6, new[] { 2 })]
    [InlineData(Vertical.Center, "v\nv", 4, new[] { 1, 2 })]
    [InlineData(Vertical.Center, "v\nv", 5, new[] { 1, 2 })]
    [InlineData(Vertical.Center, "v\nv", 6, new[] { 2, 3 })]
    [InlineData(Vertical.Bottom, "v", 5, new[] { 4 })]
    [InlineData(Vertical.Bottom, "v\nv", 2, new[] { 0, 1 })]
    [InlineData(Vertical.Bottom, "v\nv", 3, new[] { 1, 2 })]
    [InlineData(Vertical.Bottom, "v\nv\nv\nv", 5, new[] { 1, 2, 3, 4 })]
    public void DrawWithVerticalAlignment(Vertical alignment, string content, int canvasSize, int[] expectedPositions)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(canvasSize, canvasSize));
        Component.SetContent(content);
        Component.SetAlignment(Horizontal.Left);
        Component.SetAlignment(alignment);

        Craftsman(canvas).Draw(Component, Position.Default, canvas.Size);

        foreach (var expectedPencilPosition in expectedPositions)
        {
            Mock.Get(canvas).VerifyPositionOnce(0, expectedPencilPosition);
        }
    }

    [Theory]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    [InlineData(Horizontal.Left, Vertical.Top, 0, 0)]
    [InlineData(Horizontal.Left, Vertical.Center, 0, 2)]
    [InlineData(Horizontal.Left, Vertical.Bottom, 0, 4)]
    [InlineData(Horizontal.Center, Vertical.Top, 2, 0)]
    [InlineData(Horizontal.Center, Vertical.Center, 2, 2)]
    [InlineData(Horizontal.Center, Vertical.Bottom, 2, 4)]
    [InlineData(Horizontal.Right, Vertical.Top, 4, 0)]
    [InlineData(Horizontal.Right, Vertical.Center, 4, 2)]
    [InlineData(Horizontal.Right, Vertical.Bottom, 4, 4)]
    public void DrawWithAlignment(Horizontal horizontal, Vertical vertical, int expectedLeft,
        int expectedTop)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(6, 5));
        Component.SetContent("VV");
        Component.SetAlignment(horizontal);
        Component.SetAlignment(vertical);

        Craftsman(canvas).Draw(Component, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(expectedLeft, expectedTop);
    }
}