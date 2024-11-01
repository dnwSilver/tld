using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Nodes;

namespace TUI.Engine.Tests.Stubs;

public static class Prepare
{
    public static TestComponent Component()
    {
        var testComponent = new TestComponent();
        testComponent.SetAlignment(Horizontal.Left);
        testComponent.SetAlignment(Vertical.Top);
        return testComponent;
    }

    public static TestContainer Container(params INode[] nodes)
    {
        var testContainer = new TestContainer();
        testContainer.SetNodes(nodes);
        return testContainer;
    }
}