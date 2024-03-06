using FluentAssertions;
using TUI.Components.Controls.Statics;
using TUI.Engine.Nodes.Attributes.Alignments;
using TUI.Engine.Theme;

namespace Widgets.Tests;

public class ComponentBaseTests
{
    [Fact]
    public void WhenUseChainingSaveAllChange()
    {
        var component = new Logo()
            .Set(Level.Normal)
            .Set(vertical: Vertical.Center, horizontal: Horizontal.Center);

        component.Padding.Top.Should().Be(Level.Normal);
        component.Padding.Left.Should().Be(Level.Normal);
        component.Padding.Bottom.Should().Be(Level.Normal);
        component.Padding.Right.Should().Be(Level.Normal);
        component.Alignment.Horizontal.Should().Be(Horizontal.Center);
        component.Alignment.Vertical.Should().Be(Vertical.Center);
    }

    [Fact]
    public void WhenSetPaddingsSaveAllChange()
    {
        var component = new Logo();

        component.Set(Level.Normal);

        component.Padding.Top.Should().Be(Level.Normal);
        component.Padding.Left.Should().Be(Level.Normal);
        component.Padding.Bottom.Should().Be(Level.Normal);
        component.Padding.Right.Should().Be(Level.Normal);
    }

    [Theory]
    [InlineData(Vertical.Bottom)]
    [InlineData(Vertical.Center)]
    [InlineData(Vertical.Top)]
    public void WhenSetVerticalAlignSaveAllChange(Vertical alignment)
    {
        var component = new Logo();

        component.Set(vertical: alignment);

        component.Alignment.Vertical.Should().Be(alignment);
    }

    [Theory]
    [InlineData(Horizontal.Left)]
    [InlineData(Horizontal.Center)]
    [InlineData(Horizontal.Right)]
    public void WhenSetHorizontalAlignSaveAllChange(Horizontal alignment)
    {
        var component = new Logo();

        component.Set(horizontal: alignment);

        component.Alignment.Horizontal.Should().Be(alignment);
    }
}