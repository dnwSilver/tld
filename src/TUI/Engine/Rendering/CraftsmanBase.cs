using System.Diagnostics;
using TUI.Engine.Nodes;
using TUI.Engine.Nodes.Attributes;

namespace TUI.Engine.Rendering;

public abstract class CraftsmanBase
{
    protected void Debug(Position pencilPosition, Size allowableSize)
    {
        Debugger.Log(0, "Draw", $"{pencilPosition}{GetType().Name}.\n");
        Helper.ShowBackground(pencilPosition, allowableSize);
    }
}