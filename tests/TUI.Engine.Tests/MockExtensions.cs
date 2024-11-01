using Moq;
using TUI.Engine.Nodes;
using TUI.Engine.Rendering.Canvas;

namespace TUI.Engine.Tests;

public static class MockExtensions
{
    public static void VerifyPositionOnce<T>(this Mock<T> mock, int left, int top) where T : class, ICanvas
    {
        mock.Verify(w => w.SetPencil(new Position(left, top)), Times.Exactly(1));
    }
    public static void VerifyPositionTimes<T>(this Mock<T> mock, int left, int top, int times) where T : class, ICanvas
    {
        mock.Verify(w => w.SetPencil(new Position(left, top)), Times.Exactly(times));
    }

    public static void VerifyPositionOnce<T>(this Mock<T> mock, Position position) where T : class, ICanvas
    {
        mock.Verify(w => w.SetPencil(position), Times.Exactly(1));
    }
}