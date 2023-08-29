namespace TUI.Controls;

public interface IControl
{
    void Render(Position position);
}

public interface IControl<in TProps>
{
    // bool IsFocused { get; }

    void Render(TProps props, Position position, int? height);
}