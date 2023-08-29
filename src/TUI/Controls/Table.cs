using Pastel;
using TUI.UserInterface;


namespace TUI.Controls;

public record TableProps(IEnumerable<string> HeaderCells,
                         IEnumerable<string> Rows,
                         int TitleWidth,
                         int ColumnWidth);

public class Table : IControl<TableProps>
{
    private readonly Dictionary<int, string> _rows = new();
    private Position _position;
    private int _selectedRowId;

    public void Render(TableProps props, Position position, int? height = 0)
    {
        _position = position;
        Console.SetCursorPosition(_position.Left, _position.Top);

        Console.Write(' '.Repeat(props.TitleWidth));
        foreach (var headerCell in props.HeaderCells)
            Console.Write(' '.Repeat(props.ColumnWidth - headerCell.Width()) + headerCell);
    }

    public void RenderRow(int rowId, string rowText, string? bgColor = default)
    {
        var padRight = ' '.Repeat(Console.WindowWidth - rowText.Width() - Theme.BorderWidth * 2);
        _rows[rowId] = rowText + padRight;

        Console.SetCursorPosition(_position.Left, _position.Top + rowId);
        Console.Write(string.IsNullOrEmpty(bgColor) ? rowText : rowText.PastelBg(bgColor));
    }

    public void Next()
    {
        if (_selectedRowId >= _rows.Count) return;

        RemoveHoverFromCurrentRow();
        RenderRow(++_selectedRowId, _rows[_selectedRowId], Palette.HoverColor);
    }

    private void RemoveHoverFromCurrentRow()
    {
        if (_rows.TryGetValue(_selectedRowId, out var row)) RenderRow(_selectedRowId, row);
    }

    public void Previous()
    {
        if (_selectedRowId == 0) Next();

        if (_selectedRowId == 1) return;

        RemoveHoverFromCurrentRow();
        RenderRow(--_selectedRowId, _rows[_selectedRowId], Palette.HoverColor);
    }
}