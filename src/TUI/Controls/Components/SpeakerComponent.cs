using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Components;
using TUI.Engine.Nodes;

namespace TUI.Controls.Components;

public class SpeakerComponent : ComponentBase
{
    private string _message = "";

    private Position? _pencil;

    private SpeakerComponent()
    {
    }

    public static SpeakerComponent Instance { get; } = new();

    protected override Sketch DrawComponent(Size minSize)
    {
        return new Sketch(_message);
    }

    private void Clear()
    {
        _message = new string(' ', DrawContext.MaxSize.Width);
        DrawContext.Canvas.Draw(Instance, _pencil, DrawContext.MaxSize);
    }

    public void Shout(string emoji, string message)
    {
        if (DrawContext is null)
        {
            return;
        }

        _pencil ??= DrawContext.Pencil with { };

        Clear();
        _message = emoji + Symbols.Space + message;
        DrawContext.Canvas.Draw(Instance, _pencil, DrawContext.MaxSize);

        Task.Delay(2000).ContinueWith(_ =>
        {
            Clear();
            return Task.CompletedTask;
        });
    }
}