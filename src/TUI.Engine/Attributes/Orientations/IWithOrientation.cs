namespace TUI.Engine.Attributes.Orientations;

public interface IWithOrientation
{
    internal Orientation Orientation { get; }

    public void SetOrientationHorizontal();
    public void SetOrientationVertical();
}