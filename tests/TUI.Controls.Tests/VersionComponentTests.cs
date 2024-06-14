using FluentAssertions;
using TUI.Controls.Components;
using TUI.Domain;
using TUI.Engine.Attributes;
using TUI.Engine.Components;

namespace TUI.Engine.Tests
{
    public class VersionComponentTests
    {
        [Theory]
        [Trait("Category", nameof(Sketch))]
        [InlineData(VersionStatus.BeNice, "\u001b[38;2;132;186;100m10.12.33\u001b[0m")]
        [InlineData(VersionStatus.SoGood, "\u001b[38;2;113;121;126m10.12.33\u001b[0m")]
        [InlineData(VersionStatus.ToNew, "\u001b[38;2;37;121;159m10.12.33\u001b[0m")]
        [InlineData(VersionStatus.TooOld, "\u001b[38;2;236;151;6m10.12.33\u001b[0m")]
        public void DrawSketchVersionTypes(VersionStatus versionStatus, string expected)
        {
            var brand = new Brand("Docker", "󰡨", "#1d63ed");
            var version = new VersionComponent("10.12.33", brand, versionStatus);
            
            var sketch = (version as IComponent).MakeSketch(new Size(10, 2));
            
            sketch.ToString().Should().Be(expected);
        }
        
        [Theory]
        [Trait("Category", nameof(Dependency))]
        [InlineData("1.0.0", "0.0.1", VersionStatus.ToNew)]
        [InlineData("1.0.0", "0.1.1", VersionStatus.ToNew)]
        [InlineData("1.0.0", "0.1.0", VersionStatus.ToNew)]
        [InlineData("1.2.0", "1.0.0", VersionStatus.ToNew)]
        [InlineData("1.2.0", "1.0.1", VersionStatus.ToNew)]
        [InlineData("1.2.0", "1.1.0", VersionStatus.ToNew)]
        [InlineData("1.0.0", "1.0.0-rc", VersionStatus.ToNew)]
        [InlineData("1.0.0", "1.0.0", VersionStatus.SoGood)]
        [InlineData("^1.0.0", "1.0.0", VersionStatus.SoGood)]
        [InlineData("1.2.0", "1.3.0", VersionStatus.BeNice)]
        [InlineData("1.3.1", "1.3.3", VersionStatus.BeNice)]
        [InlineData("1.2.0", "2.1.0", VersionStatus.TooOld)]
        [InlineData("1.2.0", "2.0.1", VersionStatus.TooOld)]
        [InlineData("1.2.0", "2.3.1", VersionStatus.TooOld)]
        public void ComparisonDependencies(string actual, string convention, VersionStatus expectedType)
        {
            var brand = new Brand("Poker", "󱢢", "#1d63ed");
            var actualDependency = new Dependency(actual, brand);
            var conventionDependency = new Dependency(convention, brand);
            
            var status = actualDependency.Comparison(conventionDependency);
            
            status.Should().Be(expectedType);
        }
    }
}