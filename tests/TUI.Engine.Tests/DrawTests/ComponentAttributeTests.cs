using FluentAssertions;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Tests.Stubs;
using TUI.Engine.Theme;

namespace TUI.Engine.Tests.DrawTests;

public class ComponentAttributeTests
{
    [Fact]
    public void WhenUseChainingSaveAllChange()
    {
        var logo = new TestComponent();
        logo.SetPadding(Level.Normal);
        logo.SetAlignment(Vertical.Center);
        logo.SetAlignment(Horizontal.Center);

        logo.Padding.Top.Should().Be(Level.Normal);
        logo.Padding.Left.Should().Be(Level.Normal);
        logo.Padding.Bottom.Should().Be(Level.Normal);
        logo.Padding.Right.Should().Be(Level.Normal);
        logo.Alignment.Horizontal.Should().Be(Horizontal.Center);
        logo.Alignment.Vertical.Should().Be(Vertical.Center);
    }

    [Fact]
    public void WhenSetPaddingsSaveAllChange()
    {
        var component = new TestComponent();

        component.SetPadding(Level.Normal);

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
        var component = new TestComponent();

        component.SetAlignment(alignment);

        component.Alignment.Vertical.Should().Be(alignment);
    }

    [Theory]
    [InlineData(Horizontal.Left)]
    [InlineData(Horizontal.Center)]
    [InlineData(Horizontal.Right)]
    public void WhenSetHorizontalAlignSaveAllChange(Horizontal alignment)
    {
        var component = new TestComponent();

        component.SetAlignment(alignment);

        component.Alignment.Horizontal.Should().Be(alignment);
    }
}