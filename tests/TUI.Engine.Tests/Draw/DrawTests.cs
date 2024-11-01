using Moq;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Orientations;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering;
using TUI.Engine.Rendering.Canvas;
using TUI.Engine.Tests.Stubs;

namespace TUI.Engine.Tests.Draw;

public class DrawCraftsmanTests : ComponentBaseTests
{
    [Fact]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawSimple()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(9, 1));

        Craftsman(canvas).Draw(Component, Position.Default, canvas.Size);

        Mock.Get(canvas).Verify(w => w.SetPencil(Position.Default), Times.Once());
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Once());
    }

    [Fact]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawVerticalWithDoubleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(10, 2));
        var root = Prepare.Container(Component, Component);
        root.SetOrientationVertical();

        Craftsman(canvas).Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(0, 1);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Fact]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawHorizontalWithDoubleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(10, 1));
        var container = Prepare.Container(Component, Component);
        container.SetOrientationHorizontal();

        Craftsman(canvas).Draw(container, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(5, 0);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }

    [Fact]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawWithMultipleComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(24, 1));
        var root = Prepare.Container(Component, Component, Component, Component);

        Craftsman(canvas).Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(6, 0);
        Mock.Get(canvas).VerifyPositionOnce(12, 0);
        Mock.Get(canvas).VerifyPositionOnce(18, 0);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(4));
    }

    [Fact]
    [Trait("Category", nameof(IDrawable<INode>.Draw))]
    public void DrawWithContainerAndComponent()
    {
        var canvas = Mock.Of<ICanvas>(w => w.Size == new Size(10, 2));
        var container = Prepare.Container(Component);
        var root = Prepare.Container(container, Component);
        root.SetAdaptive(Orientation.Vertical);
        root.SetOrientationVertical();

        Craftsman(canvas).Draw(root, Position.Default, canvas.Size);

        Mock.Get(canvas).VerifyPositionOnce(Position.Default);
        Mock.Get(canvas).VerifyPositionOnce(0, 1);
        Mock.Get(canvas).Verify(w => w.Paint("Lorem"), Times.Exactly(2));
    }
}