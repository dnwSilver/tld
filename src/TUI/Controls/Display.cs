using TUI.Dashboards;
using TUI.Domain;
using TUI.UserInterface;


namespace TUI.Controls;

public class Display
{
    private bool _headerInDisplay = true;
    public readonly Header Header;

    public readonly Copyright Copyright;

    public readonly DependencyDashboard DependencyDashboard;
    private Project _currentProject;

    public Display()
    {
        Header = new Header();
        Copyright = new Copyright();
        DependencyDashboard = new DependencyDashboard();

        Render();
    }

    public void OpenDeps(Project project)
    {
        _currentProject = project;
        var dashboardPosition = new Position(0, Header.Height);
        DependencyDashboard.Render(_currentProject, dashboardPosition);
    }

    private void ResizeDependencies(bool full)
    {
        var dashboardPosition = new Position(0, full ? 0 : Header.Height);
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

        ResizeDependencies(!_headerInDisplay);
    }

    public void Next()
    {
        DependencyDashboard.Next();
    }

    public void Previous()
    {
        DependencyDashboard.Previous();
    }
}