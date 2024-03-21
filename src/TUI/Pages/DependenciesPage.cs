using System.Diagnostics;
using TUI.Controls.Containers;
using TUI.Controls.Layouts;
using TUI.Controls.Statics;
using TUI.Domain;
using TUI.Engine.Rendering.Canvas;
using TUI.Providers.Dependencies;

namespace TUI.Pages;

interface IPage
{
    void Open();
    void Render();
    void Bind();
}

public abstract class PageBase : IPage
{
    public void Open()
    {
        Debugger.Log(0, "Event", $"Open page ${GetType().UnderlyingSystemType.Name}\n");
        Bind();
        Render();
    }

    public abstract void Render();

    public abstract void Bind();
}

public class DependenciesPage : PageBase
{
    private IEnumerable<Dependency> ConventionDependencies;

    public override void Render()
    {
        ICanvas canvas = new ConsoleCanvas();

        var header = new HeaderContainer();
        var copyright = new CopyrightComponent();
        var dashboard = new DashboardContainer();
        var dependenciesHeader = new DependenciesContainer();
        dependenciesHeader.AddTitleStub();

        foreach (var conventionDependency in ConventionDependencies)
        {
            dependenciesHeader.AddDependency(conventionDependency);
        }

        dashboard.AddChildren(dependenciesHeader);
        // CommandLine = new CommandLine();
        // DependenciesView = new DependenciesView();

        var layout = new DashboardLayout(header, dashboard, copyright);
        canvas.Draw(layout);
    }

    public override void Bind()
    {
        var repo = new DependencyRepository();
        ConventionDependencies = repo.Read("javascript");
    }

    // private bool _commandLineInDisplay;

    // private ProjectDto _currentProjectDto;

    // private bool _headerInDisplay = true;


    // public string FocusedElement { get; set; } = "";

    // public void OpenDeps(ProjectDto projectDto)
    // {
    //     _currentProjectDto = projectDto;
    //     var dashboardPosition = new ControlPosition(0, Header.Height);
    //     // DependencyDashboard.IsFocused = true;
    //     // CommandLine.IsFocused = false;
    //     DependenciesView.Render(_currentProjectDto, dashboardPosition);
    // }
    //
    // private void ResizeDependencies()
    // {
    //     var topPosition = 0;
    //     topPosition += _commandLineInDisplay ? CommandLine.Height : 0;
    //     topPosition += _headerInDisplay ? Header.Height : 0;
    //     var dashboardPosition = new ControlPosition(0, topPosition);
    //     DependenciesView.Render(_currentProjectDto, dashboardPosition);
    // }
    //
    // public void Render()
    // {
    //     var headerPosition = new ControlPosition(0, 0);
    //     Header.Render(headerPosition);
    //
    //     const string copyrightText = "Kolosov Aleksandr";
    //     var copyrightPosition = new ControlPosition(
    //             Console.WindowWidth - copyrightText.Width(),
    //             Console.WindowHeight);
    //     CopyrightControl.Render(copyrightText, copyrightPosition);
    // }
    //
    // public void ToggleHeader()
    // {
    //     _headerInDisplay = !_headerInDisplay;
    //     if (_headerInDisplay)
    //     {
    //         var headerPosition = new ControlPosition(0, 0);
    //         Header.Render(headerPosition);
    //     }
    //
    //     ResizeDependencies();
    // }
    //
    // public void Next()
    // {
    //     DependenciesView.Next();
    // }
    //
    // public void Previous()
    // {
    //     DependenciesView.Previous();
    // }
    //
    // public void OpenCommandLine()
    // {
    //     var commandLinePosition = new ControlPosition(0, _headerInDisplay ? Header.Height : 0);
    //     // CommandLine.IsFocused = true;
    //     // DependencyDashboard.IsFocused = false;
    //     FocusedElement = nameof(CommandLine);
    //     CommandLine.Render(commandLinePosition);
    //     _commandLineInDisplay = true;
    //     ResizeDependencies();
    //     Console.SetCursorPosition(commandLinePosition.Left + Theme.Padding + Theme.BorderWidth + 2,
    //             commandLinePosition.Top + Theme.BorderWidth);
    //     Console.CursorVisible = true;
    // }
}