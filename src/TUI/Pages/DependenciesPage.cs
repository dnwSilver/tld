using TUI.Controls.Components;
using TUI.Controls.Containers;
using TUI.Controls.Layouts;
using TUI.Engine.Rendering.Canvas;
using TUI.Store;

namespace TUI.Pages;

public record DependenciesState(HeaderContainer Header, DashboardContainer Dashboard, FooterContainer Footer);

public class DependenciesPage : PageBase
{
    private DependenciesStore _store;
    
    private DependenciesState _state;
    
    public override void Initial()
    {
        var header = new HeaderContainer();
        var dashboard = new DashboardContainer();
        var dependenciesHeader = new DependenciesContainer();
        dependenciesHeader.AddTitleStub();
        
        foreach (var conventionDependency in _store.ConventionDependencies)
        {
            dependenciesHeader.AddDependency(conventionDependency);
        }
        
        dashboard.AddChildren(dependenciesHeader);
        
        foreach (var project in _store.Projects)
        {
            var projectDependencies = new DependenciesContainer(project);
            projectDependencies.AddTitle(new ProjectTitle(project));
            dashboard.AddChildren(projectDependencies);
        }
        
        var breadCrumbs = new BreadCrumbsComponent("  Dependencies", "JavaScript");
        var footer = new FooterContainer(breadCrumbs);
        
        _state = new DependenciesState(header, dashboard, footer);
    }
    
    public override void Render()
    {
        ICanvas canvas = new ConsoleCanvas();
        var layout = new DashboardLayout(_state.Header, _state.Dashboard, _state.Footer);
        canvas.Draw(layout);
    }
    
    public void LoadDependencies()
    {
        Initial();
        var projects = _state.Dashboard.GetContent();
        foreach (var projectDependencies in projects.Cast<DependenciesContainer?>().Skip(1))
        {
            if (projectDependencies is null)
            {
                continue;
            }
            
            var project = projectDependencies.Project;
            var actualDependencies = _store.ActualDependencies(project).ToArray();
            
            if (!actualDependencies.Any())
            {
                projectDependencies.AddError();
            }
            else
            {
                foreach (var conventionDependency in _store.ConventionDependencies)
                {
                    var actualDependency = actualDependencies.SingleOrDefault(
                        dependency => string.Equals(dependency.Brand.Name, conventionDependency.Brand.Name,
                            StringComparison.CurrentCultureIgnoreCase));
                    
                    if (actualDependency is null)
                    {
                        projectDependencies.AddDependencyStub();
                        continue;
                    }
                    
                    var versionType = actualDependency.Comparison(conventionDependency);
                    projectDependencies.AddDependency(actualDependency, versionType);
                }
            }
            
            Render();
        }
    }
    
    public override void Bind()
    {
        _store = new DependenciesStore();
        _store.Bind();
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