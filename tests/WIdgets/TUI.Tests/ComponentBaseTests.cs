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
            .Set(vertical: Vertical.Center, alignmentHorizontal: AlignmentHorizontal.Center);

        component.Padding.Top.Should().Be(Level.Normal);
        component.Padding.Left.Should().Be(Level.Normal);
        component.Padding.Bottom.Should().Be(Level.Normal);
        component.Padding.Right.Should().Be(Level.Normal);
        component.Alignment.AlignmentHorizontal.Should().Be(AlignmentHorizontal.Center);
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
    [InlineData(AlignmentHorizontal.Left)]
    [InlineData(AlignmentHorizontal.Center)]
    [InlineData(AlignmentHorizontal.Right)]
    public void WhenSetHorizontalAlignSaveAllChange(AlignmentHorizontal alignment)
    {
        var component = new Logo();

        component.Set(alignmentHorizontal: alignment);

        component.Alignment.AlignmentHorizontal.Should().Be(alignment);
    }
    
}