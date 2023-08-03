namespace TUI.Controls;

public interface IControl
{
    void Render(Position position);
}

public interface IControl<in TProps>
{
    void Render(TProps props, Position position);
}