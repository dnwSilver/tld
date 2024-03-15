using TUI.Components.Controls;
using TUI.Domain;
using TUI.Engine;
using TUI.Engine.Attributes;
using TUI.Engine.Attributes.Alignments;
using TUI.Engine.Components;

namespace TUI.Components.Views;

public class DependenciesView : ComponentAttribute, IComponent
{
    private const string ViewName = "Dependencies";

    private DevelopmentStack _developmentStack;

    public void Bind(DevelopmentStack developmentStack)
    {
        _developmentStack = developmentStack;
    }

    public void Render(Horizontal horizontal, Size size)
    {
        var dashboardTitle = _developmentStack.Icon + Symbols.Space + ViewName;
        var dashboard = new Dashboard(dashboardTitle);

        // Add(dashboard);
    }

    // private const int TitleWidth = 25;
    // private const int ColumnWidth = 10;

    // private readonly DashboardControl _dashboard = new();

    // public bool IsFocused
    // {
    // get => _dashboard.IsFocused;
    // set => _dashboard.IsFocused = value;
    // }

    // public void Render(ProjectDto projectDto, ControlPosition position)
    // {
    //     _dashboard.Render(projectDto.Icon + " Dependencies", position);
    //     var header = projectDto.Dependencies.Select(GetConventionVersion).ToArray();
    //     var rows = projectDto.Sources.Select(GetTitle).ToArray();
    //
    //     var tablePosition = new ControlPosition(
    //         position.Left + Theme.BorderWidth,
    //         position.Top + Theme.BorderWidth);
    //
    //     var tableProps = new TableProps(header, rows, TitleWidth, ColumnWidth);
    //     _table.Render(tableProps, tablePosition);
    //
    //     for (var rowId = 0; rowId < rows.Length; rowId++)
    //     {
    //         var actualDependencies = GetDependencies(projectDto.Sources[rowId], projectDto.Dependencies);
    //         _table.RenderRow(rowId + 1, rows[rowId] + actualDependencies);
    //     }
    // }

    // private static string GetDependencies(SourceDto sourceDto, IEnumerable<DependencyDto> conventionDependencies)
    // {
    //     try
    //     {
    //         var package = DownloadPackage(sourceDto);
    //
    //         return string.Join("",
    //             conventionDependencies
    //                 .Select(dependency => GetVersion(dependency, package))
    //                 .Select(RenderCurrentVersion));
    //     }
    //     catch (HttpRequestException exception)
    //     {
    //         switch (exception.StatusCode)
    //         {
    //             case HttpStatusCode.BadRequest:
    //                 return " Request have errors.".Pastel(Palette.ErrorColor);
    //             case HttpStatusCode.Forbidden:
    //                 return " Not enough rights.".Pastel(Palette.ErrorColor);
    //             case HttpStatusCode.NotFound:
    //                 return " Repository or branch master not found.".Pastel(Palette.ErrorColor);
    //         }
    //
    //         throw;
    //     }
    //     catch (Exception exception)
    //     {
    //         Debugger.Break();
    //         return "󰋔 We tried to send a request but couldn't. Check your configuration.".Pastel(Palette.ErrorColor);
    //     }
    // }
    //
    // private static string GetVersion(DependencyDto dependency, Package package)
    // {
    //     var currentVersion = package.ParseVersion(dependency.Name);
    //
    //     if (currentVersion == null) return Icons.NotFound;
    //
    //     var conventionVersion = dependency.Version?.ToVersion();
    //     return PaintingVersion(currentVersion, conventionVersion);
    // }
    //
    // private static string PaintingVersion(Version current, Version? convention)
    // {
    //     var textVersion = current.ToString();
    //
    //     if (current > convention) return textVersion.Info();
    //
    //     if (current < convention)
    //         return current.Major == convention.Major ? textVersion.Primary() : textVersion.Warning();
    //
    //     return textVersion.Hint();
    // }
    //
    // private static string GetConventionVersion(DependencyDto dependencyDto)
    // {
    //     return dependencyDto.Icon.Pastel(dependencyDto.Color) + dependencyDto.Version.Primary();
    // }
    //
    // private static string RenderCurrentVersion(string version)
    // {
    //     var versionWidth = version.Width();
    //     if (versionWidth == 0) return ' '.Repeat(ColumnWidth - 1) + Icons.NotFound.Hint();
    //
    //     return ' '.Repeat(ColumnWidth - versionWidth) + version;
    // }
    //
    // private static string GetTitle(SourceDto sourceDto)
    // {
    //     var title = "";
    //
    //     title += RenderPadding();
    //     title += RenderTags(sourceDto);
    //     if (title.Width() + sourceDto.Name.Length + Theme.Padding <= TitleWidth)
    //     {
    //         title += sourceDto.Name;
    //     }
    //     else
    //     {
    //         var maxNameWidth = TitleWidth - title.Width() - Theme.Padding;
    //         title += $"{sourceDto.Name[..(maxNameWidth - 1)]}{"#".Hint()}";
    //     }
    //
    //     title += RenderPadding();
    //     return $"{title}{' '.Repeat(TitleWidth - title.Width())}";
    // }
    //
    // public void Next()
    // {
    //     _table.Next();
    // }
    //
    // public void Previous()
    // {
    //     _table.Previous();
    // }
    protected override Sketch DrawComponent()
    {
        throw new NotImplementedException();
    }
}