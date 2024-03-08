using TUI.Engine.Nodes.Attributes;

namespace TUI.Engine.Nodes.Components;

public sealed class Sketch : IEnumerable<string>
{
    private IEnumerable<string> ContentRows { get; }

    public Sketch(string content) => ContentRows = content.Split(Symbols.LineBreak);

    public IEnumerator<string> GetEnumerator() => ContentRows.GetEnumerator();

    public IEnumerable<string> Rows(int maxWidth, int maxHeight) =>
        ContentRows.Where(row => maxWidth >= row.Width()).Take(maxHeight).ToArray();

    public Size GetSize()
    {
        var width = ContentRows.Select(row => row.Width()).DefaultIfEmpty(0).Max();
        var height = ContentRows.Count();
        return new Size(width, height);
    }

    public override string ToString() => string.Join(Symbols.LineBreak, ContentRows);

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();
}