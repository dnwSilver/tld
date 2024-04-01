namespace TUI.Domain;

public record Project(int Id, string Name, IEnumerable<string> Tags, string Hub)
{
    private IEnumerable<Dependency> Dependencies => new List<Dependency>();

    public bool IsPublicNetwork => Tags.Contains("public");

    public bool HasAuth => Tags.Contains("auth");

    public bool SeoDependent => Tags.Contains("seo");
}