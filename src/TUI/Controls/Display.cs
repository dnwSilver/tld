using TUI.Dashboards;
using TUI.Domain;
using TUI.UserInterface;


namespace TUI.Controls;

public class Display
{
    public readonly CommandLine CommandLine;

    public readonly Copyright Copyright;

    public readonly DependencyDashboard DependencyDashboard;
    public readonly Header Header;
    private bool _commandLineInDisplay;

    private Project _currentProject;
    private bool _headerInDisplay = true;

    public Display()
    {
        Header = new Header();
        Copyright = new Copyright();
        CommandLine = new CommandLine();
        DependencyDashboard = new DependencyDashboard();
        Render();
    }

    public string FocusedElement { get; set; } = "";

    public void OpenDeps(Project project)
    {
        _currentProject = project;
        var dashboardPosition = new Position(0, Header.Height);
        DependencyDashboard.IsFocused = true;
        CommandLine.IsFocused = false;
        DependencyDashboard.Render(_currentProject, dashboardPosition);
    }

    private void ResizeDependencies()
    {
        var topPosition = 0;
        topPosition += _commandLineInDisplay ? CommandLine.Height : 0;
        topPosition += _headerInDisplay ? Header.Height : 0;
        var dashboardPosition = new Position(0, topPosition);
        DependencyDashboard.Render(_currentProject, dashboardPosition);
    }

    public void Render()
    {
        var headerPosition = new Position(0, 0);
        Header.Render(headerPosition);

        const string copyrightText = "Kolosov Aleksandr";
        var copyrightPosition = new Position(
                Console.WindowWidth - copyrightText.Width(),
                Console.WindowHeight);
        Copyright.Render(copyrightText, copyrightPosition);
    }

    public void ToggleHeader()
    {
        _headerInDisplay = !_headerInDisplay;
        if (_headerInDisplay)
        {
            var headerPosition = new Position(0, 0);
            Header.Render(headerPosition);
        }

        ResizeDependencies();
    }

    public void Next()
    {
        DependencyDashboard.Next();
    }

    public void Previous()
    {
        DependencyDashboard.Previous();
    }

    public void OpenCommandLine()
    {
        var commandLinePosition = new Position(0, _headerInDisplay ? Header.Height : 0);
        CommandLine.IsFocused = true;
        DependencyDashboard.IsFocused = false;
        FocusedElement = nameof(CommandLine);
        CommandLine.Render(commandLinePosition);
        _commandLineInDisplay = true;
        ResizeDependencies();
        Console.SetCursorPosition(commandLinePosition.Left + Theme.Padding + Theme.BorderWidth + 2,
                commandLinePosition.Top + Theme.BorderWidth);
        Console.CursorVisible = true;
    }
}