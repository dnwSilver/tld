using System.Diagnostics;
using System.Net;
using System.Text;
using System.Text.Json;
using Pastel;
using TUI.Controls;
using TUI.Domain;
using TUI.Settings;
using TUI.UserInterface;


namespace TUI.Dashboards;

public class DependencyDashboard : IControl<Project>
{
    private const int TitleWidth = 25;
    private const int ColumnWidth = 10;

    private readonly Table _table = new();

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
    }

    private static string GetDependencies(SourceDto sourceDto, IEnumerable<DependencyDto> conventionDependencies)
    {
        try
        {
            var package = DownloadPackage(sourceDto);

            return string.Join("",
                    conventionDependencies
                           .Select(package.Dependencies.GetVersion)
                           .Select(RenderCurrentVersion));
        }
        catch (HttpRequestException exception)
        {
            switch (exception.StatusCode)
            {
                case HttpStatusCode.BadRequest:
                    return " Request have errors.".Pastel(Palette.ErrorColor);
                case HttpStatusCode.Forbidden:
                    return " Not enough rights.".Pastel(Palette.ErrorColor);
                case HttpStatusCode.NotFound:
                    return " Repository not found.".Pastel(Palette.ErrorColor);
            }

            throw;
        }
        catch (Exception exception)
        {
            Debugger.Break();
            return "󰋔 We tried to send a request but couldn't. Check your configuration.".Pastel(Palette.ErrorColor);
        }
    }

    private readonly static Dictionary<string, Package> Packages = new();

    private static Package DownloadPackage(SourceDto sourceDto)
    {
        if (Packages.TryGetValue(sourceDto.Repo, out var downloadPackage))
        {
            return downloadPackage;
        }

        using HttpClient client = new();
        var endpoint = sourceDto.Tags.Have("gitlab") ? GetGitlabEndpoint(sourceDto) : sourceDto.Repo;
        var json = client.GetStringAsync(endpoint).GetAwaiter().GetResult();
        var package = JsonSerializer.Deserialize<Package>(json);
        Packages.Add(sourceDto.Repo, package);
        return package;
    }

    private static string GetGitlabEndpoint(SourceDto sourceDto)
    {
        var token = Environment.GetEnvironmentVariable("TLD_GITLAB_PAT");
        return $"{sourceDto.Repo}/api/v4/projects/{sourceDto.ProjectId}/repository/files/package.json/raw?" +
               $"private_token={token}&ref=master";
    }

    private static string GetConventionVersion(DependencyDto dependencyDto)
    {
        return dependencyDto.Icon.Pastel(dependencyDto.Color) + dependencyDto.Version.Primary();
    }

    private static string RenderCurrentVersion(string version)
    {
        var versionWidth = version.Width();
        if (versionWidth == 0)
        {
            return ' '.Repeat(ColumnWidth - 1) + "".Hint();
        }

        return ' '.Repeat(ColumnWidth - versionWidth) + version;
    }

    private static string GetTitle(SourceDto sourceDto)
    {
        var rowText = new StringBuilder();

        RenderPadding(rowText);
        RenderTags(rowText, sourceDto);
        rowText.Append(sourceDto.Name);
        RenderPadding(rowText);
        var text = rowText.ToString();
        return $"{text}{' '.Repeat(TitleWidth - text.Width())}";
    }

    private static void RenderPadding(StringBuilder rowText)
    {
        rowText.Append(new string(' ', Theme.Padding));
    }

    private static void RenderTags(StringBuilder rowText, SourceDto sourceDto)
    {
        rowText.Append(GetGitApplication(sourceDto));
        rowText.Append(sourceDto.Tags.Have("public")
                ? GetIcon("󰞉", "00FFFF")
                : GetIcon("󰕑", "AFE1AF"));
        rowText.Append(GetIcon("󰚩", "4285F4", sourceDto.Tags.Have("seo")));
        rowText.Append(GetIcon("",  "FFD700", sourceDto.Tags.Have("auth")));
        rowText.Append(GetApplicationType(sourceDto));
    }

    private static string GetApplicationType(SourceDto sourceDto)
    {
        if (sourceDto.Tags.Have("site"))
            return GetIcon("", "BF40BF");
        if (sourceDto.Tags.Have("api"))
            return GetIcon("", "7F52FF");
        if (sourceDto.Tags.Have("package"))
            return GetIcon("", "CB0000");
        if (sourceDto.Tags.Have("image"))
            return GetIcon("󰡨", "086DD7");

        return GetIcon("", "CB0000");
    }

    private static string GetGitApplication(SourceDto sourceDto) => sourceDto.Repo switch
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