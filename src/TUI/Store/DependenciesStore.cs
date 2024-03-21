using System.Text.Json;
using TUI.Providers.Dependencies;

namespace TUI.Store;

public static class DependenciesStore
{
    private readonly static Dictionary<string, Package> Packages = new();

    private static Package DownloadPackage(ProjectDto projectDto)
    {
        // var endpoint = projectDto.Tags.Have("gitlab") ? GetGitlabEndpoint(projectDto) : projectDto.Repo;
        var endpoint = "";
        if (Packages.TryGetValue(endpoint, out var downloadPackage)) return downloadPackage;

        using HttpClient client = new();
        var json = client.GetStringAsync(endpoint).GetAwaiter().GetResult();
        var package = JsonSerializer.Deserialize<Package>(json);
        Packages.Add(endpoint, package);
        return package;
    }

    private static string GetGitlabEndpoint(ProjectDto projectDto)
    {
        var token = Environment.GetEnvironmentVariable("TLD_GITLAB_PAT");
        // return $"{projectDto.Repo}/api/v4/projects/{projectDto.ProjectId}/repository/files/package.json/raw?" +
        // $"private_token={token}&ref=dev";
        return "";
    }
}