using FluentAssertions;
using TUI.Engine.Rendering;

namespace Widgets.Tests;

public class IntegerTests
{
    [Fact]
    public void IntegerGreaterMax()
    {
        var result = 5.Max(10);

        result.Should().Be(5);
    }
    [Fact]
    public void IntegerLessMax()
    {
        var result = 5.Max(3);

        result.Should().Be(3);
    }
}