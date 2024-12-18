using System.Text;
using TUI.Engine.Attributes;

namespace TUI.Engine.Components;

public sealed class Sketch : IEnumerable<string>
{
    private IEnumerable<string> ContentRows { get; }

    public Sketch(string content) => ContentRows = content.Split(Symbols.LineBreak);

    public Sketch(StringBuilder builder) => ContentRows = builder.ToString().Split(Symbols.LineBreak);

    public IEnumerator<string> GetEnumerator() => ContentRows.GetEnumerator();

    public IEnumerable<string> Crop(Size maxSize) => ContentRows
        .Select(row => maxSize.Width < row.GetWidth() ? row[..maxSize.Width] : row)
        .Take(maxSize.Height)
        .ToArray();

    public Size GetSize()
    {
        var width = ContentRows.Select(row => row.GetWidth()).DefaultIfEmpty(0).Max();
        var height = ContentRows.Count();
        return new Size(width, height);
    }

    public override string ToString() => string.Join(Symbols.LineBreak, ContentRows);

    System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator() => GetEnumerator();
}