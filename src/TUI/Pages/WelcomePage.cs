using TUI.Controls.Common;
using TUI.Controls.Components;
using TUI.Controls.Containers;
using TUI.Controls.Layouts;
using TUI.Controls.Statics;
using TUI.Engine.Attributes;
using TUI.Engine.Rendering.Canvas;

namespace TUI.Pages;

public class WelcomePage : PageBase
{
    public override void Render()
    {
        ICanvas canvas = new ConsoleCanvas();

        var header = new StubComponent(new Size(1, 1));

        var logo = new LogoComponent();

        var breadCrumbs = new BreadCrumbsComponent();

        var footer = new FooterContainer(breadCrumbs);

        var layout = new DashboardLayout(header, logo, footer);

        canvas.Draw(layout);
    }

    public override void Bind()
    {
    }
}