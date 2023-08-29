using System.Diagnostics;
using System.Net;
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

    private readonly static Dictionary<string, Package> Packages = new();
    private readonly Dashboard _dashboard = new();

    private readonly Table _table = new();

    public bool IsFocused
    {
        get => _dashboard.IsFocused;
        set => _dashboard.IsFocused = value;
    }

    public void Render(Project project, Position position, int? height = 0)
    {
        _dashboard.Render(project.Icon + " Dependencies", position);
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
                           .Select(dependency => GetVersion(dependency, package))
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
                    return " Repository or branch master not found.".Pastel(Palette.ErrorColor);
            }

            throw;
        }
        catch (Exception exception)
        {
            Debugger.Break();
            return "󰋔 We tried to send a request but couldn't. Check your configuration.".Pastel(Palette.ErrorColor);
        }
    }


    private static string GetVersion(DependencyDto dependency, Package package)
    {
        var currentVersion = package.ParseVersion(dependency.Name);

        if (currentVersion == null) return Icons.NotFound;

        var conventionVersion = dependency.Version?.ToVersion();
        return PaintingVersion(currentVersion, conventionVersion);
    }

    private static string PaintingVersion(Version current, Version? convention)
    {
        var textVersion = current.ToString();

        if (current > convention) return textVersion.Info();

        if (current < convention)
            return current.Major == convention.Major ? textVersion.Primary() : textVersion.Warning();

        return textVersion.Hint();
    }

    private static Package DownloadPackage(SourceDto sourceDto)
    {
        var endpoint = sourceDto.Tags.Have("gitlab") ? GetGitlabEndpoint(sourceDto) : sourceDto.Repo;
        if (Packages.TryGetValue(endpoint, out var downloadPackage)) return downloadPackage;

        using HttpClient client = new();
        var json = client.GetStringAsync(endpoint).GetAwaiter().GetResult();
        var package = JsonSerializer.Deserialize<Package>(json);
        Packages.Add(endpoint, package);
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
        if (versionWidth == 0) return ' '.Repeat(ColumnWidth - 1) + Icons.NotFound.Hint();

        return ' '.Repeat(ColumnWidth - versionWidth) + version;
    }

    private static string GetTitle(SourceDto sourceDto)
    {
        var title = "";

        title += RenderPadding();
        title += RenderTags(sourceDto);
        if (title.Width() + sourceDto.Name.Length + Theme.Padding <= TitleWidth)
        {
            title += sourceDto.Name;
        }
        else
        {
            var maxNameWidth = TitleWidth - title.Width() - Theme.Padding;
            title += $"{sourceDto.Name[..(maxNameWidth - 1)]}{"#".Hint()}";
        }

        title += RenderPadding();
        return $"{title}{' '.Repeat(TitleWidth - title.Width())}";
    }

    private static string RenderPadding()
    {
        return new string(' ', Theme.Padding);
    }

    private static string RenderTags(SourceDto sourceDto)
    {
        var tags = "";
        tags += GetGitApplication(sourceDto);
        tags += " ";
        tags += sourceDto.Tags.Have("public") ? Icons.NetworkPublic : Icons.NetworkPrivate;
        tags += " ";
        tags += sourceDto.Tags.Have("seo") ? Icons.SEO : Icons.SEO.Disable();
        tags += " ";
        tags += sourceDto.Tags.Have("auth") ? Icons.Auth : Icons.Auth.Disable();
        tags += " ";
        tags += GetApplicationType(sourceDto);
        tags += " ";

        return tags;
    }

    private static string GetApplicationType(SourceDto sourceDto)
    {
        foreach (var application in Icons.Applications)
            if (sourceDto.Tags.Have(application.Value))
                return application.Key;

        return Icons.Undefined;
    }

    private static string GetGitApplication(SourceDto sourceDto)
    {
        return sourceDto.Repo switch
        {
            { } url when url.Contains("gitlab") => Icons.GitLab,
            { } url when url.Contains("github") => Icons.GitHub,
            _                                   => Icons.Git
        };
    }

    public void Next()
    {
        _table.Next();
    }

    public void Previous()
    {
        _table.Previous();
    }
}