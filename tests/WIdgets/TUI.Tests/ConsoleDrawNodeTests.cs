using Moq;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Attributes.Orientations;
using TUI.Engine.Nodes.Attributes.Resizing;
using TUI.Engine.Nodes.Containers;
using TUI.Engine.Rendering;

namespace Widgets.Tests;

public class NodeCraftsmanTests
{
    private readonly TestComponent _component;

    public NodeCraftsmanTests()
    {
        _component = new TestComponent();
        _component.SetAlignment(Horizontal.Left);
        _component.SetAlignment(Vertical.Top);
    }

    [Fact]
    public void DrawSimple()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Width == 9 && w.Height == 1);

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(_component, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(0, 0), Times.Once());
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Once());
    }

    [Theory]
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
        var canvas = Mock.Of<ICanvas>(w => w.Width == canvasSize && w.Height == canvasSize);
        _component.SetContent(content);
        _component.SetAlignment(Vertical.Top);
        _component.SetAlignment(alignment);

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(_component, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.Paint(content), Times.Once());
        Mock.Get(canvas).Verify(w => w.SetPencil(expectedPosition, 0), Times.Once());
    }

    [Theory]
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
        var canvas = Mock.Of<ICanvas>(w => w.Width == canvasSize && w.Height == canvasSize);
        _component.SetContent(content);
        _component.SetAlignment(Horizontal.Left);
        _component.SetAlignment(alignment);

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(_component, Position.Default, canvas.GetSize());

        foreach (var expectedCursorPosition in expectedPositions)
        {
            Mock.Get(canvas).Verify(w => w.SetPencil(0, expectedCursorPosition), Times.Once());
        }
    }

    [Theory]
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
        var canvas = Mock.Of<ICanvas>(w => w.Width == 6 && w.Height == 5);
        _component.SetContent("VV");
        _component.SetAlignment(horizontal);
        _component.SetAlignment(vertical);

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(_component, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(expectedLeft, expectedTop), Times.Once());
    }

    [Theory]
    [InlineData(Orientation.Horizontal, 9, 1)]
    [InlineData(Orientation.Vertical, 5, 1)]
    public void DrawWithOverload(Orientation orientation, int rootWidth, int rootHeight)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Width == rootWidth && w.Height == rootHeight);
        var nodes = new Nodes { _component, _component };
        var root = Mock.Of<IContainer>(r => r.GetNodes() == nodes && r.Orientation == orientation);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(0, 0), Times.Once());
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Once());
    }

    [Fact]
    public void DrawVerticalWithDoubleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Height == 2 && w.Width == 10);
        var nodes = new Nodes { _component, _component };
        var root = Mock.Of<IContainer>(r => r.GetNodes() == nodes && r.Orientation == Orientation.Vertical);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        new NodeCraftsman(componentCraftsman, containerCraftsman).Draw(root, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(0, 0), Times.Once());
        Mock.Get(canvas).Verify(w => w.SetPencil(0, 1), Times.Once());
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Fact]
    public void DrawHorizontalWithDoubleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Width == 10 && w.Height == 1);
        var nodes = new Nodes { _component, _component };
        var container = Mock.Of<ContainerBase>(g => g.GetNodes() == nodes);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        new NodeCraftsman(componentCraftsman, containerCraftsman).Draw(container, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(0, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.SetPencil(5, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Fact]
    public void DrawWithMultipleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Width == 24 && w.Height == 1);
        var nodes = new Nodes { _component, _component, _component, _component };
        var root = Mock.Of<IContainer>(r => r.GetNodes() == nodes);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(0, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.SetPencil(6, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.SetPencil(12, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.SetPencil(18, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(4));
    }

    [Fact]
    public void DrawWithContainerAndComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Width == 10 && w.Height == 2);
        var container = Mock.Of<IContainer>(c => c.GetNodes() == new Nodes { _component });
        var nodes = new Nodes { container, _component };
        var root = Mock.Of<IContainer>(r => r.GetNodes() == nodes && r.Orientation == Orientation.Vertical);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(0, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.SetPencil(0, 1), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Theory]
    [InlineData(Resizing.Fixed, 6)]
    [InlineData(Resizing.Fixed, 3)]
    [InlineData(Resizing.Adaptive, 10)]
    public void DrawResizingContainer(Resizing resizing, int expectedCursorPosition)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Width == 20 && w.Height == 2);
        var container =
            Mock.Of<IContainer>(c =>
                c.GetNodes() == new Nodes { _component } && c.ResizingHorizontal == resizing &&
                c.GetFixedSize() == new Size(6, 2));
        var nodes = new Nodes { container, _component };
        var root = Mock.Of<IContainer>(r => r.GetNodes() == nodes && r.Orientation == Orientation.Horizontal);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.GetSize());

        Mock.Get(canvas).Verify(w => w.SetPencil(0, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.SetPencil(expectedCursorPosition, 0), Times.Exactly(1));
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }
}