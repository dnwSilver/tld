using System.Text;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Components;

namespace TUI.Controls.Components;

public class CellsComponentBase : ComponentBase, IComponent
{
    private const int MaxCellWidth = 10;

    private readonly IEnumerable<string> _cells;

    public CellsComponentBase(IEnumerable<string> cells)
    {
        _cells = cells;
    }

    public void Render(Horizontal horizontal, Size size)
    {
        var content = new StringBuilder();
        foreach (var cell in _cells)
        {
            content.Append(Symbols.Space.Repeat(MaxCellWidth - cell.GetWidth()));
            content.Append(cell);
        }

        // base.Render(content, position, size);
    }

    protected override Sketch DrawComponent(Size minSize)
    {
        throw new NotImplementedException();
    }
}