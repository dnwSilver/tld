using System.Diagnostics;
using TUI.Components.Controls;
using TUI.Components.Layouts;
using TUI.Engine.Rendering.Canvas;

namespace TUI.Pages;

public class DependenciesPage
{
    public void Open()
    {
        Debugger.Log(0, "Event", "Open page dependencies\n");

        ICanvas canvas = new ConsoleCanvas();

        var header = new HeaderContainer();
        var copyright = new Copyright();
        var dashboard = new Dashboard("Dependencies");
        var layout = new DashboardLayout();
        layout.AddHeader(header);
        layout.AddFooter(copyright);
        layout.AddDashboard(dashboard);
        // CommandLine = new CommandLine();
        // DependenciesView = new DependenciesView();

        canvas.Draw(layout);
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