using FluentAssertions;

namespace TUI.Engine.Tests.Draw;

public class IntegerTests
{
    [Theory]
    [Trait("Category", "Primitives")]
    [InlineData(5, 10, 5)]
    [InlineData(5, 5, 5)]
    [InlineData(5, 3, 3)]
    public void Max(int value, int max, int expected)
    {
        var result = value.Max(max);
        result.Should().Be(expected);
    }

    [Theory]
    [Trait("Category", "Primitives")]
    [InlineData(5, 10, 10)]
    [InlineData(5, 5, 5)]
    [InlineData(5, 3, 5)]
    public void Min(int value, int min, int expected)
    {
        var result = value.Min(min);
        result.Should().Be(expected);
    }
}