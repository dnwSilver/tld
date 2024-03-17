using FluentAssertions;
using TUI.Controls.Components;
using TUI.Engine.Attributes;
using TUI.Engine.Components;

namespace TUI.Engine.Tests
{
    public class VersionComponentTests
    {
        [Theory]
        [Trait("Category", nameof(Sketch))]
        [InlineData(VersionType.Convention, "\u001b[38;2;132;186;100m10.12.33\u001b[0m")]
        [InlineData(VersionType.BeNice, "\u001b[38;2;132;186;100m10.12.33\u001b[0m")]
        [InlineData(VersionType.SoGood, "\u001b[38;2;113;121;126m10.12.33\u001b[0m")]
        [InlineData(VersionType.ToNew, "\u001b[38;2;37;121;159m10.12.33\u001b[0m")]
        [InlineData(VersionType.TooOld, "\u001b[38;2;236;151;6m10.12.33\u001b[0m")]
        public void DrawSketchVersionTypes(VersionType versionType, string expected)
        {
            var version = new VersionComponent(versionType, "10.12.33");

            var sketch = (version as IComponent).MakeSketch(new Size(10, 2));

            sketch.ToString().Should().Be(expected);
        }
    }
}