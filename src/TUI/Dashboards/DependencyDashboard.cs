using System.Drawing;
using System.Text;
using System.Text.Json;
using Pastel;
using TUI.Controls;
using TUI.Domain;
using TUI.UserInterface;


namespace TUI.Dashboards;

public class DependencyDashboard : IControl<Project>
{
    private int _selectedRowNumber = 0;

    private const int TitleWidth = 35;
    private const int ColumnWidth = 10;

    private Table _table = new();

    public void Render(Project project, Position position)
    {
        var dashboard = new Dashboard();
        dashboard.Render(project.Icon, position);

        var header = project.Dependencies.Select(GetConventionVersion).ToArray();
        var rows = project.Sources.Select(GetTitle).ToArray();

        var tablePosition = new Position(
                position.Left + Theme.BorderWidth,
                position.Top + Theme.BorderWidth);

        var tableProps = new TableProps(header, rows, TitleWidth, ColumnWidth);

        _table.Render(tableProps, tablePosition);

        for (var rowId = 0; rowId < rows.Length; rowId++)
        {
            var actualDependencies = GetDependencies(project.Sources[rowId], project.Dependencies);
            _table.RenderRow(rowId + 1, rows[rowId] + actualDependencies);
        }

        // Panel.RenderRows(project.Sources.ToArray(), _selectedRowNumber);
    }

    private string GetDependencies(Source source, IEnumerable<Dependency> conventionDependencies)
    {
        try
        {
            var package = DownloadPackage(source);

            return string.Join("",
                    conventionDependencies
                           .Select(package.Dependencies.GetVersion)
                           .Select(GetCurrentVersion));
        }
        catch
        {
            return "󰋔 We tried to send a request but couldn't. Check your configuration.".Pastel(Palette.ErrorColor);
        }
    }

    private readonly static Dictionary<string, Package> Packages = new();

    private static Package DownloadPackage(Source source)
    {
        if (Packages.TryGetValue(source.Repo, out var downloadPackage))
        {
            return downloadPackage;
        }

        using HttpClient client = new();
        var endpoint = source.Tags.Have("gitlab") ? GetGitlabEndpoint(source) : source.Repo;
        var json = client.GetStringAsync(endpoint).GetAwaiter().GetResult();
        var package = JsonSerializer.Deserialize<Package>(json);
        Packages.Add(source.Repo, package);
        return package;
    }

    private static string GetGitlabEndpoint(Source source)
    {
        var token = Environment.GetEnvironmentVariable("TLD_GITLAB_PAT");
        return $"{source.Repo}/api/v4/projects/{source.ProjectId}/repository/files/package.json/raw?" +
               $"private_token={token}&ref=master";
    }
    
    private static string GetConventionVersion(Dependency dependency)
    {
        return dependency.Icon.Pastel(dependency.Color) + dependency.Version.Primary();
    }

    private static string GetCurrentVersion(string version)
    {
        return ' '.Repeat(ColumnWidth - version.Width()) + version;
    }

    private static string GetTitle(Source source)
    {
        var rowText = new StringBuilder();

        RenderPadding(rowText);
        RenderTags(rowText, source);
        rowText.Append(source.Name);
        RenderPadding(rowText);
        var text = rowText.ToString();
        return $"{text}{' '.Repeat(TitleWidth - text.Width())}";
    }

    private static void RenderPadding(StringBuilder rowText)
    {
        rowText.Append(new string(' ', Theme.Padding));
    }

    private static void RenderTags(StringBuilder rowText, Source source)
    {
        rowText.Append(GetGitApplication(source));
        rowText.Append(source.Tags.Have("public")
                ? GetIcon("󰞉", "00FFFF")
                : GetIcon("󰕑", "AFE1AF"));
        rowText.Append(GetIcon("󰚩", "4285F4", source.Tags.Have("seo")));
        rowText.Append(GetIcon("",  "FFD700", source.Tags.Have("auth")));
        rowText.Append(GetApplicationType(source));
    }

    private static string GetApplicationType(Source source)
    {
        if (source.Tags.Have("site"))
            return GetIcon("", "BF40BF");
        if (source.Tags.Have("api"))
            return GetIcon("", "7F52FF");
        if (source.Tags.Have("package"))
            return GetIcon("", "CB0000");
        if (source.Tags.Have("image"))
            return GetIcon("󰡨", "086DD7");

        return GetIcon("", "CB0000");
    }

    private static string GetGitApplication(Source source) => source.Repo switch
    {
        { } url when url.Contains("gitlab") => GetIcon("", "E24329"),
        { } url when url.Contains("github") => GetIcon("", "ADBAC7"),
        _                                   => GetIcon("", "F14E32")
    };

    private static string GetIcon(string icon, string activeColor, bool enabled = true) =>
            (icon.Pastel(enabled ? activeColor : "71797E") + " ").PadLeft(2);

    public void Next()
    {
        _table.Next();
    }

    public void Previous()
    {
        _table.Previous();
    }
}