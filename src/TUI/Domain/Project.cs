namespace TUI.Domain;

public record Project(int Id, string Name, IEnumerable<string> Tags, Hub Hub)
{
    private IEnumerable<Dependency> Dependencies => new List<Dependency>();

    public bool IsPublicNetwork => Tags.Contains("public");

    public bool HasAuth => Tags.Contains("auth");

    public bool SeoDependent => Tags.Contains("seo");

    public bool Legacy => Tags.Contains("legacy");

    public bool Freeze => Tags.Contains("freeze");
}