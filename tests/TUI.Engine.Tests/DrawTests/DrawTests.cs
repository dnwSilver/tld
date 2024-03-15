using Moq;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Containers;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering;
using TUI.Engine.Rendering.Canvas;
using TUI.Engine.Tests.Stubs;

namespace TUI.Engine.Tests.DrawTests;

public class DrawCraftsmanTests
{
    public TestComponent _component;

    public DrawCraftsmanTests()
    {
        _component = Prepare.Component();
    }

    [Fact]
    public void DrawSimple()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(9, 1));

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(_component, Position.Default, canvas.Size);

        Mock.Get(canvas).Verify(w => w.SetPencil(Position.Default), Times.Once());
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
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(canvasSize, canvasSize));
        var component = Prepare.Component();
        component.SetContent(content);
        component.SetAlignment(Vertical.Top);
        component.SetAlignment(alignment);

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(component, Position.Default, canvas.Size);

        Mock.Get(canvas).Verify(w => w.Paint(content), Times.Once());
        Mock.Get(canvas).Verify(w => w.SetPencil(new Position(expectedPosition, 0)), Times.Once());
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
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(canvasSize, canvasSize));
        _component.SetContent(content);
        _component.SetAlignment(Horizontal.Left);
        _component.SetAlignment(alignment);

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(_component, Position.Default, canvas.Size);

        foreach (var expectedPencilPosition in expectedPositions)
        {
            Mock.Get(canvas).VerifyPositionOnce(0, expectedPencilPosition);
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
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(6, 5));
        _component.SetContent("VV");
        _component.SetAlignment(horizontal);
        _component.SetAlignment(vertical);

        var componentCraftsman = new ComponentCraftsman(canvas);
        componentCraftsman.Draw(_component, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(expectedLeft, expectedTop);
    }

    [Theory]
    [InlineData(Orientation.Horizontal, 9, 1)]
    public void DrawWithOverloadHorizontal(Orientation orientation, int rootWidth, int rootHeight)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(rootWidth, rootHeight));
        var root = Prepare.Container(_component, _component);
        root.SetOrientationHorizontal();

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(4, 0);
        Mock.Get(canvas).Verify(w => w.Paint("Lore"), Times.Exactly(2));
    }


    [Theory]
    [InlineData(4, 4, new[] { 0, 1, 2, 3 })]
    public void DrawWithOverloadVertical(int rootWidth, int rootHeight, int[] expectedTopPositions)
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(rootWidth, rootHeight));
        _component.SetContent("Lorem\nLorem\nLorem");
        var root = Prepare.Container(_component, _component);
        root.SetOrientationVertical();

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.Size);

        foreach (var expectedTopPosition in expectedTopPositions)
        {
            Mock.Get(canvas).VerifyPositionOnce(0, expectedTopPosition);
        }

        Mock.Get(canvas).Verify(w => w.Paint("Lore"), Times.Exactly(rootHeight));
    }

    [Fact]
    public void DrawVerticalWithDoubleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(10, 2));
        var root = Prepare.Container(_component, _component);
        root.SetOrientationVertical();

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        new DrawCraftsman(componentCraftsman, containerCraftsman).Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(0, 1);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Fact]
    public void DrawHorizontalWithDoubleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(10, 1));
        var nodes = new Nodes.Nodes { _component, _component };
        var container = Mock.Of<ContainerBase>(g => g.GetNodes() == nodes);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        new DrawCraftsman(componentCraftsman, containerCraftsman).Draw(container, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(5, 0);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Fact]
    public void DrawWithMultipleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(24, 1));
        var root = Prepare.Container(_component, _component, _component, _component);

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(6, 0);
        Mock.Get(canvas).VerifyPositionOnce(12, 0);
        Mock.Get(canvas).VerifyPositionOnce(18, 0);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(4));
    }

    [Fact]
    public void DrawWithContainerAndComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(10, 2));
        var container = Prepare.Container(_component);
        var root = Prepare.Container(container, _component);
        root.SetAdaptive(Orientation.Vertical);
        root.SetOrientationVertical();

        var componentCraftsman = new ComponentCraftsman(canvas);
        var containerCraftsman = new ContainerCraftsman(componentCraftsman);
        containerCraftsman.Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(0, 1);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }
}