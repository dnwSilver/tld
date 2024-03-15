using System.Text.Json;
using TUI.Domain;
using TUI.Engine;
using TUI.Settings;

namespace TUI.Store;

public static class DependenciesStore
{
    private readonly static Dictionary<string, Package> Packages = new();

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
               $"private_token={token}&ref=dev";
    }
}