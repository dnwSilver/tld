using Moq;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Nodes.Components;
using TUI.Engine.Nodes.Containers;
using TUI.Engine.Rendering;

namespace Widgets.Tests;

public class ConsoleRenderingEngineTests
{
    private readonly IComponent _component;

    public ConsoleRenderingEngineTests()
    {
        _component = Mock.Of<IComponent>(c =>
            c.Render() == new Content("Lorem") &&
            c.Alignment == new Alignment(Horizontal.Left, Vertical.Top));
    }

    [Fact]
    public void RenderSimple()
    {
        var window = Mock.Of<IWindow>(w => w.Width == 9 && w.Height == 1);
        var nodes = new Nodes { _component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes);

        new ConsoleRenderingEngine(window).Render(root);

        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 0), Times.Once());
        Mock.Get(window).Verify(w => w.Write("Lorem"), Times.Once());
    }

    [Theory]
    [InlineData(Horizontal.Left, "Lorem", 10, 0)]
    [InlineData(Horizontal.Center, "Lorem", 10, 2)]
    [InlineData(Horizontal.Center, "Lo", 10, 4)]
    [InlineData(Horizontal.Center, "Lorem", 9, 2)]
    [InlineData(Horizontal.Center, "Lorem", 11, 3)]
    [InlineData(Horizontal.Right, "Lorem", 10, 5)]
    [InlineData(Horizontal.Right, "Lo", 10, 8)]
    public void RenderWithHorizontalAlignment(Horizontal alignment, string content, int windowSize,
        int expectedPosition)
    {
        var window = Mock.Of<IWindow>(w => w.Width == windowSize && w.Height == windowSize);
        var component = Mock.Of<IComponent>(c => c.Render() == new Content(content) &&
                                                 c.Alignment == new Alignment(alignment, Vertical.Top));
        var nodes = new Nodes { component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes);

        new ConsoleRenderingEngine(window).Render(root);

        Mock.Get(window).Verify(w => w.Write(content), Times.Once());
        Mock.Get(window).Verify(w => w.SetCursorPosition(expectedPosition, 0), Times.Once());
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
    public void RenderWithVerticalAlignment(Vertical alignment, string content, int windowSize, int[] expectedPositions)
    {
        var window = Mock.Of<IWindow>(w => w.Width == windowSize && w.Height == windowSize);
        var component = Mock.Of<IComponent>(c => c.Render() == new Content(content) &&
                                                 c.Alignment == new Alignment(Horizontal.Left, alignment));
        var nodes = new Nodes { component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes);

        new ConsoleRenderingEngine(window).Render(root);

        foreach (var expectedCursorPosition in expectedPositions)
        {
            Mock.Get(window).Verify(w => w.SetCursorPosition(0, expectedCursorPosition), Times.Once());
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
    public void RenderWithAlignment(Horizontal horizontal, Vertical vertical, int expectedLeft, int expectedTop)
    {
        var window = Mock.Of<IWindow>(w => w.Width == 6 && w.Height == 5);
        var component = Mock.Of<IComponent>(c => c.Render() == new Content("VV") &&
                                                 c.Alignment == new Alignment(horizontal, vertical));
        var nodes = new Nodes { component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes);

        new ConsoleRenderingEngine(window).Render(root);

        Mock.Get(window).Verify(w => w.SetCursorPosition(expectedLeft, expectedTop), Times.Once());
    }

    [Theory]
    [InlineData(Orientation.Horizontal, 9, 1)]
    [InlineData(Orientation.Vertical, 5, 1)]
    public void RenderWithOverload(Orientation orientation, int rootWidth, int rootHeight)
    {
        var window = Mock.Of<IWindow>(w => w.Width == rootWidth && w.Height == rootHeight);
        var nodes = new Nodes { _component, _component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes && r.Orientation == orientation);

        new ConsoleRenderingEngine(window).Render(root);

        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 0), Times.Once());
        Mock.Get(window).Verify(w => w.Write("Lorem"), Times.Once());
    }

    [Fact]
    public void RenderVerticalWithDoubleComponent()
    {
        var window = Mock.Of<IWindow>(w => w.Height == 2 && w.Width == 10);
        var nodes = new Nodes { _component, _component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes && r.Orientation == Orientation.Vertical);

        new ConsoleRenderingEngine(window).Render(root);

        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 0), Times.Once());
        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 1), Times.Once());
        Mock.Get(window).Verify(w => w.Write("Lorem"), Times.Exactly(2));
    }

    [Fact]
    public void RenderHorizontalWithDoubleComponent()
    {
        var window = Mock.Of<IWindow>(w => w.Width == 10 && w.Height == 1);
        var nodes = new Nodes { _component, _component };
        var container = Mock.Of<IContainer>(g => g.Nodes == nodes);

        new ConsoleRenderingEngine(window).Render(container);

        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 0), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.SetCursorPosition(5, 0), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.Write("Lorem"), Times.Exactly(2));
    }

    [Fact]
    public void RenderWithMultipleComponent()
    {
        var window = Mock.Of<IWindow>(w => w.Width == 24 && w.Height == 1);
        var nodes = new Nodes { _component, _component, _component, _component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes);

        new ConsoleRenderingEngine(window).Render(root);

        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 0), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.SetCursorPosition(6, 0), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.SetCursorPosition(12, 0), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.SetCursorPosition(18, 0), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.Write("Lorem"), Times.Exactly(4));
    }

    [Fact]
    public void RenderWithContainerAndComponent()
    {
        var window = Mock.Of<IWindow>(w => w.Width == 10 && w.Height == 2);
        var container = Mock.Of<IContainer>(c => c.Nodes == new Nodes { _component });
        var nodes = new Nodes { container, _component };
        var root = Mock.Of<IContainer>(r => r.Nodes == nodes && r.Orientation == Orientation.Vertical);

        new ConsoleRenderingEngine(window).Render(root);

        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 0), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.SetCursorPosition(0, 1), Times.Exactly(1));
        Mock.Get(window).Verify(w => w.Write("Lorem"), Times.Exactly(2));
    }
}