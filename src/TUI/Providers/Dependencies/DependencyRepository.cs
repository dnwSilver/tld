using System.Text.Json;
using System.Text.Json.Nodes;
using TUI.Domain;
using TUI.Logs;
using YamlDotNet.Serialization;
using YamlDotNet.Serialization.NamingConventions;

namespace TUI.Providers.Dependencies;

public class DependencyRepository
{
    private DependenciesDto? _dependenciesDto;

    private DependenciesDto DependenciesDto
    {
        get
        {
            if (_dependenciesDto is not null)
            {
                return _dependenciesDto;
            }

            var deserializer = new DeserializerBuilder()
                .WithNamingConvention(UnderscoredNamingConvention.Instance)
                .Build();

            using var sr = new StreamReader("dependencies.yaml");
            _dependenciesDto = deserializer.Deserialize<DependenciesDto>(sr.ReadToEnd());

            return _dependenciesDto;
        }
    }

    public IEnumerable<Dependency> ReadConventions(string stackName)
    {
        return DependenciesDto.Stacks
            .Single(stack => stack.Name == stackName)
            .Conventions
            .Select(convention =>
            {
                var brand = new Brand(convention.Name, convention.Icon, convention.Color);
                return new Dependency(convention.Version, brand);
            });
    }

    public IEnumerable<Project> ReadProjects(string stackName)
    {
        var projects = new List<Project>();

        var hubs = DependenciesDto.Stacks
            .Single(stack => stack.Name == stackName)
            .Hubs;

        foreach (var hub in hubs)
        {
            projects.AddRange(hub
                .Projects
                .Select(proj => new Project(proj.Id, proj.Name, proj.Tags, new Hub(hub.Origin, hub.Type))));
        }

        return projects;
    }


    public IEnumerable<Dependency> ReadActual(Project project)
    {
        var dependencies = new List<Dependency>();

        if (project.Hub.Type == "gitlab")
        {
            var endpoint = GetGitlabEndpoint(project.Hub.Origin, project.Id);
            using HttpClient client = new();
            var json = client.GetStringAsync(endpoint).GetAwaiter().GetResult();
            var packageJson = JsonSerializer.Deserialize<PackageJson>(json);
            Log.Debug($"Fetch dependencies for project {project.Name}.");
            dependencies.AddRange(Map(packageJson?.Dependencies));
            dependencies.AddRange(Map(packageJson?.DevDependencies));
            dependencies.AddRange(Map(packageJson?.Engines));
        }

        return dependencies;
    }

    private static string GetGitlabEndpoint(string origin, int projectId)
    {
        var token = Environment.GetEnvironmentVariable("TLD_GITLAB_PAT");
        return $"{origin}/api/v4/projects/{projectId}/repository/files/package.json/raw?" +
               $"private_token={token}&ref=dev";
    }

    private static IEnumerable<Dependency> Map(JsonObject? dependencies)
    {
        if (dependencies is null)
        {
            yield break;
        }

        foreach (var dependency in dependencies)
        {
            var actualVersion = dependency.Value?.ToString();
            var brand = new Brand(dependency.Key);

            if (actualVersion is null)
            {
                continue;
            }

            yield return new Dependency(actualVersion, brand);
        }
    }
}


// private static Package DownloadPackage(ProjectDto project)
// {
//     // var endpoint = projectDto.Tags.Have("gitlab") ? GetGitlabEndpoint(projectDto) : projectDto.Repo;
//     var endpoint = "";
//     if (Packages.TryGetValue(endpoint, out var downloadPackage)) return downloadPackage;
//
//     using HttpClient client = new();
//     var json = client.GetStringAsync(endpoint).GetAwaiter().GetResult();
//
//     Packages.Add(endpoint, package);
//     return package;
// }
//